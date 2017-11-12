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
			t.Errorf("%s should be valid, but isValidOsdName returns false.", s)
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
			t.Errorf("%s should be invalid, but isValidOsdName returns true.", s)
		}
	}
}

func TestIsValidVendorID(t *testing.T) {
	invalid := uint32(0x01000000)
	if isValidVendorId(invalid) {
		t.Errorf("%s should be invalid, but isValidVendorId return true.", invalid)
	}
}
