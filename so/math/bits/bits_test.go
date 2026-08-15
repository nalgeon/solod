// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bits

import (
	stdbits "math/bits"
	"runtime"
	"testing"
)

const (
	divZeroError  = "runtime error: integer divide by zero"
	overflowError = "runtime error: integer overflow"
)

func TestDivPanicOverflow(t *testing.T) {
	// Expect a panic
	defer func() {
		if err := recover(); err == nil {
			t.Error("Div should have panicked when y<=hi")
		} else if err.(string) != overflowError {
			t.Errorf("Div expected panic: %q, got: %q ", overflowError, err.(string))
		}
	}()
	q, r := Div(1, 0, 1)
	t.Errorf("undefined q, r = %v, %v calculated when Div should have panicked", q, r)
}

func TestDiv32PanicOverflow(t *testing.T) {
	// Expect a panic
	defer func() {
		if err := recover(); err == nil {
			t.Error("Div32 should have panicked when y<=hi")
		} else if err.(string) != overflowError {
			t.Errorf("Div32 expected panic: %q, got: %q ", overflowError, err.(string))
		}
	}()
	q, r := Div32(1, 0, 1)
	t.Errorf("undefined q, r = %v, %v calculated when Div32 should have panicked", q, r)
}

func TestDiv64PanicOverflow(t *testing.T) {
	// Expect a panic
	defer func() {
		if err := recover(); err == nil {
			t.Error("Div64 should have panicked when y<=hi")
		} else if err.(string) != overflowError {
			t.Errorf("Div64 expected panic: %q, got: %q ", overflowError, err.(string))
		}
	}()
	q, r := Div64(1, 0, 1)
	t.Errorf("undefined q, r = %v, %v calculated when Div64 should have panicked", q, r)
}

func TestDivPanicZero(t *testing.T) {
	// Expect a panic
	defer func() {
		if err := recover(); err == nil {
			t.Error("Div should have panicked when y==0")
		} else if err.(string) != divZeroError {
			t.Errorf("Div expected panic: %q, got: %q ", divZeroError, err.(string))
		}
	}()
	q, r := Div(1, 1, 0)
	t.Errorf("undefined q, r = %v, %v calculated when Div should have panicked", q, r)
}

func TestDiv32PanicZero(t *testing.T) {
	// Expect a panic
	defer func() {
		if err := recover(); err == nil {
			t.Error("Div32 should have panicked when y==0")
		} else if e, ok := err.(runtime.Error); !ok || e.Error() != divZeroError {
			t.Errorf("Div32 expected panic: %q, got: %q ", divZeroError, err)
		}
	}()
	q, r := Div32(1, 1, 0)
	t.Errorf("undefined q, r = %v, %v calculated when Div32 should have panicked", q, r)
}

func TestDiv64PanicZero(t *testing.T) {
	// Expect a panic
	defer func() {
		if err := recover(); err == nil {
			t.Error("Div64 should have panicked when y==0")
		} else if err.(string) != divZeroError {
			t.Errorf("Div64 expected panic: %q, got: %q ", divZeroError, err.(string))
		}
	}()
	q, r := Div64(1, 1, 0)
	t.Errorf("undefined q, r = %v, %v calculated when Div64 should have panicked", q, r)
}

