// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings_test

import (
	"unsafe"

	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
	"solod.dev/so/unicode/utf8"
)

func TestClone(t *testing.T) {
	alloc := t.Allocator()
	long := strings.Repeat(alloc, "a", 42)
	defer mem.FreeString(alloc, long)

	cases := []string{"", "short", long, long[:0]}
	for _, in := range cases {
		got := strings.Clone(alloc, in)
		if got != in {
			t.Errorf("Clone(%s) = %s, want %s", in, got, in)
		}
		if len(in) != 0 && unsafe.StringData(got) == unsafe.StringData(in) {
			t.Errorf("Clone(%s) shares the memory of the input", in)
		}
		mem.FreeString(alloc, got)
	}
}

// A countCase is a test case of Count.
type countCase struct {
	s, sep string
	want   int
}

var countCases = []countCase{
	{"", "", 1},
	{"", "notempty", 0},
	{"notempty", "", 9},
	{"smaller", "not smaller", 0},
	{"12345678987654321", "6", 2},
	{"611161116", "6", 3},
	{"notequal", "NotEqual", 0},
	{"equal", "equal", 1},
	{"abc1231231123q", "123", 3},
	{"11111", "11", 2},
	{faces, "", 4},
	{faces, "☻", 1},
	{"\xff\xff", "\xff", 2},
}

func TestCount(t *testing.T) {
	for _, tc := range countCases {
		if got := strings.Count(tc.s, tc.sep); got != tc.want {
			var sbuf, sepbuf [64]byte
			t.Errorf("Count(%s, %s) = %d, want %d",
				dump(sbuf[:], tc.s), dump(sepbuf[:], tc.sep), got, tc.want)
		}
	}
}

func TestCountSweep(t *testing.T) {
	// Count agrees with the reference over every short word.
	var sbuf, sepbuf [maxWord]byte
	for _, sw := range sweeps {
		words := wordTotal(sw.alpha, sw.maxWord)
		seps := wordTotal(sw.alpha, sw.maxSep)
		for i := range words {
			s := wordAt(sbuf[:], sw.alpha, sw.maxWord, i)
			for j := range seps {
				sep := wordAt(sepbuf[:], sw.alpha, sw.maxSep, j)
				if len(sep) == 0 {
					continue
				}
				got := strings.Count(s, sep)
				want := countBrute(s, sep)
				if got != want {
					var d1, d2 [2 * maxWord]byte
					t.Errorf("Count(%s, %s) = %d, want %d",
						dump(d1[:], s), dump(d2[:], sep), got, want)
					return
				}
			}
		}
	}
}

func TestCountEmptySep(t *testing.T) {
	// An empty separator counts the code points and adds one.
	cases := []string{"", "a", "abc", faces, "\xff\xff\xff", "a\x80b"}
	for _, s := range cases {
		want := utf8.RuneCountInString(s) + 1
		if got := strings.Count(s, ""); got != want {
			var buf [32]byte
			t.Errorf("Count(%s, empty) = %d, want %d", dump(buf[:], s), got, want)
		}
	}
}

// A cutCase is a test case of Cut, CutPrefix and CutSuffix.
type cutCase struct {
	s, sep        string
	before, after string
	found         bool
}

var cutCases = []cutCase{
	{"abc", "b", "a", "c", true},
	{"abc", "a", "", "bc", true},
	{"abc", "c", "ab", "", true},
	{"abc", "abc", "", "", true},
	{"abc", "", "", "abc", true},
	{"abc", "d", "abc", "", false},
	{"", "d", "", "", false},
	{"", "", "", "", true},
}

func TestCut(t *testing.T) {
	for _, tc := range cutCases {
		before, after := strings.Cut(tc.s, tc.sep)
		if before != tc.before || after != tc.after {
			t.Errorf("Cut(%s, %s) = %s, %s, want %s, %s",
				tc.s, tc.sep, before, after, tc.before, tc.after)
		}
	}
}

func TestCutViews(t *testing.T) {
	// Cut gives views of the input, not copies.
	const s = "abcdef"
	before, after := strings.Cut(s, "cd")
	if unsafe.StringData(before) != unsafe.StringData(s) {
		t.Error("Cut() copied the text before the separator")
	}
	if unsafe.StringData(after) != unsafe.StringData(s[4:]) {
		t.Error("Cut() copied the text after the separator")
	}
}

var cutPrefixCases = []cutCase{
	{"abc", "a", "", "bc", true},
	{"abc", "abc", "", "", true},
	{"abc", "", "", "abc", true},
	{"abc", "d", "", "abc", false},
	{"", "d", "", "", false},
	{"", "", "", "", true},
}

func TestCutPrefix(t *testing.T) {
	for _, tc := range cutPrefixCases {
		after, found := strings.CutPrefix(tc.s, tc.sep)
		if after != tc.after || found != tc.found {
			t.Errorf("CutPrefix(%s, %s) = %s, %t, want %s, %t",
				tc.s, tc.sep, after, found, tc.after, tc.found)
		}
	}
}

var cutSuffixCases = []cutCase{
	{"abc", "bc", "a", "", true},
	{"abc", "abc", "", "", true},
	{"abc", "", "abc", "", true},
	{"abc", "d", "abc", "", false},
	{"", "d", "", "", false},
	{"", "", "", "", true},
}

