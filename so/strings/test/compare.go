// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

// A compareCase is a test case of Compare.
type compareCase struct {
	a, b string
	want int
}

var compareCases = []compareCase{
	{"", "", 0},
	{"a", "", 1},
	{"", "a", -1},
	{"abc", "abc", 0},
	{"ab", "abc", -1},
	{"abc", "ab", 1},
	{"x", "ab", 1},
	{"ab", "x", -1},
	{"x", "a", 1},
	{"b", "x", -1},
	// The comparison reads several bytes at a time, so a longer word checks
	// the tail of the read.
	{"abcdefgh", "abcdefgh", 0},
	{"abcdefghi", "abcdefghi", 0},
	{"abcdefghi", "abcdefghj", -1},
	// A high byte is above an ASCII byte, because the comparison is unsigned.
	{"\xff", "a", 1},
	{"a", "\xff", -1},
	{"\x00", "", 1},
}

func TestCompare(t *testing.T) {
	// The second argument moves through a buffer, so the comparison runs at
	// every alignment.
	const shifts = 16
	var buf [shifts + 32]byte
	for i, tc := range compareCases {
		for offset := 0; offset <= shifts; offset++ {
			b := buf[offset : offset+len(tc.b)]
			copy(b, tc.b)
			if got := strings.Compare(tc.a, string(b)); got != tc.want {
				t.Errorf("Compare() case %d at offset %d = %d, want %d",
					i, offset, got, tc.want)
				return
			}
		}
	}
}

func TestCompareSame(t *testing.T) {
	// One string compares equal with itself, and above a shorter view.
	const s = "Hello Gophers!"
	if got := strings.Compare(s, s); got != 0 {
		t.Errorf("Compare(s, s) = %d, want 0", got)
	}
	if got := strings.Compare(s, s[:1]); got != +1 {
		t.Errorf("Compare(s, s[:1]) = %d, want +1", got)
	}
}

func TestCompareLengths(t *testing.T) {
	// Every length up to 128 gets the same bytes in both strings, then one
	// byte of b moves up and down.
	const n = 128
	var abuf, bbuf [n]byte
	for i := range n {
		// Data that repeats slowly, and holds no 0 or 255.
		abuf[i] = byte(1 + 31*i%254)
		bbuf[i] = abuf[i]
	}
	for size := 0; size <= n; size++ {
		a, b := string(abuf[:size]), string(bbuf[:size])
		if got := strings.Compare(a, b); got != 0 {
			t.Errorf("Compare of equal strings with length %d = %d, want 0", size, got)
			return
		}
		if size == 0 {
			continue
		}
		if got := strings.Compare(a[:size-1], b); got != -1 {
			t.Errorf("Compare of the shorter a with length %d = %d, want -1", size, got)
			return
		}
		if got := strings.Compare(a, b[:size-1]); got != +1 {
			t.Errorf("Compare of the shorter b with length %d = %d, want +1", size, got)
			return
		}
		for k := 0; k < size; k++ {
			bbuf[k] = abuf[k] - 1
			if got := strings.Compare(a, b); got != +1 {
				t.Errorf("Compare with a lower byte %d of %d = %d, want +1", k, size, got)
				return
			}
			bbuf[k] = abuf[k] + 1
			if got := strings.Compare(a, b); got != -1 {
				t.Errorf("Compare with a higher byte %d of %d = %d, want -1", k, size, got)
				return
			}
			bbuf[k] = abuf[k]
		}
	}
}

// The lengths of the long comparison cases. A length near a power of two
// checks the block size of the comparison.
var longLengths = []int{256, 512, 1024, 1333, 4095, 4096, 4097}

func TestCompareLong(t *testing.T) {
	// A long comparison differs at the first byte, the middle byte and the
	// last byte.
	alloc := t.Allocator()
	const max = 4097
	abuf := mem.AllocSlice[byte](alloc, max, max)
	defer mem.FreeSlice(alloc, abuf)
	bbuf := mem.AllocSlice[byte](alloc, max, max)
	defer mem.FreeSlice(alloc, bbuf)
	for i := range max {
		abuf[i] = byte(1 + 31*i%254)
		bbuf[i] = abuf[i]
	}
	for _, size := range longLengths {
		a, b := string(abuf[:size]), string(bbuf[:size])
		if got := strings.Compare(a, b); got != 0 {
			t.Errorf("Compare of equal strings with length %d = %d, want 0", size, got)
			return
		}
		if got := strings.Compare(a[:size-1], b); got != -1 {
			t.Errorf("Compare of the shorter a with length %d = %d, want -1", size, got)
			return
		}
		spots := []int{0, size / 2, size - 1}
		for _, k := range spots {
			bbuf[k] = abuf[k] - 1
			if got := strings.Compare(a, b); got != +1 {
				t.Errorf("Compare with a lower byte %d of %d = %d, want +1", k, size, got)
				return
			}
			bbuf[k] = abuf[k] + 1
			if got := strings.Compare(a, b); got != -1 {
				t.Errorf("Compare with a higher byte %d of %d = %d, want -1", k, size, got)
				return
			}
			bbuf[k] = abuf[k]
		}
	}
}
