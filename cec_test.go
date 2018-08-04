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

package cec_test

import (
	"reflect"
	"testing"

	"github.com/krynr/cec/device/fake"

	. "github.com/krynr/cec"
)

func TestCec(t *testing.T) {
	tests := []struct {
		name  string
		setup func(c *Cec)
		in    []Packet
		out   []Packet
	}{
		// Tests with DefaultHandler only
		{
			name:  "give_physical_address",
			setup: func(c *Cec) { c.AddHandler(&DefaultHandler{}) },
			in: []Packet{
				{TV, AudioSystem, OpGivePhysicalAddress, nil},
			},
			out: []Packet{
				{AudioSystem, Broadcast, OpReportPhysicalAddress, append(fake.PhysicalAddress.Bytes(), byte(DeviceTypeAudio))},
			},
		}, {
			// This message is allowed to be send from unregistered because it produces a broadcast response
			name:  "give_physical_address_from_unregistered",
			setup: func(c *Cec) { c.AddHandler(&DefaultHandler{}) },
			in: []Packet{
				{Unregistered, AudioSystem, OpGivePhysicalAddress, nil},
			},
			out: []Packet{
				{AudioSystem, Broadcast, OpReportPhysicalAddress, append(fake.PhysicalAddress.Bytes(), byte(DeviceTypeAudio))},
			},
		}, {
			// This message needs to be ignored because it may only appear as a direct message
			name:  "give_physical_address_as_broadcast",
			setup: func(c *Cec) { c.AddHandler(&DefaultHandler{}) },
			in: []Packet{
				{TV, Broadcast, OpGivePhysicalAddress, nil},
			},
			out: []Packet{},
		}, {
			name:  "give_osd_name",
			setup: func(c *Cec) { c.AddHandler(&DefaultHandler{}) },
			in: []Packet{
				{TV, AudioSystem, OpGiveOSDName, nil},
			},
			out: []Packet{
				{AudioSystem, TV, OpSetOSDName, []byte("test")},
			},
		}, {
			name:  "give_device_vendor_id",
			setup: func(c *Cec) { c.AddHandler(&DefaultHandler{}) },
			in: []Packet{
				{TV, AudioSystem, OpGiveDeviceVendorID, nil},
			},
			out: []Packet{
				{AudioSystem, Broadcast, OpDeviceVendorID, []byte{0x10, 0x10, 0x10}},
			},
		}, {
			name:  "get_cec_version",
			setup: func(c *Cec) { c.AddHandler(&DefaultHandler{}) },
			in: []Packet{
				{TV, AudioSystem, OpGetCECVersion, nil},
			},
			out: []Packet{
				{AudioSystem, TV, OpCECVersion, []byte{0x04}},
			},
		},

		// Tests overriding an response from the default handler
		{
			name: "override_default_handler",
			setup: func(c *Cec) {
				c.AddHandleFunc(func(x *Cec, msg Message) bool {
					switch msg.Cmd.(type) {
					case GetCECVersion:
						x.Reply(msg.Initiator, CECVersion{Version: 0x05})
						return true
					}
					return false
				})
				c.AddHandler(&DefaultHandler{})
			},
			in: []Packet{
				{TV, AudioSystem, OpGiveDeviceVendorID, nil},
				{TV, AudioSystem, OpGetCECVersion, nil},
				{TV, AudioSystem, OpGiveAudioStatus, nil},
			},
			out: []Packet{
				{AudioSystem, Broadcast, OpDeviceVendorID, []byte{0x10, 0x10, 0x10}},
				{AudioSystem, TV, OpCECVersion, []byte{0x05}},
				{AudioSystem, TV, OpFeatureAbort, []byte{byte(OpGiveAudioStatus), byte(AbortUnrecognizedOpCode)}},
			},
		},

		// Tests sending things early
		{
			name: "send_early",
			setup: func(c *Cec) {
				c.Send(TV, Standby{})
			},
			in: []Packet{},
			out: []Packet{
				{AudioSystem, TV, OpStandby, []byte{}},
			},
		},

		// Tests without any handlers
		{
			// Unhandled broadcasts are ignored
			name: "unhandled_broadcast",
			in: []Packet{
				{TV, Broadcast, OpActiveSource, PhysicalAddress(0x1010).Bytes()},
			},
			out: []Packet{},
		}, {
			// A direct-only message send as broadcast must be ignored
			name: "direct_only_message_as_broadcast",
			in: []Packet{
				{TV, AudioSystem, OpActiveSource, PhysicalAddress(0x1010).Bytes()},
			},
			out: []Packet{},
		}, {
			// Must be ignored because it's from unregistered and non of the exceptions match
			name: "message_from_unregistered",
			in: []Packet{
				{Unregistered, Broadcast, OpActiveSource, PhysicalAddress(0x1010).Bytes()},
			},
			out: []Packet{},
		}, {
			// No reply for standby
			name: "standby",
			in: []Packet{
				{TV, AudioSystem, OpStandby, nil},
			},
			out: []Packet{},
		}, {
			name: "standby_as_broadcast",
			in: []Packet{
				{TV, Broadcast, OpStandby, nil},
			},
			out: []Packet{},
		}, {
			// Standby from unregistered is allowed but doesn't produce a response
			name: "standby_from_unregistered",
			in: []Packet{
				{Unregistered, AudioSystem, OpStandby, nil},
			},
			out: []Packet{},
		}, {
			// Feature aborts don't get a response
			name: "unexpected_feature_abort",
			in: []Packet{
				{TV, AudioSystem, OpFeatureAbort, []byte{byte(OpSetOSDName), byte(AbortInvalidOperand)}},
			},
			out: []Packet{},
		}, {
			// Unhandled direct messages are answered by a feature abort
			name: "unhandled_message",
			in: []Packet{
				{TV, AudioSystem, OpGiveDeviceVendorID, nil},
			},
			out: []Packet{
				{AudioSystem, TV, OpFeatureAbort, []byte{byte(OpGiveDeviceVendorID), byte(AbortUnrecognizedOpCode)}},
			},
		}, {
			// Invalid messages are answered by a feature abort
			name: "invalid_message",
			in: []Packet{
				{TV, AudioSystem, OpSetOSDName, nil},
			},
			out: []Packet{
				{AudioSystem, TV, OpFeatureAbort, []byte{byte(OpSetOSDName), byte(AbortInvalidOperand)}},
			},
		}, {
			// Unknown opcodes are answered by a feature abort
			name: "unkown_opcode",
			in: []Packet{
				{TV, AudioSystem, OpCode(254), nil},
			},
			out: []Packet{
				{AudioSystem, TV, OpFeatureAbort, []byte{byte(OpCode(254)), byte(AbortUnrecognizedOpCode)}},
			},
		}, {
			// Unknown opcodes as broadcasts are ignored
			name: "unkown_opcode_broadcast",
			in: []Packet{
				{TV, Broadcast, OpCode(254), nil},
			},
			out: []Packet{},
		}, {
			// Unknown opcodes from unregistered are ignored
			name: "unkown_opcode_broadcast",
			in: []Packet{
				{Unregistered, AudioSystem, OpCode(254), nil},
			},
			out: []Packet{},
		}, {
			// Invalid message broadcasts are ignored
			name: "invalid_message_as_broadcast",
			in: []Packet{
				{TV, Broadcast, OpSetOSDName, nil},
			},
			out: []Packet{},
		}, {
			// Invalid messages from unregistered are ignored
			name:  "invalid_message_from_unregistered",
			setup: func(c *Cec) {},
			in: []Packet{
				{Unregistered, AudioSystem, OpSetOSDName, nil},
			},
			out: []Packet{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := fake.New(AudioSystem, DeviceTypeAudio)
			c, err := New(d, Config{
				OSDName: "test",
			})
			if err != nil {
				t.Errorf("Error setting up %s", err)
				return
			}

			if test.setup != nil {
				test.setup(c)
			}
			out := d.Run(test.in, func() { c.Run() })
			expected := test.out
			if len(out) != len(expected) {
				t.Errorf("Expected %d outputs, but received %d", len(expected), len(out))
				return
			}
			for i := range expected {
				if !reflect.DeepEqual(out[i], expected[i]) {
					t.Errorf("Expected %d-th output to be %s but got %s.", i, expected[i], out[i])
				}
			}
		})
	}
}

func TestInvalidOsdName(t *testing.T) {
	_, err := New(fake.New(AudioSystem, DeviceTypeAudio), Config{
		OSDName: "This OSD name is too long",
	})
	if err == nil {
		t.Errorf("Expected failure due to invalid OSD name, but succeeded.")
	}
}