func TestCutSuffix(t *testing.T) {
	for _, tc := range cutSuffixCases {
		before, found := strings.CutSuffix(tc.s, tc.sep)
		if before != tc.before || found != tc.found {
			t.Errorf("CutSuffix(%s, %s) = %s, %t, want %s, %t",
				tc.s, tc.sep, before, found, tc.before, tc.found)
		}
	}
}

// A prefixCase is a test case of HasPrefix and HasSuffix.
type prefixCase struct {
	s, fix     string
	wantPrefix bool
	wantSuffix bool
}

var prefixCases = []prefixCase{
	{"", "", true, true},
	{"", "a", false, false},
	{"a", "", true, true},
	{"abc", "abc", true, true},
	{"abc", "a", true, false},
	{"abc", "c", false, true},
	{"abc", "abcd", false, false},
	{faces, "☺", true, false},
	{faces, "☹", false, true},
	{"\xff\x00", "\xff", true, false},
	{"\xff\x00", "\x00", false, true},
}

func TestHasPrefixSuffix(t *testing.T) {
	for _, tc := range prefixCases {
		var sbuf, fbuf [32]byte
		if got := strings.HasPrefix(tc.s, tc.fix); got != tc.wantPrefix {
			t.Errorf("HasPrefix(%s, %s) = %t, want %t",
				dump(sbuf[:], tc.s), dump(fbuf[:], tc.fix), got, tc.wantPrefix)
		}
		if got := strings.HasSuffix(tc.s, tc.fix); got != tc.wantSuffix {
			t.Errorf("HasSuffix(%s, %s) = %t, want %t",
				dump(sbuf[:], tc.s), dump(fbuf[:], tc.fix), got, tc.wantSuffix)
		}
	}
}

// A replaceCase is a test case of Replace.
type replaceCase struct {
	in       string
	old, new string
	n        int
	want     string
}

var replaceCases = []replaceCase{
	{"hello", "l", "L", 0, "hello"},
	{"hello", "l", "L", -1, "heLLo"},
	{"hello", "x", "X", -1, "hello"},
	{"", "x", "X", -1, ""},
	{"radar", "r", "<r>", -1, "<r>ada<r>"},
	{"", "", "<>", -1, "<>"},
	{"banana", "a", "<>", -1, "b<>n<>n<>"},
	{"banana", "a", "<>", 1, "b<>nana"},
	{"banana", "a", "<>", 1000, "b<>n<>n<>"},
	{"banana", "an", "<>", -1, "b<><>a"},
	{"banana", "ana", "<>", -1, "b<>na"},
	{"banana", "", "<>", -1, "<>b<>a<>n<>a<>n<>a<>"},
	{"banana", "", "<>", 10, "<>b<>a<>n<>a<>n<>a<>"},
	{"banana", "", "<>", 6, "<>b<>a<>n<>a<>n<>a"},
	{"banana", "", "<>", 5, "<>b<>a<>n<>a<>na"},
	{"banana", "", "<>", 1, "<>banana"},
	{"banana", "a", "a", -1, "banana"},
	{"banana", "a", "a", 1, "banana"},
	{faces, "", "<>", -1, "<>☺<>☻<>☹<>"},
}

func TestReplace(t *testing.T) {
	alloc := t.Allocator()
	for _, tc := range replaceCases {
		got := strings.Replace(alloc, tc.in, tc.old, tc.new, tc.n)
		if got != tc.want {
			t.Errorf("Replace(%s, %s, %s, %d) = %s, want %s",
				tc.in, tc.old, tc.new, tc.n, got, tc.want)
		}
		mem.FreeString(alloc, got)
	}
}

func TestReplaceAll(t *testing.T) {
	// ReplaceAll works like Replace with the count -1.
	alloc := t.Allocator()
	for _, tc := range replaceCases {
		if tc.n != -1 {
			continue
		}
		got := strings.ReplaceAll(alloc, tc.in, tc.old, tc.new)
		if got != tc.want {
			t.Errorf("ReplaceAll(%s, %s, %s) = %s, want %s",
				tc.in, tc.old, tc.new, got, tc.want)
		}
		mem.FreeString(alloc, got)
	}
}

// A runesCase is a test case of the conversion between a string and a slice of
// code points.
type runesCase struct {
	in    string
	want  string // the code points as decimal numbers, separated by a space
	lossy bool   // the string holds an invalid byte
}

var runesCases = []runesCase{
	{"", "", false},
	{" ", "32", false},
	{"ABC", "65 66 67", false},
	{"abc", "97 98 99", false},
	{"日本語", "26085 26412 35486", false},
	{"ab\x80c", "97 98 65533 99", true},
	{"ab\xc0c", "97 98 65533 99", true},
}

func TestRunes(t *testing.T) {
	alloc := t.Allocator()
	for _, tc := range runesCases {
		rs := []rune(tc.in)
		b := strings.NewBuilder(alloc)
		for i, r := range rs {
			if i > 0 {
				b.WriteByte(' ')
			}
			writeInt(&b, int(r))
		}
		if got := b.String(); got != tc.want {
			t.Errorf("[]rune(%s) = %s, want %s", tc.in, got, tc.want)
		}
		b.Free()

		if tc.lossy {
			// The invalid bytes became RuneError, so the text cannot come back.
			continue
		}
		if got := string(rs); got != tc.in {
			t.Errorf("string([]rune(%s)) = %s, want %s", tc.in, got, tc.in)
		}
	}
}

// writeInt writes the decimal form of n to the builder.
func writeInt(b *strings.Builder, n int) {
	if n >= 10 {
		writeInt(b, n/10)
	}
	b.WriteByte(byte('0' + n%10))
}
