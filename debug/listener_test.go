// Copyright 2018 Google LLC
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

// Package debug provides utils to help debugging HDMI
package debug

import (
	"testing"

	"znkr.io/cec"
)

func TestLoggingHandler(t *testing.T) {
	l := NewLoggingListener(4)
	l.Message(cec.Message{
		Initiator: cec.TV,
		Follower:  cec.AudioSystem,
		Cmd:       cec.GivePhysicalAddress{},
	})
	l.Message(cec.Message{
		Initiator: cec.AudioSystem,
		Follower:  cec.Broadcast,
		Cmd:       cec.ReportPhysicalAddress{cec.PhysicalAddress(0xabcd), cec.DeviceTypeAudio},
	})

	if len(l.GetLogged()) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(l.GetLogged()))
	}
}
