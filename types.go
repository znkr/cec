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

//go:generate stringer -type=PowerStatus
//go:generate stringer -type=OpCode
//go:generate stringer -type=UserControl
//go:generate stringer -type=LogicalAddr
//go:generate stringer -type=DeviceType
//go:generate stringer -type=AbortReason

import "fmt"

// A physical HDMI CEC address.
type PhysicalAddress uint16

func (a PhysicalAddress) String() string {
	a0 := int((a >> 12) & 0x0f)
	a1 := int((a >> 8) & 0x0f)
	a2 := int((a >> 4) & 0x0f)
	a3 := int((a >> 0) & 0x0f)
	return fmt.Sprintf("%x.%x.%x.%x", a0, a1, a2, a3)
}

func (a PhysicalAddress) Bytes() []byte {
	return []byte{
		byte((a >> 8) & 0xff),
		byte((a >> 0) & 0xff),
	}
}

// A logical HDMI CEC address.
type LogicalAddr byte

const (
	TV           LogicalAddr = 0x0
	Rec1         LogicalAddr = 0x1
	Rec2         LogicalAddr = 0x2
	Tuner1       LogicalAddr = 0x3
	Playback1    LogicalAddr = 0x4
	AudioSystem  LogicalAddr = 0x5
	Tuner2       LogicalAddr = 0x6
	Tuner3       LogicalAddr = 0x7
	Playback2    LogicalAddr = 0x8
	Rec3         LogicalAddr = 0x9
	Tuner4       LogicalAddr = 0xa
	Playback3    LogicalAddr = 0xb
	Reserved1    LogicalAddr = 0xc
	Reserved2    LogicalAddr = 0xd
	FreeUse      LogicalAddr = 0xe
	Unregistered LogicalAddr = 0xf
	Broadcast    LogicalAddr = 0xf
)

// A HDMI CEC device type.
type DeviceType byte

const (
	DeviceTypeTV       DeviceType = 0x0
	DeviceTypeRec      DeviceType = 0x1
	DeviceTypeReserved DeviceType = 0x2
	DeviceTypeTuner    DeviceType = 0x3
	DeviceTypePlayback DeviceType = 0x4
	DeviceTypeAudio    DeviceType = 0x5
	DeviceTypeSwitch   DeviceType = 0x6
	DeviceTypeVidProc  DeviceType = 0x7
	DeviceTypeInvalid  DeviceType = 0xf
)

// Representation of the power status.
type PowerStatus byte

const (
	PowerStatusOn                PowerStatus = 0x00
	PowerStatusStandby           PowerStatus = 0x01
	PowerStatusOnTransition      PowerStatus = 0x02
	PowerStatusStandbyTransition PowerStatus = 0x03
)

// Representation of an HDMI CEC opcode.
type OpCode byte

const (
	OpFeatureAbort              OpCode = 0x00
	OpImageViewOn               OpCode = 0x04
	OpTunerStepIncrement        OpCode = 0x05
	OpTunerStepDecrement        OpCode = 0x06
	OpTunerDeviceStatus         OpCode = 0x07
	OpGiveTunerDeviceStatus     OpCode = 0x08
	OpRecordOn                  OpCode = 0x09
	OpRecordStatus              OpCode = 0x0A
	OpRecordOff                 OpCode = 0x0B
	OpTextViewOn                OpCode = 0x0D
	OpRecordTVScreen            OpCode = 0x0F
	OpGiveDeckStatus            OpCode = 0x1A
	OpDeckStatus                OpCode = 0x1B
	OpSetMenuLanguage           OpCode = 0x32
	OpClearAnalogTimer          OpCode = 0x33
	OpSetAnalogTimer            OpCode = 0x34
	OpTimerStatus               OpCode = 0x35
	OpStandby                   OpCode = 0x36
	OpPlay                      OpCode = 0x41
	OpDeckControl               OpCode = 0x42
	OpTimerClearedStatus        OpCode = 0x43
	OpUserControlPressed        OpCode = 0x44
	OpUserControlReleased       OpCode = 0x45
	OpGiveOSDName               OpCode = 0x46
	OpSetOSDName                OpCode = 0x47
	OpSetOSDString              OpCode = 0x64
	OpSetTimerProgramTitle      OpCode = 0x67
	OpSystemAudioModeRequest    OpCode = 0x70
	OpGiveAudioStatus           OpCode = 0x71
	OpSetSystemAudioMode        OpCode = 0x72
	OpReportAudioStatus         OpCode = 0x7A
	OpGiveSystemAudioModeStatus OpCode = 0x7D
	OpSystemAudioModeStatus     OpCode = 0x7E
	OpRoutingChange             OpCode = 0x80
	OpRoutingInformation        OpCode = 0x81
	OpActiveSource              OpCode = 0x82
	OpGivePhysicalAddress       OpCode = 0x83
	OpReportPhysicalAddress     OpCode = 0x84
	OpRequestActiveSource       OpCode = 0x85
	OpSetStreamPath             OpCode = 0x86
	OpDeviceVendorID            OpCode = 0x87
	OpVendorCommand             OpCode = 0x89
	OpVendorRemoteButtonDown    OpCode = 0x8A
	OpVendorRemoteButtonUp      OpCode = 0x8B
	OpGiveDeviceVendorID        OpCode = 0x8C
	OpMenuRequest               OpCode = 0x8D
	OpMenuStatus                OpCode = 0x8E
	OpGiveDevicePowerStatus     OpCode = 0x8F
	OpReportPowerStatus         OpCode = 0x90
	OpGetMenuLanguage           OpCode = 0x91
	OpSelectAnalogService       OpCode = 0x92
	OpSelectDigitalService      OpCode = 0x93
	OpSetDigitalTimer           OpCode = 0x97
	OpClearDigitalTimer         OpCode = 0x99
	OpSetAudioRate              OpCode = 0x9A
	OpInactiveSource            OpCode = 0x9D
	OpCECVersion                OpCode = 0x9E
	OpGetCECVersion             OpCode = 0x9F
	OpVendorCommandWithID       OpCode = 0xA0
	OpClearExternalTimer        OpCode = 0xA1
	OpSetExternalTimer          OpCode = 0xA2
	OpAbort                     OpCode = 0xFF
)

