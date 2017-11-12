// Package cec provides support for interacting with HDMI CEC devices.
package cec

import (
	"log"
)

const cecVersion = 0x04 // CEC 1.3a

// Configuration for this CEC endpoint.
type Config struct {
	OSDName string // Name to display in OSD menus, must be between 1 and 14 ASCII characters.
}

// Main type to communicate with the CEC bus.
type Cec struct {
	dev      Device
	osd      string
	handlers []Handler
	started  bool
}

// A Handler for HDMI CEC messages.
type Handler interface {
	// Handles a message and returns true if the message was handled.
	HandleMessage(x *Cec, msg Message) bool
}

// Function wrapper for Handler.
type HandlerFunc func(x *Cec, msg Message) bool

// HandlerFunc implements Handler.
func (f HandlerFunc) HandleMessage(x *Cec, msg Message) bool {
	return f(x, msg)
}

// The UnhandledHandler is used whenever an incoming CEC message isn't handled by any Handler.
type UnhandledHandler struct{}

// UnhandledHandler implements Handler.
func (f UnhandledHandler) HandleMessage(x *Cec, msg Message) bool {
	if msg.Initiator == Unregistered {
		// Ignore messages ending up here that are send from Unregistered.
		log.Printf("Unexpected message from unregistered initiator: %s", msg)
		return true
	}

	switch msg.Cmd.(type) {
	case FeatureAbort:
		log.Printf("Unexpected feature abort: %s", msg)
		return true

	case Standby:
		// Standby is mandatory, but may be ignored. Returning here in order to not send a feature abort.
		return true
	}

	// Send FeatureAbort if this message was directly addressed to us. Unhandled broadcasts are ignored.
	if msg.Follower != Broadcast {
		log.Printf("Unexpected message: %s", msg)
		x.Reply(msg.Initiator, FeatureAbort{
			Abort:  msg.Cmd.Op(),
			Reason: AbortUnrecognizedOpCode,
		})
	}

	return true
}

// The DefaultHandler handles a set of standard messages. The handles messages are for physical address, vendor id, and
// CEC version.
type DefaultHandler struct{}

// DefaultHandler implements Handler.
func (h DefaultHandler) HandleMessage(x *Cec, msg Message) bool {
	switch msg.Cmd.(type) {
	case GivePhysicalAddress:
		x.Reply(Broadcast, ReportPhysicalAddress{
			Addr: x.dev.GetPhysicalAddress(),
			Type: x.dev.GetDeviceType(),
		})
		return true

	case GiveOSDName:
		x.Reply(msg.Initiator, SetOSDName{
			Name: x.osd,
		})
		return true

	case GiveDeviceVendorID:
		x.Reply(Broadcast, DeviceVendorID{
			VendorID: x.dev.GetVendorID(),
		})
		return true

	case GetCECVersion:
		x.Reply(msg.Initiator, CECVersion{
			Version: cecVersion,
		})
		return true
	}
	return false
}

// Creates a new Cec object using dev to communicate with the hardware.
func New(dev Device, c Config) (*Cec, error) {
	if !isValidOsdName(c.OSDName) {
		return nil, InvalidOSDName{}
	}
	return &Cec{
		dev:      dev,
		osd:      c.OSDName,
		handlers: []Handler{},
	}, nil
}

// Adds a handler. If more than one handler is added all handlers are tried until one handler returns true.
// May only be called before Start() was called.
func (x *Cec) AddHandler(h Handler) {
	if x.started {
		log.Panic("Already started.")
	}
	x.handlers = append(x.handlers, h)
}

// Adds a handler in the form of a function. This is a convenience wrapper around AddHandler. May only be called before
// Start() was called.
func (x *Cec) AddHandleFunc(f func(x *Cec, msg Message) bool) {
	x.AddHandler(HandlerFunc(f))
}

// Starts receiving and handling CEC messages.
func (x *Cec) Run() {
	unhandledHandler := UnhandledHandler{}
	for p := range x.dev.Receive() {
		msg, err := UnmarshalMessage(p)
		if err != nil {
			log.Printf("Unable to unmarshal message %s: %s", p, err)
			if p.Follower != Broadcast && p.Initiator != Unregistered {
				// Signal the initiator that acting upon the received packet is not possible.
				x.Reply(p.Initiator, FeatureAbort{
					Abort:  p.Op,
					Reason: AbortInvalidOperand,
				})
			}
			continue
		}

		// Some messages need to be ignored according to the spec.
		flags, ok := getOpCodeFlags(msg.Cmd.Op())
		if !ok {
			// We don't know anything about this opcode.
			log.Printf("Received message with unkown opcode: %s", msg)
			if p.Follower != Broadcast && p.Initiator != Unregistered {
				x.Reply(p.Initiator, FeatureAbort{
					Abort:  p.Op,
					Reason: AbortUnrecognizedOpCode,
				})
			}
			continue
		} else if msg.Follower == Broadcast && (flags&fBroadcast) == 0 {
			// Message is not valid in broadcast mode, but was broadcast.
			log.Printf("Received bradcast message which should be direct: %s", msg)
			continue
		} else if msg.Follower != Broadcast && (flags&fDirect) == 0 {
			// Message is not valid in direct mode, but directly addressed.
			log.Printf("Received direct message which should be a broadcast: %s", msg)
			continue
		} else if msg.Initiator == Unregistered && msg.Cmd.Op() != OpStandby &&
			(flags&(fBroadcastResponse|fSwitchMessage) == 0) {
			// Initiator is unregistered, ignore all messages except standby, switch messages and messages
			// answered by a broadcast response.
			continue
		}

		// Dispatch incoming message to handlers.
		handled := false
		for _, h := range x.handlers {
			if h.HandleMessage(x, msg) {
				handled = true
				break
			}
		}
		if !handled {
			unhandledHandler.HandleMessage(x, msg)
		}
	}
}

// Sends cmd to follower.
func (x *Cec) Send(follower LogicalAddr, cmd Command) error {
	data, err := cmd.Marshal()
	if err != nil {
		return err
	}
	x.dev.Send(follower, cmd.Op(), data)
	return nil
}

// Sends cmd to follower as a reply.
func (x *Cec) Reply(follower LogicalAddr, cmd Command) error {
	data, err := cmd.Marshal()
	if err != nil {
		return err
	}
	x.dev.Reply(follower, cmd.Op(), data)
	return nil
}
