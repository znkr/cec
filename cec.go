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
	spy      chan<- Message
	spyDone  <-chan struct{}
	started  bool
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

// A Handler for HDMI CEC messages.
type Handler interface {
	// Handles a message and returns true if the message was handled. Once a message is handled, it
	// is considered done and is not passed to other handlers.
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
		// Standby is mandatory, but may be ignored. Returning here in order to not send a feature
		// abort.
		return true
	}

	// Send FeatureAbort if this message was directly addressed to us. Unhandled broadcasts are
	// ignored.
	if msg.Follower != Broadcast {
		log.Printf("Unexpected message: %s", msg)
		x.Reply(msg.Initiator, FeatureAbort{
			Abort:  msg.Cmd.Op(),
			Reason: AbortUnrecognizedOpCode,
		})
	}

	return true
}

// The DefaultHandler handles a set of standard messages. The handles messages are for physical
// address, vendor id, and CEC version.
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

// Adds a handler. If more than one handler is added all handlers are tried until one handler
// returns true. May only be called before Start() was called.
func (x *Cec) AddHandler(h Handler) {
	if x.started {
		log.Panic("Already started.")
	}
	x.handlers = append(x.handlers, h)
}

// Adds a handler in the form of a function. This is a convenience wrapper around AddHandler. May
// only be called before Start() was called.
func (x *Cec) AddHandleFunc(f func(x *Cec, msg Message) bool) {
	x.AddHandler(HandlerFunc(f))
}

// A Listener for all incoming and outgoing HDMI CEC messages.
type Listener interface {
	Message(msg Message)
}

// A function wrapper for Listener.
type ListenerFunc func(msg Message)

func (f ListenerFunc) Message(msg Message) {
	f(msg)
}

func spy(l Listener) (chan<- Message, <-chan struct{}) {
	c := make(chan Message, 64)
	done := make(chan struct{})
	go func() {
		for msg := range c {
			l.Message(msg)
		}
		close(done)
	}()
	return c, done
}

// Sets the listener. May only be called once before Start() was called.
func (x *Cec) SetListener(l Listener) {
	if x.started {
		log.Panic("Already started.")
	}
	if x.spy != nil {
		log.Panic("Listener already set.")
	}
	x.spy, x.spyDone = spy(l)
}

// Sets the listener. May only be called once before Start() was called.
func (x *Cec) SetListenerFunc(f func(msg Message)) {
	x.SetListener(ListenerFunc(f))
}

func (x *Cec) spyIncoming(msg Message) {
	if x.spy != nil {
		x.spy <- msg
	}
}

func (x *Cec) spyIncomingError(p Packet) {
	if x.spy != nil {
		x.spy <- Message{
			Initiator: p.Initiator,
			Follower:  p.Follower,
			Cmd:       MakeUnknownCmd(p.Op, p.Data),
		}
	}
}

// Starts receiving and handling CEC messages.
func (x *Cec) Run() {
	if x.started {
		log.Panic("Already started.")
	}
	x.started = true
	unhandledHandler := UnhandledHandler{}
	for p := range x.dev.Receive() {
		msg, err := UnmarshalMessage(p)
		if err != nil {
			x.spyIncomingError(p)
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
		x.spyIncoming(msg)

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
	if x.spy != nil {
		close(x.spy)
		<-x.spyDone
	}
}

func (x *Cec) spyOutgoing(follower LogicalAddr, cmd Command) {
	if x.spy != nil {
		x.spy <- Message{
			Initiator: x.dev.GetLogicalAddress(),
			Follower:  follower,
			Cmd:       cmd,
		}
	}
}

// Sends cmd to follower.
func (x *Cec) Send(follower LogicalAddr, cmd Command) error {
	data, err := cmd.Marshal()
	if err != nil {
		return err
	}
	x.spyOutgoing(follower, cmd)
	x.dev.Send(follower, cmd.Op(), data)
	return nil
}

// Sends cmd to follower as a reply.
func (x *Cec) Reply(follower LogicalAddr, cmd Command) error {
	data, err := cmd.Marshal()
	if err != nil {
		return err
	}
	x.spyOutgoing(follower, cmd)
	x.dev.Reply(follower, cmd.Op(), data)
	return nil
}
