package cec

import (
	"bytes"
	"reflect"
	"testing"
)

var addr = PhysicalAddress(0xabcd)

var cmdTests = []struct {
	name    string
	cmd     Command
	op      OpCode
	payload []byte
}{
	{"unkown_cmd", UnkownCmd{OpClearAnalogTimer, []byte{0x01}}, OpClearAnalogTimer, []byte{0x01}},
	{"feature_abort_default", FeatureAbort{}, OpFeatureAbort, []byte{0x00, 0x00}},
	{"feature_abort", FeatureAbort{OpSetOSDName, AbortNotInCorrectMode}, OpFeatureAbort, []byte{0x47, 0x01}},
	{"report_physical_address", ReportPhysicalAddress{PhysicalAddress(0xabcd), DeviceTypeSwitch}, OpReportPhysicalAddress, []byte{0xab, 0xcd, 0x06}},
	{"report_audio_status_0", ReportAudioStatus{0, false}, OpReportAudioStatus, []byte{0x00}},
	{"report_audio_status_0_muted", ReportAudioStatus{0, true}, OpReportAudioStatus, []byte{0x80}},
	{"report_audio_status_32", ReportAudioStatus{32, false}, OpReportAudioStatus, []byte{0x20}},
	{"report_audio_status_32_muted", ReportAudioStatus{32, true}, OpReportAudioStatus, []byte{0xa0}},
	{"report_audio_status_100", ReportAudioStatus{100, false}, OpReportAudioStatus, []byte{0x64}},
	{"report_audio_status_100_muted", ReportAudioStatus{100, true}, OpReportAudioStatus, []byte{0xe4}},
	{"report_audio_status_-1", ReportAudioStatus{-1, true}, OpReportAudioStatus, []byte{0xff}},
	{"report_audio_status_-1_muted", ReportAudioStatus{-1, false}, OpReportAudioStatus, []byte{0x7f}},
	{"report_power_status", ReportPowerStatus{PowerStatusOnTransition}, OpReportPowerStatus, []byte{0x02}},
	{"set_osd_name", SetOSDName{"osd name"}, OpSetOSDName, []byte("osd name")},
	{"set_system_audio_mode_false", SetSystemAudioMode{false}, OpSetSystemAudioMode, []byte{0x00}},
	{"set_system_audio_mode_true", SetSystemAudioMode{true}, OpSetSystemAudioMode, []byte{0x01}},
	{"give_audio_status", GiveAudioStatus{}, OpGiveAudioStatus, []byte{}},
	{"give_system_audio_mode_status", GiveSystemAudioModeStatus{}, OpGiveSystemAudioModeStatus, []byte{}},
	{"give_osd_name", GiveOSDName{}, OpGiveOSDName, []byte{}},
	{"give_device_power_status", GiveDevicePowerStatus{}, OpGiveDevicePowerStatus, []byte{}},
	{"give_device_vendor_id", GiveDeviceVendorID{}, OpGiveDeviceVendorID, []byte{}},
	{"give_physical_address", GivePhysicalAddress{}, OpGivePhysicalAddress, []byte{}},
	{"get_cec_version", GetCECVersion{}, OpGetCECVersion, []byte{}},
	{"system_audio_mode_request", SystemAudioModeRequest{nil}, OpSystemAudioModeRequest, []byte{}},
	{"sysetm_audio_mode_request_with_addr", SystemAudioModeRequest{&addr}, OpSystemAudioModeRequest, addr.Bytes()},
	{"device_vendor_id", DeviceVendorID{0xabcd}, OpDeviceVendorID, []byte{0x00, 0xab, 0xcd}},
	{"cec_version", CECVersion{cecVersion}, OpCECVersion, []byte{cecVersion}},
	{"user_control_pressed", UserControlPressed{UcBackward}, OpUserControlPressed, []byte{0x4c}},
	{"user_control_released", UserControlReleased{UcBackward}, OpUserControlReleased, []byte{0x4c}},
	{"standby", Standby{}, OpStandby, []byte{}},
	{"active_source", ActiveSource{}, OpActiveSource, []byte{}},
	{"vendor_command_with_id", VendorCommandWithID{}, OpVendorCommandWithID, []byte{}},
}

func TestCommand_Marshal(t *testing.T) {
	for _, test := range cmdTests {
		t.Run(test.name, func(t *testing.T) {
			payload, err := test.cmd.Marshal()
			if err != nil {
				t.Errorf("Failed to marshal %v: %s", test.cmd, err)
				return
			}
			if test.cmd.Op() != test.op {
				t.Errorf("OpCode for Cmd %T is %s, expected %s", test.cmd, test.cmd.Op(), test.op)
			}
			if !bytes.Equal(payload, test.payload) {
				t.Errorf("Cmd %v marshaled payload %#v, expected %#v", test.cmd, payload, test.payload)
			}
		})
	}
}

