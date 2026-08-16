// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rand

import (
	stdrand "math/rand/v2"
	"testing"
)

func TestNPanic(t *testing.T) {
	p := NewPCG(1, 2)
	r := New(&p)
	tests := []struct {
		name string
		want string
		fn   func()
	}{
		{"Int64N(0)", "invalid argument to Int64N", func() { r.Int64N(0) }},
		{"Int64N(-1)", "invalid argument to Int64N", func() { r.Int64N(-1) }},
		{"Int32N(0)", "invalid argument to Int32N", func() { r.Int32N(0) }},
		{"Int32N(-1)", "invalid argument to Int32N", func() { r.Int32N(-1) }},
		{"IntN(0)", "invalid argument to IntN", func() { r.IntN(0) }},
		{"IntN(-1)", "invalid argument to IntN", func() { r.IntN(-1) }},
		{"Uint64N(0)", "invalid argument to Uint64N", func() { r.Uint64N(0) }},
		{"Uint32N(0)", "invalid argument to Uint32N", func() { r.Uint32N(0) }},
		{"UintN(0)", "invalid argument to UintN", func() { r.UintN(0) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				switch err := recover().(type) {
				case nil:
					t.Errorf("%s did not panic", test.name)
				case string:
					if err != test.want {
						t.Errorf("%s panicked with %q, want %q", test.name, err, test.want)
					}
				default:
					t.Errorf("%s panicked with %v, want %q", test.name, err, test.want)
				}
			}()
			test.fn()
		})
	}
}

func FuzzUint32n(f *testing.F) {
	// A freestanding target has a 32-bit word, so uint64n takes the uint32n
	// path for a bound that fits a uint32. Both paths must give the same
	// result for the same state of the source.
	f.Add(uint64(1), uint64(2), uint32(10))
	f.Add(uint64(1), uint64(2), uint32(1))
	f.Add(uint64(0), uint64(0), uint32(1)<<31)
	f.Add(^uint64(0), ^uint64(0), ^uint32(0))
	f.Fuzz(func(t *testing.T, seed1, seed2 uint64, n uint32) {
		if n == 0 {
			return
		}
		p1 := NewPCG(seed1, seed2)
		r1 := New(&p1)
		p2 := NewPCG(seed1, seed2)
		r2 := New(&p2)
		for i := range 10 {
			got, want := uint64(r1.uint32n(n)), r2.uint64n(uint64(n))
			if got != want {
				t.Fatalf("uint32n(%d) #%d = %d, want %d", n, i, got, want)
			}
		}
	})
}

func FuzzRand(f *testing.F) {
	// Compare the methods against the standard library.
	f.Add(uint64(1), uint64(2), uint64(10))
	f.Add(uint64(0), uint64(0), uint64(1))
	f.Add(^uint64(0), ^uint64(0), ^uint64(0))
	f.Add(uint64(0x123456789abcdef0), uint64(0xfedcba9876543210), uint64(1)<<40)
	f.Fuzz(func(t *testing.T, seed1, seed2, n uint64) {
		p := NewPCG(seed1, seed2)
		r := New(&p)
		stdP := stdrand.NewPCG(seed1, seed2)
		std := stdrand.New(stdP)

		checkUint(t, "Uint64", r.Uint64(), std.Uint64())
		checkUint(t, "Int64", uint64(r.Int64()), uint64(std.Int64()))
		checkUint(t, "Uint32", uint64(r.Uint32()), uint64(std.Uint32()))
		checkUint(t, "Int32", uint64(r.Int32()), uint64(std.Int32()))
		checkUint(t, "Uint", uint64(r.Uint()), uint64(std.Uint()))
		checkUint(t, "Int", uint64(r.Int()), uint64(std.Int()))
		checkFloat(t, "Float64", r.Float64(), std.Float64())
		checkFloat(t, "Float32", float64(r.Float32()), float64(std.Float32()))

		// A bound must be positive, so zero becomes one.
		n64 := max(n, 1)
		n32 := max(uint32(n), 1)
		i64 := max(int64(n&^(1<<63)), 1)
		i32 := max(int32(n32&^(1<<31)), 1)

		checkUint(t, "Uint64N", r.Uint64N(n64), std.Uint64N(n64))
		checkUint(t, "Uint32N", uint64(r.Uint32N(n32)), uint64(std.Uint32N(n32)))
		checkUint(t, "UintN", uint64(r.UintN(uint(n64))), uint64(std.UintN(uint(n64))))
		checkUint(t, "Int64N", uint64(r.Int64N(i64)), uint64(std.Int64N(i64)))
		checkUint(t, "Int32N", uint64(r.Int32N(i32)), uint64(std.Int32N(i32)))
		checkUint(t, "IntN", uint64(r.IntN(int(i64))), uint64(std.IntN(int(i64))))
	})
}

// checkUint checks one integer result against the standard library.
func checkUint(t *testing.T, name string, got, want uint64) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %#x, want %#x", name, got, want)
	}
}

// checkFloat checks one float result against the standard library.
func checkFloat(t *testing.T, name string, got, want float64) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}
