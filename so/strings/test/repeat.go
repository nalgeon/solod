// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

// A repeatCase is a test case of Repeat.
type repeatCase struct {
	in    string
	count int
	want  string
}

var repeatCases = []repeatCase{
	{"", 0, ""},
	{"", 1, ""},
	{"", 2, ""},
	{"-", 0, ""},
	{"-", 1, "-"},
	{"-", 10, "----------"},
	{"abc ", 3, "abc abc abc "},
	{" ", 1, " "},
	{"--", 2, "----"},
	{"===", 2, "======"},
	{"000", 3, "000000000"},
	{"\t\t\t\t", 4, "\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t"},
	{"\x00", 3, "\x00\x00\x00"},
	{"\xff", 2, "\xff\xff"},
	{faces, 2, facesTwice},
}

func TestRepeat(t *testing.T) {
	alloc := t.Allocator()
	for i, tc := range repeatCases {
		got := strings.Repeat(alloc, tc.in, tc.count)
		if got != tc.want {
			t.Errorf("case %d: Repeat() = %d bytes, want %d", i, len(got), len(tc.want))
		}
		mem.FreeString(alloc, got)
	}
}

// The one byte strings with a table of repeated bytes in the package. A count
// above the table length must give the same result as a count below it.
var fastBytes = []string{" ", "-", "0", "=", "\t"}

// The counts of the fast path cases. The table of every fast byte is 64 bytes
// or 128 bytes long, so the counts cross both lengths.
var fastCounts = []int{0, 1, 2, 63, 64, 65, 127, 128, 129, 300}

func TestRepeatFastBytes(t *testing.T) {
	// A one byte string with a table in the package gives the same result as
	// the general path.
	alloc := t.Allocator()
	for _, s := range fastBytes {
		for _, count := range fastCounts {
			got := strings.Repeat(alloc, s, count)
			if len(got) != count {
				t.Errorf("Repeat(%s, %d) = %d bytes, want %d", s, count, len(got), count)
				mem.FreeString(alloc, got)
				return
			}
			for i := range count {
				if got[i] != s[0] {
					t.Errorf("Repeat(%s, %d) has the byte %d at %d", s, count, got[i], i)
					break
				}
			}
			mem.FreeString(alloc, got)
		}
	}
}

func TestRepeatChunks(t *testing.T) {
	// The result is longer than the chunk limit of 8192 bytes, so the copy
	// runs several times.
	alloc := t.Allocator()
	const count = 12 * 1024
	got := strings.Repeat(alloc, "ab", count)
	defer mem.FreeString(alloc, got)
	if len(got) != 2*count {
		t.Errorf("Repeat() = %d bytes, want %d", len(got), 2*count)
		return
	}
	for i := range len(got) {
		want := byte('a')
		if i%2 == 1 {
			want = 'b'
		}
		if got[i] != want {
			t.Errorf("Repeat() has the byte %d at %d, want %d", got[i], i, want)
			return
		}
	}
}

func TestRepeatLongInput(t *testing.T) {
	// An input longer than the chunk limit gets copied whole every time.
	alloc := t.Allocator()
	const n = 9000
	long := strings.Repeat(alloc, "z", n)
	defer mem.FreeString(alloc, long)
	got := strings.Repeat(alloc, long, 2)
	defer mem.FreeString(alloc, got)
	if len(got) != 2*n {
		t.Errorf("Repeat() = %d bytes, want %d", len(got), 2*n)
		return
	}
	for i := range len(got) {
		if got[i] != 'z' {
			t.Errorf("Repeat() has the byte %d at %d, want 122", got[i], i)
			return
		}
	}
}
