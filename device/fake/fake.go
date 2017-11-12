// A fake CEC device.
package fake

import (
	"github.com/krynr/cec"
)

type Device struct {
	addr cec.LogicalAddr
	typ  cec.DeviceType
	ci   chan cec.Packet
	co   chan cec.Packet
}

const PhysicalAddress cec.PhysicalAddress = 0xabcd
const VendorDeviceID = 0x101010

func New(addr cec.LogicalAddr, typ cec.DeviceType) *Device {
	return &Device{
		addr: addr,
		typ:  typ,
		ci:   make(chan cec.Packet, 10),
		co:   make(chan cec.Packet, 10),
	}
}

func (d *Device) Run(in []cec.Packet, await func()) (out []cec.Packet) {
	done := make(chan struct{})
	go func() {
		for _, p := range in {
			d.ci <- p
		}
		close(d.ci)
	}()
	go func() {
		for p := range d.co {
			out = append(out, p)
		}
		close(done)
	}()
	await()
	close(d.co)
	<-done
	return out
}

func (d *Device) Receive() <-chan cec.Packet {
	return d.ci
}

func (d *Device) Send(follower cec.LogicalAddr, op cec.OpCode, payload []byte) {
	d.co <- cec.Packet{
		Initiator: d.addr,
		Follower:  follower,
		Op:        op,
		Data:      payload,
	}
}

func (d *Device) Reply(follower cec.LogicalAddr, op cec.OpCode, payload []byte) {
	d.Send(follower, op, payload)
}

func (d *Device) GetVendorID() uint32 {
	return VendorDeviceID
}

func (d *Device) GetDeviceType() cec.DeviceType {
	return d.typ
}

func (d *Device) GetPhysicalAddress() cec.PhysicalAddress {
	return PhysicalAddress
}

func (d *Device) GetLogicalAddress() cec.LogicalAddr {
	return d.addr
}