func FuzzCount(f *testing.F) {
	// Compare the counting functions against the standard library.
	f.Add(uint64(0))
	f.Add(uint64(1))
	f.Add(^uint64(0))
	f.Add(uint64(0x0102040810204080))
	f.Fuzz(func(t *testing.T, x uint64) {
		x8, x16, x32 := uint8(x), uint16(x), uint32(x)

		checkInt(t, "LeadingZeros8", LeadingZeros8(x8), stdbits.LeadingZeros8(x8))
		checkInt(t, "LeadingZeros16", LeadingZeros16(x16), stdbits.LeadingZeros16(x16))
		checkInt(t, "LeadingZeros32", LeadingZeros32(x32), stdbits.LeadingZeros32(x32))
		checkInt(t, "LeadingZeros64", LeadingZeros64(x), stdbits.LeadingZeros64(x))
		checkInt(t, "LeadingZeros", LeadingZeros(uint(x)), stdbits.LeadingZeros(uint(x)))

		checkInt(t, "TrailingZeros8", TrailingZeros8(x8), stdbits.TrailingZeros8(x8))
		checkInt(t, "TrailingZeros16", TrailingZeros16(x16), stdbits.TrailingZeros16(x16))
		checkInt(t, "TrailingZeros32", TrailingZeros32(x32), stdbits.TrailingZeros32(x32))
		checkInt(t, "TrailingZeros64", TrailingZeros64(x), stdbits.TrailingZeros64(x))
		checkInt(t, "TrailingZeros", TrailingZeros(uint(x)), stdbits.TrailingZeros(uint(x)))

		checkInt(t, "OnesCount8", OnesCount8(x8), stdbits.OnesCount8(x8))
		checkInt(t, "OnesCount16", OnesCount16(x16), stdbits.OnesCount16(x16))
		checkInt(t, "OnesCount32", OnesCount32(x32), stdbits.OnesCount32(x32))
		checkInt(t, "OnesCount64", OnesCount64(x), stdbits.OnesCount64(x))
		checkInt(t, "OnesCount", OnesCount(uint(x)), stdbits.OnesCount(uint(x)))

		checkInt(t, "Len8", Len8(x8), stdbits.Len8(x8))
		checkInt(t, "Len16", Len16(x16), stdbits.Len16(x16))
		checkInt(t, "Len32", Len32(x32), stdbits.Len32(x32))
		checkInt(t, "Len64", Len64(x), stdbits.Len64(x))
		checkInt(t, "Len", Len(uint(x)), stdbits.Len(uint(x)))
	})
}

func FuzzShuffle(f *testing.F) {
	// Compare the bit and byte shuffling functions against the standard library.
	f.Add(uint64(0), 0)
	f.Add(^uint64(0), 1)
	f.Add(uint64(0x0102040810204080), 33)
	f.Fuzz(func(t *testing.T, x uint64, k int) {
		x8, x16, x32 := uint8(x), uint16(x), uint32(x)

		checkUint(t, "Reverse8", uint64(Reverse8(x8)), uint64(stdbits.Reverse8(x8)))
		checkUint(t, "Reverse16", uint64(Reverse16(x16)), uint64(stdbits.Reverse16(x16)))
		checkUint(t, "Reverse32", uint64(Reverse32(x32)), uint64(stdbits.Reverse32(x32)))
		checkUint(t, "Reverse64", Reverse64(x), stdbits.Reverse64(x))
		checkUint(t, "Reverse", uint64(Reverse(uint(x))), uint64(stdbits.Reverse(uint(x))))

		checkUint(t, "ReverseBytes16", uint64(ReverseBytes16(x16)), uint64(stdbits.ReverseBytes16(x16)))
		checkUint(t, "ReverseBytes32", uint64(ReverseBytes32(x32)), uint64(stdbits.ReverseBytes32(x32)))
		checkUint(t, "ReverseBytes64", ReverseBytes64(x), stdbits.ReverseBytes64(x))
		checkUint(t, "ReverseBytes", uint64(ReverseBytes(uint(x))), uint64(stdbits.ReverseBytes(uint(x))))

		checkUint(t, "RotateLeft8", uint64(RotateLeft8(x8, k)), uint64(stdbits.RotateLeft8(x8, k)))
		checkUint(t, "RotateLeft16", uint64(RotateLeft16(x16, k)), uint64(stdbits.RotateLeft16(x16, k)))
		checkUint(t, "RotateLeft32", uint64(RotateLeft32(x32, k)), uint64(stdbits.RotateLeft32(x32, k)))
		checkUint(t, "RotateLeft64", RotateLeft64(x, k), stdbits.RotateLeft64(x, k))
		checkUint(t, "RotateLeft", uint64(RotateLeft(uint(x), k)), uint64(stdbits.RotateLeft(uint(x), k)))
	})
}

