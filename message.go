package cec

import (
	"fmt"
)

type IncorrectPacketDataLength struct {
	expected int
	actual   int
}

func (e IncorrectPacketDataLength) Error() string {
	return fmt.Sprintf("Incorrect data length; expected %d, actual %d.", e.expected, e.actual)
}

type InvalidOSDName struct{}

func (e InvalidOSDName) Error() string {
	return fmt.Sprintf("Invalid data for OSD name.")
}

type InvalidVendorId struct{}

func (e InvalidVendorId) Error() string {
	return fmt.Sprintf("Invalid vendor id.")
}

type InvalidVolume struct {
	volume int
}

func (e InvalidVolume) Error() string {
	return fmt.Sprintf("Invalid volume: %d", e.volume)
}

// A Message is a representation of an HDMI CEC message.
type Message struct {
	Initiator LogicalAddr // The sender of this message.
	Follower  LogicalAddr // The receiver of this message.
	Cmd       Command     // The HDMI CEC command.
}

// Message implements the Stringer interface.
func (m Message) String() string {
	// Broadcast and Unregistered share the same value, the meaning depends on context.
	i := m.Initiator.String()
	f := m.Follower.String()
	if m.Initiator == Unregistered {
		i = "Unregistered"
	}
	if m.Follower == Broadcast {
		f = "Broadcast"
	}

	s := fmt.Sprintf("%v", m.Cmd)
	if s == "{{}}" {
		s = "{}" // Cleanup emptyCommand results
	}

	return fmt.Sprintf("%s → %s: %T %v", i, f, m.Cmd, s)
}

// Creates a message from a given Packet.
func UnmarshalMessage(p Packet) (Message, error) {
	cmd, err := unmarshalCommand(p.Op, p.Data)
	if err != nil {
		return Message{}, err
	}

	return Message{
		Initiator: p.Initiator,
		Follower:  p.Follower,
		Cmd:       cmd,
	}, nil
}

type (
	// Command is an interface representing an HDMI CEC command.
	Command interface {
		Op() OpCode // The OpCode of this command.
		Marshal() ([]byte, error)
	}

	emptyCommand struct{}

	// An unknown command.
	UnkownCmd struct {
		op   OpCode
		data []byte
	}

	// FeatureAbort is used to communicate an errors.
	FeatureAbort struct {
		Abort  OpCode
		Reason AbortReason
	}

	// ReportPhysicalAddress is used to report the physical address of a device. It's usually send in reply to a
	// GivePhysicalAddress command.
	ReportPhysicalAddress struct {
		Addr PhysicalAddress // The physical address of the reporting device.
		Type DeviceType      // The device type of the reporting device.
	}

	// ReportAudioStatus is used report the audio status.
	ReportAudioStatus struct {
		Volume int  // The volume between 0 and 100 or a negative value if the volume is unknown.
		Muted  bool // Whether the audio is muted or not.
	}

	// ReportPowerStatus is used to report the power status of a device.
	ReportPowerStatus struct {
		Power PowerStatus
	}

	// Sets the OSD name of this device. This is usually used in response to a GiveOSDName.
	SetOSDName struct {
		Name string // The OSD name, must be 1 to 14 ASCII characters.
	}

	// Sets the system audio mode for this device. This is usually used in response to SystemAudioModeRequest, but can
	// be send outside of a reply as well.
	SetSystemAudioMode struct {
		On bool
	}

	// Requests the current AudioStatus from a device. This should be answered with a ReportAudioStatus command.
	GiveAudioStatus struct {
		emptyCommand
	}

	// Requests the system audio mode status. This should be answered with a SetSystemAudioMode command.
	GiveSystemAudioModeStatus struct {
		emptyCommand
	}

	// Requests the OSD name for this device. This should be answered with a SetOSDName command.
	GiveOSDName struct {
		emptyCommand
	}

	// Requests the power status of a device. This should be answered with ReportPowerStatus.
	GiveDevicePowerStatus struct {
		emptyCommand
	}

	// Requests the device vendor ID of a device. This should be answered with DeviceVendorID.
	GiveDeviceVendorID struct {
		emptyCommand
	}

	// Requests the physical address of a device. This should be answered with ReportPhysicalAddress.
	GivePhysicalAddress struct {
		emptyCommand
	}

	// Requests the CEC version a device has implemented. This should be answered with CECVersion.
	GetCECVersion struct {
		emptyCommand
	}

	// Requests that a device initiates system audio control for the device in Addr. This is usually called by the TV.
	SystemAudioModeRequest struct {
		// The physical address of the device to use for system audio control or nil if the address is not set.
		Addr *PhysicalAddress
	}

	// Reports the vendor ID of this device. This is usually send in response to GiveDeviceVendorID.
	DeviceVendorID struct {
		VendorID uint32 // The vendor id.
	}

	// Reports the cec version this device implements. This is usually send in response to GiveCECVersionID.
	CECVersion struct {
		Version byte
	}

	// Reports that the user pressed a control.
	UserControlPressed struct {
		Pressed UserControl // The control that was pressed.
	}

	// Reports that the user released a control.
	UserControlReleased struct {
		Released UserControl // The control that was released.
	}

	// Requests standby.
	Standby struct {
		emptyCommand
	}

	// TODO: Not yet implemented.
	ActiveSource struct {
		emptyCommand
	}

	// TODO: Not yet implemented.
	VendorCommandWithID struct {
		emptyCommand
	}
)

