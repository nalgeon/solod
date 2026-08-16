// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rand

import (
	"bytes"
	stdrand "math/rand/v2"
	"testing"
)

func FuzzPCG(f *testing.F) {
	// Compare the generator against the standard library.
	f.Add(uint64(0), uint64(0))
	f.Add(uint64(1), uint64(2))
	f.Add(^uint64(0), ^uint64(0))
	f.Add(uint64(0x123456789abcdef0), uint64(0xfedcba9876543210))
	f.Fuzz(func(t *testing.T, seed1, seed2 uint64) {
		p := NewPCG(seed1, seed2)
		std := stdrand.NewPCG(seed1, seed2)
		for i := range 20 {
			got, want := p.Uint64(), std.Uint64()
			if got != want {
				t.Fatalf("PCG(%#x, %#x).Uint64() #%d = %#x, want %#x", seed1, seed2, i, got, want)
			}
		}
	})
}

func FuzzPCGMarshal(f *testing.F) {
	// Compare the encoding against the standard library.
	f.Add(uint64(0), uint64(0))
	f.Add(uint64(1), uint64(2))
	f.Add(uint64(0x123456789abcdef0), uint64(0xfedcba9876543210))
	f.Fuzz(func(t *testing.T, seed1, seed2 uint64) {
		p := NewPCG(seed1, seed2)
		got, err := p.MarshalBinary(make([]byte, 20))
		if err != nil {
			t.Fatalf("MarshalBinary(): %v", err)
		}
		std := stdrand.NewPCG(seed1, seed2)
		want, err := std.MarshalBinary()
		if err != nil {
			t.Fatalf("std MarshalBinary(): %v", err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("PCG(%#x, %#x).MarshalBinary() = %q, want %q", seed1, seed2, got, want)
		}
	})
}
