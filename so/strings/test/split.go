// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings_test

import (
	"unsafe"

	"solod.dev/so/math"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

// A splitCase is a test case of Split, SplitN and SplitAfter. The wanted
// substrings are joined with partSep.
type splitCase struct {
	s, sep string
	n      int
	want   string
	parts  int
}

var splitCases = []splitCase{
	{"", "", -1, "", 0},
	{abcd, "", 2, "a\x02bcd", 2},
	{abcd, "", 4, "a\x02b\x02c\x02d", 4},
	{abcd, "", -1, "a\x02b\x02c\x02d", 4},
	{faces, "", -1, "☺\x02☻\x02☹", 3},
	{faces, "", 3, "☺\x02☻\x02☹", 3},
	{faces, "", 17, "☺\x02☻\x02☹", 3},
	{"☺\ufffd☹", "", -1, "☺\x02\ufffd\x02☹", 3},
	{abcd, "a", 0, "", 0},
	{abcd, "a", -1, "\x02bcd", 2},
	{abcd, "z", -1, "abcd", 1},
	{commas, ",", -1, "1\x022\x023\x024", 4},
	{dots, "...", -1, "1\x02.2\x02.3\x02.4", 4},
	{faces, "☹", -1, "☺☻\x02", 2},
	{faces, "~", -1, faces, 1},
	{"1 2 3 4", " ", 3, "1\x022\x023 4", 3},
	{"1 2", " ", 3, "1\x022", 2},
	{"", "T", math.MaxInt / 4, "", 1},
	{"\xff-\xff", "", -1, "\xff\x02-\x02\xff", 3},
	{"\xff-\xff", "-", -1, "\xff\x02\xff", 2},
}

// splitDiff returns the index of the first substring that differs from the
// wanted substring, or -1.
func splitDiff(got []string, want string, parts int) int {
	if len(got) != parts {
		return len(got)
	}
	for i, s := range got {
		if s != partAt(want, i) {
			return i
		}
	}
	return -1
}

func TestSplitN(t *testing.T) {
	alloc := t.Allocator()
	for i, tc := range splitCases {
		got := strings.SplitN(alloc, tc.s, tc.sep, tc.n)
		if d := splitDiff(got, tc.want, tc.parts); d >= 0 {
			t.Errorf("case %d: SplitN() gave %d substrings, the first wrong one is %d",
				i, len(got), d)
		}
		mem.FreeSlice(alloc, got)
	}
}

func TestSplitJoin(t *testing.T) {
	// Join of the substrings gives the first string back.
	alloc := t.Allocator()
	for i, tc := range splitCases {
		if tc.n == 0 {
			continue
		}
		parts := strings.SplitN(alloc, tc.s, tc.sep, tc.n)
		got := strings.Join(alloc, parts, tc.sep)
		if got != tc.s {
			t.Errorf("case %d: Join(SplitN()) = %d bytes, want %d", i, len(got), len(tc.s))
		}
		mem.FreeString(alloc, got)
		mem.FreeSlice(alloc, parts)
	}
}

func TestSplit(t *testing.T) {
	// Split gives the same result as SplitN with a negative count.
	alloc := t.Allocator()
	for i, tc := range splitCases {
		if tc.n >= 0 {
			continue
		}
		got := strings.Split(alloc, tc.s, tc.sep)
		if d := splitDiff(got, tc.want, tc.parts); d >= 0 {
			t.Errorf("case %d: Split() gave %d substrings, the first wrong one is %d",
				i, len(got), d)
		}
		mem.FreeSlice(alloc, got)
	}
}

func TestSplitViews(t *testing.T) {
	// The substrings are views of the first string, not copies.
	alloc := t.Allocator()
	const s = "a,b,c"
	got := strings.Split(alloc, s, ",")
	defer mem.FreeSlice(alloc, got)
	if len(got) != 3 {
		t.Errorf("Split() gave %d substrings, want 3", len(got))
		return
	}
	for i, part := range got {
		if unsafe.StringData(part) != unsafe.StringData(s[2*i:]) {
			t.Errorf("Split() copied the substring %d", i)
		}
	}
}

var splitAfterCases = []splitCase{
	{abcd, "a", -1, "a\x02bcd", 2},
	{abcd, "z", -1, "abcd", 1},
	{abcd, "", -1, "a\x02b\x02c\x02d", 4},
	{commas, ",", -1, "1,\x022,\x023,\x024", 4},
	{dots, "...", -1, "1...\x02.2...\x02.3...\x02.4", 4},
	{faces, "☹", -1, "☺☻☹\x02", 2},
	{faces, "~", -1, faces, 1},
	{faces, "", -1, "☺\x02☻\x02☹", 3},
}

func TestSplitAfter(t *testing.T) {
	alloc := t.Allocator()
	for i, tc := range splitAfterCases {
		got := strings.SplitAfter(alloc, tc.s, tc.sep)
		if d := splitDiff(got, tc.want, tc.parts); d >= 0 {
			t.Errorf("case %d: SplitAfter() gave %d substrings, the first wrong one is %d",
				i, len(got), d)
			mem.FreeSlice(alloc, got)
			continue
		}
		// The separator stays at the end of every substring, so Join with an
		// empty separator gives the first string back.
		joined := strings.Join(alloc, got, "")
		if joined != tc.s {
			t.Errorf("case %d: Join(SplitAfter()) = %d bytes, want %d",
				i, len(joined), len(tc.s))
		}
		mem.FreeString(alloc, joined)
		mem.FreeSlice(alloc, got)
	}
}

// A joinCase is a test case of Join. The elements are joined with partSep.
type joinCase struct {
	elems string
	count int
	sep   string
	want  string
}

var joinCases = []joinCase{
	{"", 0, ",", ""},
	{"a", 1, ",", "a"},
	{"a\x02b", 2, ",", "a,b"},
	{"a\x02b\x02c", 3, ",", "a,b,c"},
	{"a\x02b\x02c", 3, "", "abc"},
	{"\x02\x02", 3, "-", "--"},
	{"a\x02b", 2, "☺", "a☺b"},
}

func TestJoin(t *testing.T) {
	alloc := t.Allocator()
	var elems [8]string
	for i, tc := range joinCases {
		for k := range tc.count {
			elems[k] = partAt(tc.elems, k)
		}
		got := strings.Join(alloc, elems[:tc.count], tc.sep)
		if got != tc.want {
			t.Errorf("case %d: Join() = %s, want %s", i, got, tc.want)
		}
		mem.FreeString(alloc, got)
	}
}

// A fieldsCase is a test case of Fields and FieldsFunc.
type fieldsCase struct {
	s     string
	want  string
	parts int
}

var fieldsCases = []fieldsCase{
	{"", "", 0},
	{" ", "", 0},
	{" \t ", "", 0},
	{"\u2000", "", 0},
	{"  abc  ", "abc", 1},
	{"1 2 3 4", "1\x022\x023\x024", 4},
	{"1  2  3  4", "1\x022\x023\x024", 4},
	{"1\t\t2\t\t3\t4", "1\x022\x023\x024", 4},
	{"1\u20002\u20013\u20024", "1\x022\x023\x024", 4},
	{"\u2000\u2001\u2002", "", 0},
	{"\n™\t™\n", "™\x02™", 2},
	{"\n\u20001™2\u2000 \u2001 ™", "1™2\x02™", 2},
	{"\n1\ufffd \ufffd2\u20003\ufffd4", "1\ufffd\x02\ufffd2\x023\ufffd4", 3},
	{"1\xff\u2000\xff2\xff \xff", "1\xff\x02\xff2\xff\x02\xff", 3},
	{faces, faces, 1},
}

func TestFields(t *testing.T) {
	alloc := t.Allocator()
	for i, tc := range fieldsCases {
		got := strings.Fields(alloc, tc.s)
		if d := splitDiff(got, tc.want, tc.parts); d >= 0 {
			t.Errorf("case %d: Fields() gave %d substrings, the first wrong one is %d",
				i, len(got), d)
		}
		mem.FreeSlice(alloc, got)
	}
}

func TestFieldsFuncSpace(t *testing.T) {
	// FieldsFunc with the space predicate works like Fields.
	alloc := t.Allocator()
	for i, tc := range fieldsCases {
		got := strings.FieldsFunc(alloc, tc.s, predicate(predSpace))
		if d := splitDiff(got, tc.want, tc.parts); d >= 0 {
			t.Errorf("case %d: FieldsFunc() gave %d substrings, the first wrong one is %d",
				i, len(got), d)
		}
		mem.FreeSlice(alloc, got)
	}
}

// isLetterX reports whether the code point is the letter X.
func isLetterX(r rune) bool {
	return r == 'X'
}

var fieldsFuncCases = []fieldsCase{
	{"", "", 0},
	{"XX", "", 0},
	{"XXhiXXX", "hi", 1},
	{"aXXbXXXcX", "a\x02b\x02c", 3},
	{"XaX", "a", 1},
}

func TestFieldsFunc(t *testing.T) {
	alloc := t.Allocator()
	for i, tc := range fieldsFuncCases {
		got := strings.FieldsFunc(alloc, tc.s, isLetterX)
		if d := splitDiff(got, tc.want, tc.parts); d >= 0 {
			t.Errorf("case %d: FieldsFunc() gave %d substrings, the first wrong one is %d",
				i, len(got), d)
		}
		mem.FreeSlice(alloc, got)
	}
}

func TestSplitSweep(t *testing.T) {
	// Join of Split gives the first word back, and the number of substrings
	// is one more than the number of separators.
	alloc := t.Allocator()
	var sbuf, sepbuf [maxWord]byte
	for _, sw := range allocSweeps {
		words := wordTotal(sw.alpha, sw.maxWord)
		seps := wordTotal(sw.alpha, sw.maxSep)
		for i := range words {
			s := wordAt(sbuf[:], sw.alpha, sw.maxWord, i)
			for j := range seps {
				sep := wordAt(sepbuf[:], sw.alpha, sw.maxSep, j)
				if len(sep) == 0 {
					continue
				}
				parts := strings.Split(alloc, s, sep)
				want := countBrute(s, sep) + 1
				if len(parts) != want {
					var d1, d2 [2 * maxWord]byte
					t.Errorf("Split(%s, %s) gave %d substrings, want %d",
						dump(d1[:], s), dump(d2[:], sep), len(parts), want)
					mem.FreeSlice(alloc, parts)
					return
				}
				joined := strings.Join(alloc, parts, sep)
				if joined != s {
					var d1, d2 [2 * maxWord]byte
					t.Errorf("Join(Split(%s, %s)) gave %d bytes, want %d",
						dump(d1[:], s), dump(d2[:], sep), len(joined), len(s))
					mem.FreeString(alloc, joined)
					mem.FreeSlice(alloc, parts)
					return
				}
				mem.FreeString(alloc, joined)
				mem.FreeSlice(alloc, parts)
			}
		}
	}
}