func unmarshalCommand(op OpCode, data []byte) (Command, error) {
	switch op {
	case OpActiveSource:
		return ActiveSource{}, nil

	case OpFeatureAbort:
		if len(data) != 2 {
			return nil, IncorrectPacketDataLength{3, len(data)}
		}
		return FeatureAbort{
			Abort:  OpCode(data[0]),
			Reason: AbortReason(data[1]),
		}, nil

	case OpReportPhysicalAddress:
		if len(data) != 3 {
			return nil, IncorrectPacketDataLength{3, len(data)}
		}
		return ReportPhysicalAddress{
			Addr: PhysicalAddress(int(data[0])<<8 | int(data[1])),
			Type: DeviceType(data[2]),
		}, nil

	case OpReportAudioStatus:
		if len(data) != 1 {
			return nil, IncorrectPacketDataLength{1, len(data)}
		}
		v := int(data[0] & 0x7f)
		// The values between 0x65 and 0x7f are reserved for future use and 0x7f is defined as volume unknown.
		if v == 0x7f {
			v = -1
		} else if v > 0x64 {
			return nil, InvalidVolume{v}
		}
		return ReportAudioStatus{
			Volume: v,
			Muted:  data[0]&0x80 == 0x80,
		}, nil

	case OpReportPowerStatus:
		if len(data) != 1 {
			return nil, IncorrectPacketDataLength{1, len(data)}
		}
		return ReportPowerStatus{
			Power: PowerStatus(data[0]),
		}, nil

	case OpSetOSDName:
		s := string(data)
		if !isValidOsdName(s) {
			return nil, InvalidOSDName{}
		}
		return SetOSDName{
			Name: s,
		}, nil

	case OpSetSystemAudioMode:
		if len(data) != 1 {
			return nil, IncorrectPacketDataLength{1, len(data)}
		}
		return SetSystemAudioMode{
			// data[0] must be either 1 or 0, parsing this a bit more lenient might cause problems, but so might parsing
			// this strictly...
			On: data[0] != 0,
		}, nil

	case OpGiveOSDName:
		return GiveOSDName{}, nil

	case OpGiveDevicePowerStatus:
		return GiveDevicePowerStatus{}, nil

	case OpGiveDeviceVendorID:
		return GiveDeviceVendorID{}, nil

	case OpGivePhysicalAddress:
		return GivePhysicalAddress{}, nil

	case OpGiveAudioStatus:
		return GiveAudioStatus{}, nil

	case OpGiveSystemAudioModeStatus:
		return GiveSystemAudioModeStatus{}, nil

	case OpSystemAudioModeRequest:
		if len(data) != 0 && len(data) != 2 {
			return nil, IncorrectPacketDataLength{2, len(data)}
		}
		var addr *PhysicalAddress
		if len(data) == 2 {
			addr = new(PhysicalAddress)
			*addr = PhysicalAddress(int(data[0])<<8 | int(data[1]))
		}
		return SystemAudioModeRequest{
			Addr: addr,
		}, nil

	case OpGetCECVersion:
		return GetCECVersion{}, nil

	case OpDeviceVendorID:
		if len(data) != 3 {
			return nil, IncorrectPacketDataLength{3, len(data)}
		}
		return DeviceVendorID{
			VendorID: uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2]),
		}, nil

	case OpCECVersion:
		if len(data) != 1 {
			return nil, IncorrectPacketDataLength{1, len(data)}
		}
		return CECVersion{
			Version: data[0],
		}, nil

	case OpUserControlPressed:
		if len(data) != 1 {
			return nil, IncorrectPacketDataLength{1, len(data)}
		}
		return UserControlPressed{
			Pressed: UserControl(data[0]),
		}, nil

	case OpUserControlReleased:
		if len(data) != 1 {
			return nil, IncorrectPacketDataLength{1, len(data)}
		}
		return UserControlReleased{
			Released: UserControl(data[0]),
		}, nil

	case OpVendorCommandWithID:
		return VendorCommandWithID{}, nil

	case OpStandby:
		return Standby{}, nil

	default:
		return UnkownCmd{
			op:   op,
			data: data,
		}, nil
	}
}

