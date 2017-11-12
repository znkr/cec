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
