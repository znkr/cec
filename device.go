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

package cec

import "fmt"

// A Packet contains the raw HDMI CEC data.
type Packet struct {
	Initiator LogicalAddr // The sender of this packet.
	Follower  LogicalAddr // The receiver of this packet.
	Op        OpCode      // The opcode.
	Data      []byte      // The payload.
}

// Packet implements the Stringer interface.
func (p Packet) String() string {
	return fmt.Sprintf("%s -> %s: %s %#v", p.Initiator, p.Follower, p.Op, p.Data)
}

// Device is a low level representation of a HDMI CEC device. It is used to communicate directly with hardware.
type Device interface {
	// Returns a channel of all received packets.
	Receive() <-chan Packet

	// Sends a CEC packet to follower.
	Send(follower LogicalAddr, op OpCode, payload []byte)

	// Sens a CEC packet to a follower as a reply.
	Reply(follower LogicalAddr, op OpCode, payload []byte)

	// Returns the vendor ID of this device.
	GetVendorID() uint32

	// Returns the device type of this device.
	GetDeviceType() DeviceType

	// Returns the physical address of this device.
	GetPhysicalAddress() PhysicalAddress

	// Returns the logical address of this device.
	GetLogicalAddress() LogicalAddr
}
