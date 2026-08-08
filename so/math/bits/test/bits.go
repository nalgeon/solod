package bits_test

import (
	"solod.dev/so/math/bits"
	"solod.dev/so/testing"
)

func TestAdd32(t *testing.T) {
	n1 := uint32(0b0101)
	n2 := uint32(0b0011)
	d, carry := bits.Add32(n1, n2, 0)
	if d != 0b1000 || carry != 0 {
		t.Error("Add32 failed")
	}
}

func TestSub32(t *testing.T) {
	n1 := uint32(0b0101)
	n2 := uint32(0b0011)
	d, borrow := bits.Sub32(n1, n2, 0)
	if d != 0b0010 || borrow != 0 {
		t.Error("Sub32 failed")
	}
}

func TestMul32(t *testing.T) {
	n1 := uint32(0b0101)
	n2 := uint32(0b0011)
	dh, dl := bits.Mul32(n1, n2)
	if dh != 0 || dl != 0b1111 {
		t.Error("Mul32 failed")
	}
}

func TestLeadingZeros8(t *testing.T) {
	n := uint8(0b00010000)
	if bits.LeadingZeros8(n) != 3 {
		t.Error("LeadingZeros8 failed")
	}
}

func TestTrailingZeros8(t *testing.T) {
	n := uint8(0b00010000)
	if bits.TrailingZeros8(n) != 4 {
		t.Error("TrailingZeros8 failed")
	}
}

func TestOnesCount(t *testing.T) {
	n := uint(0b101010)
	if bits.OnesCount(n) != 3 {
		t.Error("OnesCount failed")
	}
}

func TestRotateLeft8(t *testing.T) {
	n := uint8(0b00001111)
	if bits.RotateLeft8(n, 2) != 0b00111100 {
		t.Error("RotateLeft8 failed")
	}
}

func TestReverse8(t *testing.T) {
	n := uint8(0b00001111)
	if bits.Reverse8(n) != 0b11110000 {
		t.Error("Reverse8 failed")
	}
}

func TestReverseBytes16(t *testing.T) {
	n := uint16(0x1234)
	if bits.ReverseBytes16(n) != 0x3412 {
		t.Error("ReverseBytes16 failed")
	}
}

func TestLen8(t *testing.T) {
	n := uint8(0b00001111)
	if bits.Len8(n) != 4 {
		t.Error("Len8 failed")
	}
}
