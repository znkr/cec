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

import (
	"bytes"
	"testing"
)

func TestPhysicalAddress_Bytes(t *testing.T) {
	addr := PhysicalAddress(0xabcd)
	bs := addr.Bytes()
	expected := []byte{0xab, 0xcd}
	if !bytes.Equal(bs, expected) {
		t.Errorf("Not true that %q == %q", bs, expected)
	}
}

func TestPhysicalAddress_String(t *testing.T) {
	addr := PhysicalAddress(0xabcd)
	s := addr.String()
	expected := "a.b.c.d"
	if s != expected {
		t.Errorf("Not true that %q == %q.", s, expected)
	}
}