func (c UnkownCmd) Op() OpCode                 { return c.op }
func (c ActiveSource) Op() OpCode              { return OpActiveSource }
func (c FeatureAbort) Op() OpCode              { return OpFeatureAbort }
func (c ReportPhysicalAddress) Op() OpCode     { return OpReportPhysicalAddress }
func (c ReportAudioStatus) Op() OpCode         { return OpReportAudioStatus }
func (c ReportPowerStatus) Op() OpCode         { return OpReportPowerStatus }
func (c SetOSDName) Op() OpCode                { return OpSetOSDName }
func (c SetSystemAudioMode) Op() OpCode        { return OpSetSystemAudioMode }
func (c GiveOSDName) Op() OpCode               { return OpGiveOSDName }
func (c GiveDevicePowerStatus) Op() OpCode     { return OpGiveDevicePowerStatus }
func (c GiveDeviceVendorID) Op() OpCode        { return OpGiveDeviceVendorID }
func (c GivePhysicalAddress) Op() OpCode       { return OpGivePhysicalAddress }
func (c GiveSystemAudioModeStatus) Op() OpCode { return OpGiveSystemAudioModeStatus }
func (c GiveAudioStatus) Op() OpCode           { return OpGiveAudioStatus }
func (c GetCECVersion) Op() OpCode             { return OpGetCECVersion }
func (c SystemAudioModeRequest) Op() OpCode    { return OpSystemAudioModeRequest }
func (c DeviceVendorID) Op() OpCode            { return OpDeviceVendorID }
func (c CECVersion) Op() OpCode                { return OpCECVersion }
func (c VendorCommandWithID) Op() OpCode       { return OpVendorCommandWithID }
func (c Standby) Op() OpCode                   { return OpStandby }
func (c UserControlPressed) Op() OpCode        { return OpUserControlPressed }
func (c UserControlReleased) Op() OpCode       { return OpUserControlReleased }

func (c emptyCommand) Marshal() ([]byte, error) { return []byte{}, nil }

func (c UnkownCmd) Marshal() ([]byte, error) { return c.data, nil }

func (c FeatureAbort) Marshal() ([]byte, error) {
	return []byte{byte(c.Abort), byte(c.Reason)}, nil
}

func (c ReportPhysicalAddress) Marshal() ([]byte, error) {
	return append(c.Addr.Bytes(), byte(c.Type)), nil
}

func (c ReportAudioStatus) Marshal() ([]byte, error) {
	if c.Volume > 100 {
		return nil, InvalidVolume{c.Volume}
	}
	var data byte
	if c.Volume < 0 {
		data = byte(0x7f) // Volume unknown
	} else {
		data = byte(c.Volume)
	}
	if c.Muted {
		data |= 0x80
	}
	return []byte{data}, nil
}

func (c ReportPowerStatus) Marshal() ([]byte, error) {
	return []byte{byte(c.Power)}, nil
}

func (c SetOSDName) Marshal() ([]byte, error) {
	if !isValidOsdName(c.Name) {
		return nil, InvalidOSDName{}
	}
	return []byte(c.Name), nil
}

func (c SetSystemAudioMode) Marshal() ([]byte, error) {
	if c.On {
		return []byte{0x01}, nil
	} else {
		return []byte{0x00}, nil
	}
}

func (c DeviceVendorID) Marshal() ([]byte, error) {
	if !isValidVendorId(c.VendorID) {
		return nil, InvalidVendorId{}
	}
	id := c.VendorID
	return []byte{
		byte((id >> 16) & 0xff),
		byte((id >> 8) & 0xff),
		byte((id >> 0) & 0xff),
	}, nil
}

func (c CECVersion) Marshal() ([]byte, error) { return []byte{c.Version}, nil }

func (c SystemAudioModeRequest) Marshal() ([]byte, error) {
	if c.Addr != nil {
		return c.Addr.Bytes(), nil
	} else {
		return []byte{}, nil
	}
}

func (c UserControlPressed) Marshal() ([]byte, error) {
	return []byte{byte(c.Pressed)}, nil
}

func (c UserControlReleased) Marshal() ([]byte, error) {
	return []byte{byte(c.Released)}, nil
}
