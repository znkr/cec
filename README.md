cec - a Go library to work with HDMI CEC
========================================

## Intro

`cec` provides a low-level HDMI CEC interface implementing basic protocol rules, parsing, and error
checking. At the moment only a subset of all possible HDMI CEC op codes are supported, but adding
support for additional messages is straight forward.

The CEC logic itself is implemented on top of a device abstraction that should make it possible to
support all devices that allow access to the raw CEC messages. However, at the moment only an
implementation for the Raspberry Pi exists, as well as a fake for testing.

## Getting Started with a Raspberry Pi

The program below registers itself as an audio device on the HDMI CEC bus and logs all CEC messages
using the go `log` package.

```go
package main

import (
    "log"
    "os"

    "znkr.io/cec"
    "znkr.io/cec/device/raspberrypi"
)

func main() {
    d := raspberrypi.Init(cec.AudioSystem, cec.DeviceTypeAudio)
    x, err := cec.New(d, cec.Config{OSDName: "RPI"})
    if err != nil {
        log.Fatal("Unable to initalize CEC: %s", err)
    }

    // Handlers are processed in the order they are added. The first one will receive all
    // messages (after some validation). The following ones will only receive unhandled
    // messages, i.e. messages that for which all preceeding handlers returned false.
    x.AddHandlerFunc(func(x *cec.Cec, msg cec.Message) bool {
        log.Print(msg)
        return false
    })

    // This handler will handle ReportPowerStatus messages. Additionally, it supports the
    // GiveDevicePower status message when asked from other devices on the bus.
    x.AddHandler(func(x *cec.Cec, msg cec.Message) bool {
        switch cmd := msg.Cmd.(type) {
        case cec.ReportPowerStatus:
            if msg.Initiator == cec.TV {
                log.Print("TV power status is %s", cmd.Power)
                return true
            }
            return true
        case cec.GiveDevicePowerStatus:
            x.Reply(msg.Initiator, cec.ReportPowerStatus{
                Power: cec.PowerStatusOn
            })
            return true
        }
        return false
    })

    // The default handler is necessary to react to a few standard messages that must be
    // handled according to the standard. Without this, these messages would trigger an
    // abort response.
    x.AddHandler(cec.DefaultHandler{})

    // Starts listening on the CEC bus and handling messages.
    go x.Run()

    // Ask the TV for the power status. This should trigger a cec.ReportPowerStatus message which
    // is handled above.
    x.Send(cec.TV, cec.GiveDevicePowerStatus{})

    // Wait for Ctrl-C.
    signals := make(chan os.Signal, 1)
    signal.Notify(signals, os.Interrupt)
    <-signals
}
```

Since the `raspberrypi` package depends on Raspberry Pi C interfaces, it is easiest to build the
binary directly on the device instead of using a cross compiler. The package also needs
`-tags raspberrypi` on the command line in order to compile the Raspberry Pi device.

## Disclaimer

This is not an official Google product.