// Representation of a user control.
type UserControl byte

const (
	UcSelect                   UserControl = 0x00
	UcUp                       UserControl = 0x01
	UcDown                     UserControl = 0x02
	UcLeft                     UserControl = 0x03
	UcRight                    UserControl = 0x04
	UcRightUp                  UserControl = 0x05
	UcRightDown                UserControl = 0x06
	UcLeftUp                   UserControl = 0x07
	UcLeftDown                 UserControl = 0x08
	UcRootMenu                 UserControl = 0x09
	UcSetupMenu                UserControl = 0x0A
	UcContentsMenu             UserControl = 0x0B
	UcFavoriteMenu             UserControl = 0x0C
	UcExit                     UserControl = 0x0D
	UcNumber0                  UserControl = 0x20
	UcNumber1                  UserControl = 0x21
	UcNumber2                  UserControl = 0x22
	UcNumber3                  UserControl = 0x23
	UcNumber4                  UserControl = 0x24
	UcNumber5                  UserControl = 0x25
	UcNumber6                  UserControl = 0x26
	UcNumber7                  UserControl = 0x27
	UcNumber8                  UserControl = 0x28
	UcNumber9                  UserControl = 0x29
	UcDot                      UserControl = 0x2A
	UcEnter                    UserControl = 0x2B
	UcClear                    UserControl = 0x2C
	UcChannelUp                UserControl = 0x30
	UcChannelDown              UserControl = 0x31
	UcPreviousChannel          UserControl = 0x32
	UcSoundSelect              UserControl = 0x33
	UcInputSelect              UserControl = 0x34
	UcDisplayInformation       UserControl = 0x35
	UcHelp                     UserControl = 0x36
	UcPageUp                   UserControl = 0x37
	UcPageDown                 UserControl = 0x38
	UcPower                    UserControl = 0x40
	UcVolumeUp                 UserControl = 0x41
	UcVolumeDown               UserControl = 0x42
	UcMute                     UserControl = 0x43
	UcPlay                     UserControl = 0x44
	UcStop                     UserControl = 0x45
	UcPause                    UserControl = 0x46
	UcRecord                   UserControl = 0x47
	UcRewind                   UserControl = 0x48
	UcFastForward              UserControl = 0x49
	UcEject                    UserControl = 0x4A
	UcForward                  UserControl = 0x4B
	UcBackward                 UserControl = 0x4C
	UcAngle                    UserControl = 0x50
	UcSubpicture               UserControl = 0x51
	UcVideoOnDemand            UserControl = 0x52
	UcEPG                      UserControl = 0x53
	UcTimerProgramming         UserControl = 0x54
	UcInitialConfig            UserControl = 0x55
	UcPlayFunction             UserControl = 0x60
	UcPausePlayFunction        UserControl = 0x61
	UcRecordFunction           UserControl = 0x62
	UcPauseRecordFunction      UserControl = 0x63
	UcStopFunction             UserControl = 0x64
	UcMuteFunction             UserControl = 0x65
	UcRestoreVolumeFunction    UserControl = 0x66
	UcTuneFunction             UserControl = 0x67
	UcSelectDiskFunction       UserControl = 0x68
	UcSelectAVInputFunction    UserControl = 0x69
	UcSelectAudioInputFunction UserControl = 0x6A
	UcF1Blue                   UserControl = 0x71
	UcF2Red                    UserControl = 0x72
	UcF3Green                  UserControl = 0x73
	UcF4Yellow                 UserControl = 0x74
	UcF5                       UserControl = 0x75
)

