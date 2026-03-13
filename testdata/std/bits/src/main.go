package main

import "github.com/nalgeon/solod/so/math/bits"

func main() {
	{
		// Add32.
		n1 := uint32(0b0101)
		n2 := uint32(0b0011)
		d, carry := bits.Add32(n1, n2, 0)
		if d != 0b1000 || carry != 0 {
			panic("Add32 failed")
		}
	}
	{
		// Sub32.
		n1 := uint32(0b0101)
		n2 := uint32(0b0011)
		d, borrow := bits.Sub32(n1, n2, 0)
		if d != 0b0010 || borrow != 0 {
			panic("Sub32 failed")
		}
	}
	{
		// Mul32.
		n1 := uint32(0b0101)
		n2 := uint32(0b0011)
		dh, dl := bits.Mul32(n1, n2)
		if dh != 0 || dl != 0b1111 {
			panic("Mul32 failed")
		}
	}
	{
		// LeadingZeros8.
		n := uint8(0b00010000)
		if bits.LeadingZeros8(n) != 3 {
			panic("LeadingZeros8 failed")
		}
	}
	{
		// TrailingZeros8.
		n := uint8(0b00010000)
		if bits.TrailingZeros8(n) != 4 {
			panic("TrailingZeros8 failed")
		}
	}
	{
		// OnesCount.
		n := uint(0b101010)
		if bits.OnesCount(n) != 3 {
			panic("OnesCount failed")
		}
	}
	{
		// RotateLeft8.
		n := uint8(0b00001111)
		if bits.RotateLeft8(n, 2) != 0b00111100 {
			panic("RotateLeft8 failed")
		}
	}
	{
		// Reverse8.
		n := uint8(0b00001111)
		if bits.Reverse8(n) != 0b11110000 {
			panic("Reverse8 failed")
		}
	}
	{
		// ReverseBytes16.
		n := uint16(0x1234)
		if bits.ReverseBytes16(n) != 0x3412 {
			panic("ReverseBytes16 failed")
		}
	}
	{
		// Len8.
		n := uint8(0b00001111)
		if bits.Len8(n) != 4 {
			panic("Len8 failed")
		}
	}
}
