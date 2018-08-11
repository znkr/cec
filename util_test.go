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

import "testing"

func TestIsValidOsdNameValid(t *testing.T) {
	valid := []string{
		"a",
		"bc",
		"def",
		"ghij",
		"klmno",
		"pqrstu",
		"vwxyz01",
		"23456789",
		"ZYXWVUTSRQPONM",
		" LKJIHGFEDCBA~",
	}
	for _, s := range valid {
		if !isValidOsdName(s) {
			t.Errorf("%v should be valid, but isValidOsdName returns false.", s)
		}
	}
}

func TestIsValidOsdNameInvalid(t *testing.T) {
	invalid := []string{
		"",                // Too short
		"aaaaaaaaaaaaaaa", // Too long
		"fäil",            // UTF-8 is not supported.
	}
	for _, s := range invalid {
		if isValidOsdName(s) {
			t.Errorf("%v should be invalid, but isValidOsdName returns true.", s)
		}
	}
}

func TestIsValidVendorID(t *testing.T) {
	invalid := uint32(0x01000000)
	if isValidVendorId(invalid) {
		t.Errorf("%v should be invalid, but isValidVendorId return true.", invalid)
	}
}