func TestUnmarshalMessage(t *testing.T) {
	initiator := TV
	follower := Playback3
	for _, test := range cmdTests {
		t.Run(test.name, func(t *testing.T) {
			p := Packet{
				Initiator: initiator,
				Follower:  follower,
				Op:        test.op,
				Data:      test.payload,
			}
			m, err := UnmarshalMessage(p)
			if err != nil {
				t.Errorf("Failed to unmarschal %s: %s", p, err)
				return
			}
			if m.Initiator != initiator {
				t.Errorf("Incorrect initiator %s, expected %s", m.Initiator, initiator)
			}
			if m.Follower != follower {
				t.Errorf("Incorrect follower %s, expected %s", m.Follower, follower)
			}
			if !reflect.DeepEqual(m.Cmd, test.cmd) {
				t.Errorf("Incorrect Cmd %T %v, expected %T %v", m.Cmd, m.Cmd, test.cmd, test.cmd)
			}
		})
	}
}

func TestCommand_Marshal_Fail(t *testing.T) {
	tests := []struct {
		name string
		cmd  Command
		err  error
	}{
		{"invalid_volume", ReportAudioStatus{101, false}, InvalidVolume{}},
		{"empty_osd_name", SetOSDName{""}, InvalidOSDName{}},
		{"osd_name_too_long", SetOSDName{"toolongtooolong"}, InvalidOSDName{}},
		{"device_id_too_large", DeviceVendorID{0xabcdef00}, InvalidVendorId{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.cmd.Marshal()
			if err == nil {
				t.Errorf("Marshaling was successful, expected error.")
				return
			}
			if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
				t.Errorf("Expected error to be of type %T, but it is %T (%s).", test.err, err, err)
			}
		})
	}
}

func TestUnmarshalMessage_Fail(t *testing.T) {
	tests := []struct {
		name    string
		op      OpCode
		payload []byte
		err     error
	}{
		{"feature_abort_no_payload", OpFeatureAbort, []byte{}, IncorrectPacketDataLength{}},
		{"feature_abort_payload_too_short", OpFeatureAbort, []byte{0x00}, IncorrectPacketDataLength{}},
		{"feature_abort_payload_too_long", OpFeatureAbort, []byte{0x00, 0x00, 0x00}, IncorrectPacketDataLength{}},
		{"report_physical_address_no_payload", OpReportPhysicalAddress, []byte{}, IncorrectPacketDataLength{}},
		{"report_audio_status_no_payload", OpReportAudioStatus, []byte{}, IncorrectPacketDataLength{}},
		{"report_audio_status_invalid_volume", OpReportAudioStatus, []byte{0x69}, InvalidVolume{}},
		{"report_power_status_no_payload", OpReportPowerStatus, []byte{}, IncorrectPacketDataLength{}},
		{"set_osd_name_too_long", OpSetOSDName, []byte("toolongtooolong"), InvalidOSDName{}},
		{"set_osd_name_too_short", OpSetOSDName, []byte(""), InvalidOSDName{}},
		{"set_osd_name_utf8", OpSetOSDName, []byte("fäil"), InvalidOSDName{}},
		{"system_audio_mode_request_invalid_payload", OpSystemAudioModeRequest, []byte{0x00}, IncorrectPacketDataLength{}},
		{"device_vendor_id_payload_too_short", OpDeviceVendorID, []byte{0x00}, IncorrectPacketDataLength{}},
		{"cec_version_no_payload", OpCECVersion, []byte{}, IncorrectPacketDataLength{}},
		{"user_control_pressed_no_payload", OpUserControlPressed, []byte{}, IncorrectPacketDataLength{}},
		{"user_control_released_no_payload", OpUserControlReleased, []byte{}, IncorrectPacketDataLength{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := Packet{
				Initiator: TV,
				Follower:  Playback3,
				Op:        test.op,
				Data:      test.payload,
			}
			_, err := UnmarshalMessage(p)
			if err == nil {
				t.Errorf("Unmarschal was successful, expected error.")
				return
			}
			if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
				t.Errorf("Expected error to be of type %T, but it is %T (%s).", test.err, err, err)
			}
		})
	}
}

func TestMessage_String(t *testing.T) {
	tests := []struct {
		name      string
		initiator LogicalAddr
		follower  LogicalAddr
		cmd       Command
		want      string
	}{
		{"direct", TV, AudioSystem, GiveAudioStatus{}, "TV → AudioSystem: cec.GiveAudioStatus {}"},
		{"from_unregistered", Unregistered, AudioSystem, GiveAudioStatus{}, "Unregistered → AudioSystem: cec.GiveAudioStatus {}"},
		{"to_broadcast", TV, Broadcast, GiveAudioStatus{}, "TV → Broadcast: cec.GiveAudioStatus {}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Message{
				Initiator: tt.initiator,
				Follower:  tt.follower,
				Cmd:       tt.cmd,
			}
			if got := m.String(); got != tt.want {
				t.Errorf("Message.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