// Abort reason in a feature abort context
type AbortReason byte

const (
	AbortUnrecognizedOpCode  AbortReason = 0x00
	AbortNotInCorrectMode    AbortReason = 0x01
	AbortCannotProvideSource AbortReason = 0x02
	AbortInvalidOperand      AbortReason = 0x03
	AbortRefused             AbortReason = 0x04
)

type opCodeFlags int

const (
	fDirect opCodeFlags = 1 << iota
	fBroadcast
	fBroadcastResponse
	fSwitchMessage
)

var opCodeMeta = map[OpCode]opCodeFlags{
	OpFeatureAbort:              fDirect,
	OpImageViewOn:               fDirect,
	OpTunerStepIncrement:        fDirect,
	OpTunerStepDecrement:        fDirect,
	OpTunerDeviceStatus:         fDirect,
	OpGiveTunerDeviceStatus:     fDirect,
	OpRecordOn:                  fDirect,
	OpRecordStatus:              fDirect,
	OpRecordOff:                 fDirect,
	OpTextViewOn:                fDirect,
	OpRecordTVScreen:            fDirect,
	OpGiveDeckStatus:            fDirect,
	OpDeckStatus:                fDirect,
	OpSetMenuLanguage:           fBroadcast,
	OpClearAnalogTimer:          fDirect,
	OpSetAnalogTimer:            fDirect,
	OpTimerStatus:               fDirect,
	OpStandby:                   fBroadcast | fDirect,
	OpPlay:                      fDirect,
	OpDeckControl:               fDirect,
	OpTimerClearedStatus:        fDirect,
	OpUserControlPressed:        fDirect,
	OpUserControlReleased:       fDirect,
	OpGiveOSDName:               fDirect,
	OpSetOSDName:                fDirect,
	OpSetOSDString:              fDirect,
	OpSetTimerProgramTitle:      fDirect,
	OpSystemAudioModeRequest:    fDirect,
	OpGiveAudioStatus:           fDirect,
	OpSetSystemAudioMode:        fBroadcast | fDirect,
	OpReportAudioStatus:         fDirect,
	OpGiveSystemAudioModeStatus: fDirect,
	OpSystemAudioModeStatus:     fDirect,
	OpRoutingChange:             fBroadcast | fSwitchMessage,
	OpRoutingInformation:        fBroadcast | fSwitchMessage,
	OpActiveSource:              fBroadcast,
	OpGivePhysicalAddress:       fDirect | fBroadcastResponse,
	OpReportPhysicalAddress:     fBroadcast,
	OpRequestActiveSource:       fBroadcast,
	OpSetStreamPath:             fBroadcast,
	OpDeviceVendorID:            fBroadcast,
	OpVendorCommand:             fDirect,
	OpVendorRemoteButtonDown:    fBroadcast | fDirect,
	OpVendorRemoteButtonUp:      fBroadcast | fDirect,
	OpGiveDeviceVendorID:        fDirect | fBroadcastResponse,
	OpMenuRequest:               fDirect,
	OpMenuStatus:                fDirect,
	OpGiveDevicePowerStatus:     fDirect,
	OpReportPowerStatus:         fDirect,
	OpGetMenuLanguage:           fDirect | fBroadcastResponse,
	OpSelectAnalogService:       fDirect,
	OpSelectDigitalService:      fDirect,
	OpSetDigitalTimer:           fDirect,
	OpClearDigitalTimer:         fDirect,
	OpSetAudioRate:              fDirect,
	OpInactiveSource:            fDirect,
	OpCECVersion:                fDirect,
	OpGetCECVersion:             fDirect,
	OpVendorCommandWithID:       fBroadcast | fDirect,
	OpClearExternalTimer:        fDirect,
	OpSetExternalTimer:          fDirect,
	OpAbort:                     fDirect,
}

func getOpCodeFlags(op OpCode) (flags opCodeFlags, ok bool) {
	flags, ok = opCodeMeta[op]
	return
}