func FuzzAddSub(f *testing.F) {
	// Compare the add and subtract functions against the standard library.
	f.Add(uint64(0), uint64(0), uint64(0))
	f.Add(^uint64(0), ^uint64(0), uint64(1))
	f.Add(uint64(0x0102040810204080), uint64(0x8040201008040201), uint64(0))
	f.Fuzz(func(t *testing.T, x, y, c uint64) {
		c &= 1
		x32, y32, c32 := uint32(x), uint32(y), uint32(c)

		a1, a2 := Add32(x32, y32, c32)
		b1, b2 := stdbits.Add32(x32, y32, c32)
		checkPair(t, "Add32", uint64(a1), uint64(a2), uint64(b1), uint64(b2))

		a1, a2 = Sub32(x32, y32, c32)
		b1, b2 = stdbits.Sub32(x32, y32, c32)
		checkPair(t, "Sub32", uint64(a1), uint64(a2), uint64(b1), uint64(b2))

		g1, g2 := Add64(x, y, c)
		w1, w2 := stdbits.Add64(x, y, c)
		checkPair(t, "Add64", g1, g2, w1, w2)

		g1, g2 = Sub64(x, y, c)
		w1, w2 = stdbits.Sub64(x, y, c)
		checkPair(t, "Sub64", g1, g2, w1, w2)

		u1, u2 := Add(uint(x), uint(y), uint(c))
		v1, v2 := stdbits.Add(uint(x), uint(y), uint(c))
		checkPair(t, "Add", uint64(u1), uint64(u2), uint64(v1), uint64(v2))

		u1, u2 = Sub(uint(x), uint(y), uint(c))
		v1, v2 = stdbits.Sub(uint(x), uint(y), uint(c))
		checkPair(t, "Sub", uint64(u1), uint64(u2), uint64(v1), uint64(v2))
	})
}

func FuzzMulDiv(f *testing.F) {
	// Compare the multiply and divide functions against the standard library.
	// Div and Rem panic on the inputs that the standard library rejects,
	// so the fuzzer skips those inputs.
	f.Add(uint64(0), uint64(0), uint64(1))
	f.Add(^uint64(0), ^uint64(0), ^uint64(0))
	f.Add(uint64(0x0102040810204080), uint64(0x8040201008040201), uint64(111111))
	f.Fuzz(func(t *testing.T, hi, lo, y uint64) {
		hi32, lo32, y32 := uint32(hi), uint32(lo), uint32(y)

		a1, a2 := Mul32(hi32, lo32)
		b1, b2 := stdbits.Mul32(hi32, lo32)
		checkPair(t, "Mul32", uint64(a1), uint64(a2), uint64(b1), uint64(b2))

		g1, g2 := Mul64(hi, lo)
		w1, w2 := stdbits.Mul64(hi, lo)
		checkPair(t, "Mul64", g1, g2, w1, w2)

		u1, u2 := Mul(uint(hi), uint(lo))
		v1, v2 := stdbits.Mul(uint(hi), uint(lo))
		checkPair(t, "Mul", uint64(u1), uint64(u2), uint64(v1), uint64(v2))

		if y32 != 0 {
			checkUint(t, "Rem32", uint64(Rem32(hi32, lo32, y32)), uint64(stdbits.Rem32(hi32, lo32, y32)))
			if hi32 < y32 {
				a1, a2 = Div32(hi32, lo32, y32)
				b1, b2 = stdbits.Div32(hi32, lo32, y32)
				checkPair(t, "Div32", uint64(a1), uint64(a2), uint64(b1), uint64(b2))
			}
		}

		if y == 0 {
			return
		}
		checkUint(t, "Rem64", Rem64(hi, lo, y), stdbits.Rem64(hi, lo, y))
		checkUint(t, "Rem", uint64(Rem(uint(hi), uint(lo), uint(y))),
			uint64(stdbits.Rem(uint(hi), uint(lo), uint(y))))
		if hi >= y {
			return
		}
		g1, g2 = Div64(hi, lo, y)
		w1, w2 = stdbits.Div64(hi, lo, y)
		checkPair(t, "Div64", g1, g2, w1, w2)

		u1, u2 = Div(uint(hi), uint(lo), uint(y))
		v1, v2 = stdbits.Div(uint(hi), uint(lo), uint(y))
		checkPair(t, "Div", uint64(u1), uint64(u2), uint64(v1), uint64(v2))
	})
}

// checkInt checks one int result against the standard library.
func checkInt(t *testing.T, name string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %d, want %d", name, got, want)
	}
}

// checkUint checks one unsigned result against the standard library. The
// caller widens a narrow result to uint64.
func checkUint(t *testing.T, name string, got, want uint64) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %#x, want %#x", name, got, want)
	}
}

// checkPair checks a result pair against the standard library. The caller
// widens a narrow result to uint64.
func checkPair(t *testing.T, name string, got1, got2, want1, want2 uint64) {
	t.Helper()
	if got1 != want1 || got2 != want2 {
		t.Errorf("%s = %#x, %#x, want %#x, %#x", name, got1, got2, want1, want2)
	}
}
