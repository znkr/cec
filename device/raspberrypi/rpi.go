// Copyright 2017 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This package provides a way to use the Raspberry PI hardware for CEC.
package raspberrypi

//go:generate stringer -type=notify

import (
	"log"
	"unsafe"

	"github.com/krynr/cec"
)

/*
#cgo CFLAGS: -I/opt/vc/include -I/opt/vc/include/interface/vcos/pthreads -I/opt/vc/include/interface/vmcs_host/linux/
#cgo LDFLAGS: -L/opt/vc/lib -lbcm_host

#include <bcm_host.h>

extern void rpi_cec_callback(void*, uint32_t, uint32_t, uint32_t, uint32_t, uint32_t);

// Gateway function to register the callback.
static inline void register_rpi_cec_callback() {
  vc_cec_register_callback(rpi_cec_callback, NULL);
}
*/
import "C"

type notify uint32

const (
	notifyTx              notify = 1 << 0
	notifyRx              notify = 1 << 1
	notifyButtonPressed   notify = 1 << 2
	notifyButtonRelease   notify = 1 << 3
	notifyRemotePressed   notify = 1 << 4
	notifyRemoteRelease   notify = 1 << 5
	notifyLogicalAddr     notify = 1 << 6
	notifyTopology        notify = 1 << 7
	notifyLogicalAddrLost notify = 1 << 15
)

const vendorID = 0x18C086 // Broadcom

type incoming struct {
	n   notify
	rc  uint32
	msg cec.Message
}

type outgoing struct {
	addr  cec.LogicalAddr
	data  []byte
	reply bool
}

type device struct {
	out        chan outgoing
	in         chan cec.Packet
	deviceType cec.DeviceType
}

var packets chan<- cec.Packet

func handleOutgoing(out <-chan outgoing) {
	for {
		p := <-out

		var replyC C.vcos_bool_t
		if p.reply {
			replyC = C.VC_TRUE
		} else {
			replyC = C.VC_FALSE
		}
		ptr := unsafe.Pointer(&p.data[0])
		C.vc_cec_send_message(
			C.uint32_t(p.addr),
			(*C.uint8_t)(ptr),
			C.uint32_t(len(p.data)),
			replyC)
	}
}

//export rpi_cec_callback
// For a description of the arguments of this callback see
// https://github.com/raspberrypi/firmware/blob/master/opt/vc/include/interface/vmcs_host/vc_cec.h
func rpi_cec_callback(p unsafe.Pointer, hdr, p1, p2, p3, p4 uint32) {
	n := notify((hdr >> 0) & 0xffff)

	switch n {
	case notifyRx, notifyButtonPressed, notifyButtonRelease:
		l := (hdr >> 16) & 0xff
		if l < 1 {
			log.Panic("Message too small")
		}
		var data [32]byte
		for i, p := range []uint32{p1, p2, p3, p4} {
			data[4*i+0] = byte((p >> 0) & 0xff)
			data[4*i+1] = byte((p >> 8) & 0xff)
			data[4*i+2] = byte((p >> 16) & 0xff)
			data[4*i+3] = byte((p >> 24) & 0xff)
		}

		initiator := cec.LogicalAddr((data[0] >> 4) & 0xf)
		follower := cec.LogicalAddr((data[0] >> 0) & 0xf)
		op := cec.OpCode(data[1])
		var payload []byte
		if l > 2 {
			payload = data[2:l]
		}

		packets <- cec.Packet{
			Initiator: initiator,
			Follower:  follower,
			Op:        op,
			Data:      payload,
		}
	}
}

// Initializes the Raspberry CEC device.
//
// This method may only be called once.
func Init(a cec.LogicalAddr, t cec.DeviceType) cec.Device {
	if packets != nil {
		log.Fatal("Device already in use")
	}
	d := &device{
		out:        make(chan outgoing),
		in:         make(chan cec.Packet),
		deviceType: t,
	}
	packets = d.in
	C.bcm_host_init()
	C.vc_cec_set_passive(C.VC_TRUE)
	C.register_rpi_cec_callback()
	C.vc_cec_register_all()

	// If the logical address is not the one we want, try to set it.
	if d.GetLogicalAddress() != a {
		d.releaseLogicalAddress()

		if d.pollLogicalAddress(a) {
			log.Fatalf("Logical address %s already in use", a)
		}

		d.setLogicalAddress(a, t)
		if addr := d.GetLogicalAddress(); addr != a {
			log.Fatalf("Incorrect logical address: %s", addr)
		}
	}

	go handleOutgoing(d.out)

	log.Printf("Physical address: %s", d.GetPhysicalAddress())
	log.Printf("Logical address: %s", a)
	return d
}

func (d *device) Receive() <-chan cec.Packet {
	return d.in
}

func (d *device) Send(follower cec.LogicalAddr, op cec.OpCode, payload []byte) {
	d.out <- outgoing{
		addr:  follower,
		data:  append([]byte{byte(op)}, payload...),
		reply: false,
	}
}

func (d *device) Reply(follower cec.LogicalAddr, op cec.OpCode, payload []byte) {
	d.out <- outgoing{
		addr:  follower,
		data:  append([]byte{byte(op)}, payload...),
		reply: true,
	}
}

func (d *device) GetVendorID() uint32 {
	return vendorID
}

func (d *device) GetDeviceType() cec.DeviceType {
	return d.deviceType
}

func (d *device) GetPhysicalAddress() cec.PhysicalAddress {
	var address C.uint16_t
	if C.vc_cec_get_physical_address(&address) != 0 {
		log.Fatal("Failed to get physical address.")
	}
	return cec.PhysicalAddress(address)
}

func (d *device) GetLogicalAddress() cec.LogicalAddr {
	var address C.CEC_AllDevices_T
	if C.vc_cec_get_logical_address(&address) != 0 {
		log.Fatal("Failed to get logical address.")
	}
	return cec.LogicalAddr(address)
}

func (d *device) setLogicalAddress(a cec.LogicalAddr, t cec.DeviceType) {
	aC := C.CEC_AllDevices_T(a)
	tC := C.CEC_DEVICE_TYPE_T(t)
	vendorIdC := C.uint32_t(vendorID)
	if C.vc_cec_set_logical_address(aC, tC, vendorIdC) != 0 {
		log.Fatal("Failed to set logical address.")
	}
}

func (d *device) releaseLogicalAddress() {
	if C.vc_cec_release_logical_address() != 0 {
		log.Fatal("Failed to release logical address.")
	}
}

func (d *device) pollLogicalAddress(a cec.LogicalAddr) bool {
	aC := C.CEC_AllDevices_T(a)
	r := C.vc_cec_poll_address(aC)
	if r < 0 {
		log.Fatal("Failed to poll logical address.")
	} else if r == 0 {
		return true
	} else {
		return false
	}
	panic("never reached")
}
