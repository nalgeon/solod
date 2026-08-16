// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binary_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/math/rand"
	"solod.dev/so/testing"
)

// check converts one value with one order and compares Put, Uint and Append
// against the reference implementations.
func check(t *testing.T, be bool, value uint64, width int) {
	var got [8]byte
	var want [8]byte

	putRef(be, want[:width], value, width)
	putUint(be, got[:width], value, width)
	if !bytes.Equal(got[:width], want[:width]) {
		t.Errorf("%s.PutUint%d(%x) = %x, want %x",
			orderName(be), 8*width, value, got[:width], want[:width])
		return
	}

	wantValue := lowBytes(value, width)
	if v := uintAt(be, want[:width], width); v != wantValue {
		t.Errorf("%s.Uint%d(%x) = %x, want %x",
			orderName(be), 8*width, want[:width], v, wantValue)
		return
	}
	if v := uintRef(be, got[:width], width); v != wantValue {
		t.Errorf("%s: reference Uint%d(%x) = %x, want %x",
			orderName(be), 8*width, got[:width], v, wantValue)
		return
	}

	var buf [9]byte
	buf[0] = 0xa5
	res := appendUint(be, buf[:1], value, width)
	if len(res) != 1+width {
		t.Errorf("%s.AppendUint%d(%x): len = %d, want %d",
			orderName(be), 8*width, value, len(res), 1+width)
		return
	}
	if res[0] != 0xa5 {
		t.Errorf("%s.AppendUint%d(%x): prefix = %x, want a5",
			orderName(be), 8*width, value, res[0])
		return
	}
	if !bytes.Equal(res[1:], want[:width]) {
		t.Errorf("%s.AppendUint%d(%x) = %x, want %x",
			orderName(be), 8*width, value, res[1:], want[:width])
	}
}

func TestSweep16(t *testing.T) {
	// Convert every uint16 value.
	for _, be := range orders {
		for v := range 1 << 16 {
			check(t, be, uint64(v), 2)
			if t.Failed() {
				return
			}
		}
	}
}

func TestSweep32(t *testing.T) {
	// Convert every value of the low half and of the high half of a uint32.
	// Also convert the values around every power of two.
	for _, be := range orders {
		for v := range 1 << 16 {
			check(t, be, uint64(v), 4)
			check(t, be, uint64(v)<<16, 4)
			if t.Failed() {
				return
			}
		}
		for i := range 32 {
			p := uint64(1) << uint(i)
			check(t, be, p, 4)
			check(t, be, p-1, 4)
			check(t, be, ^p, 4)
			if t.Failed() {
				return
			}
		}
	}
}

func TestSweep64(t *testing.T) {
	// Convert the values around every power of two.
	// Also convert every byte value at every byte position of a uint64.
	for _, be := range orders {
		for i := range 64 {
			p := uint64(1) << uint(i)
			check(t, be, p, 8)
			check(t, be, p-1, 8)
			check(t, be, ^p, 8)
			if t.Failed() {
				return
			}
		}
		for i := range 8 {
			for b := range 256 {
				check(t, be, uint64(b)<<uint(8*i), 8)
				if t.Failed() {
					return
				}
			}
		}
	}
}

func TestSweepRandom(t *testing.T) {
	// Convert pseudo-random values of every width.
	const rounds = 50000
	src := rand.NewPCG(0x2545f4914f6cdd1d, 0x9e3779b97f4a7c15)
	for range rounds {
		v := src.Uint64()
		for _, be := range orders {
			for _, width := range widths {
				check(t, be, v, width)
				if t.Failed() {
					return
				}
			}
		}
	}
}

func TestSweepOffset(t *testing.T) {
	// Write and read a value at every offset of a buffer.
	src := rand.NewPCG(1, 2)
	buf := make([]byte, 24)
	var want [8]byte
	for range 2000 {
		v := src.Uint64()
		for _, be := range orders {
			for _, width := range widths {
				for offset := range 16 {
					putUint(be, buf[offset:], v, width)
					putRef(be, want[:width], v, width)
					if !bytes.Equal(buf[offset:offset+width], want[:width]) {
						t.Errorf("%s.PutUint%d(%x) at %d = %x, want %x", orderName(be),
							8*width, v, offset, buf[offset:offset+width], want[:width])
						return
					}
					if got := uintAt(be, buf[offset:], width); got != lowBytes(v, width) {
						t.Errorf("%s.Uint%d(%x) at %d = %x, want %x", orderName(be),
							8*width, v, offset, got, lowBytes(v, width))
						return
					}
				}
			}
		}
	}
}
