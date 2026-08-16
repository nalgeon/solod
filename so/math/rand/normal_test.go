// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rand

import (
	stdrand "math/rand/v2"
	"testing"
)

func FuzzNormFloat64(f *testing.F) {
	// Compare the generator against the standard library.
	f.Add(uint64(1), uint64(2))
	f.Add(uint64(0), uint64(0))
	f.Add(^uint64(0), ^uint64(0))
	f.Add(uint64(0x123456789abcdef0), uint64(0xfedcba9876543210))
	f.Fuzz(func(t *testing.T, seed1, seed2 uint64) {
		p := NewPCG(seed1, seed2)
		r := New(&p)
		stdP := stdrand.NewPCG(seed1, seed2)
		std := stdrand.New(stdP)
		for i := range 100 {
			got, want := r.NormFloat64(), std.NormFloat64()
			if got != want {
				t.Fatalf("NormFloat64() #%d = %v, want %v", i, got, want)
			}
		}
	})
}
