// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
	"solod.dev/so/unicode"
	"solod.dev/so/unicode/utf8"
)

// A caseCase is a test case of ToUpper and ToLower.
type caseCase struct {
	in, want string
}

var upperCases = []caseCase{
	{"", ""},
	{"ONLYUPPER", "ONLYUPPER"},
	{"abc", "ABC"},
	{"AbC123", "ABC123"},
	{"azAZ09_", "AZAZ09_"},
	{"longStrinGwitHmixofsmaLLandcAps", "LONGSTRINGWITHMIXOFSMALLANDCAPS"},
	{"longɐstringɐwithɐnonasciiⱯchars", "LONGⱯSTRINGⱯWITHⱯNONASCIIⱯCHARS"},
	{"ɐɐɐɐɐ", "ⱯⱯⱯⱯⱯ"},                         // one byte more per code point
	{"a\u0080\U0010FFFF", "A\u0080\U0010FFFF"}, // RuneSelf and MaxRune
	{"\xff\xfe", "\ufffd\ufffd"},               // every invalid byte becomes RuneError
}

var lowerCases = []caseCase{
	{"", ""},
	{"abc", "abc"},
	{"AbC123", "abc123"},
	{"azAZ09_", "azaz09_"},
	{"longStrinGwitHmixofsmaLLandcAps", "longstringwithmixofsmallandcaps"},
	{"LONGⱯSTRINGⱯWITHⱯNONASCIIⱯCHARS", "longɐstringɐwithɐnonasciiɐchars"},
	{"ⱭⱭⱭⱭⱭ", "ɑɑɑɑɑ"},                         // one byte less per code point
	{"A\u0080\U0010FFFF", "a\u0080\U0010FFFF"}, // RuneSelf and MaxRune
	{"\xff\xfe", "\ufffd\ufffd"},               // every invalid byte becomes RuneError
}

func TestToUpper(t *testing.T) {
	alloc := t.Allocator()
	for i, tc := range upperCases {
		got := strings.ToUpper(alloc, tc.in)
		if got != tc.want {
			t.Errorf("case %d: ToUpper(%s) = %s, want %s", i, tc.in, got, tc.want)
		}
		mem.FreeString(alloc, got)
	}
}

func TestToLower(t *testing.T) {
	alloc := t.Allocator()
	for i, tc := range lowerCases {
		got := strings.ToLower(alloc, tc.in)
		if got != tc.want {
			t.Errorf("case %d: ToLower(%s) = %s, want %s", i, tc.in, got, tc.want)
		}
		mem.FreeString(alloc, got)
	}
}

// toMaxRune gives the largest code point for every code point.
func toMaxRune(r rune) rune {
	_ = r
	return unicode.MaxRune
}

// toLetterA gives the letter a for every code point.
func toLetterA(r rune) rune {
	_ = r
	return 'a'
}

// rot13 moves a Latin letter 13 places along the alphabet. Two calls of rot13
// give the first code point back.
func rot13(r rune) rune {
	const step = 13
	if r >= 'a' && r <= 'z' {
		return (r-'a'+step)%26 + 'a'
	}
	if r >= 'A' && r <= 'Z' {
		return (r-'A'+step)%26 + 'A'
	}
	return r
}

// dropNotLatin drops every code point outside the Latin script.
func dropNotLatin(r rune) rune {
	if unicode.Is(unicode.Latin, r) {
		return r
	}
	return -1
}

// replaceNotLatin gives RuneError for every code point outside the Latin
// script.
func replaceNotLatin(r rune) rune {
	if unicode.Is(unicode.Latin, r) {
		return r
	}
	return utf8.RuneError
}

// swapEdges exchanges the smallest code point of two bytes and the largest
// code point. The two code points have the shortest and the longest encoding.
func swapEdges(r rune) rune {
	if r == utf8.RuneSelf {
		return unicode.MaxRune
	}
	if r == unicode.MaxRune {
		return utf8.RuneSelf
	}
	return r
}

// dropSpaces drops every space code point.
func dropSpaces(r rune) rune {
	if unicode.IsSpace(r) {
		return -1
	}
	return r
}

func TestMapGrow(t *testing.T) {
	// The result is longer than the input, so Map allocates more space.
	alloc := t.Allocator()
	want := strings.Repeat(alloc, "\U0010FFFF", 10)
	defer mem.FreeString(alloc, want)
	got := strings.Map(alloc, toMaxRune, "aaaaaaaaaa")
	defer mem.FreeString(alloc, got)
	if got != want {
		t.Errorf("Map() with a longer result gave %d bytes, want %d", len(got), len(want))
	}
}

func TestMapShrink(t *testing.T) {
	// The result is shorter than the input.
	alloc := t.Allocator()
	src := strings.Repeat(alloc, "\U0010FFFF", 10)
	defer mem.FreeString(alloc, src)
	got := strings.Map(alloc, toLetterA, src)
	defer mem.FreeString(alloc, got)
	if got != "aaaaaaaaaa" {
		t.Errorf("Map() with a shorter result = %s, want aaaaaaaaaa", got)
	}
}

func TestMapRot13(t *testing.T) {
	alloc := t.Allocator()
	once := strings.Map(alloc, rot13, "a to zed")
	defer mem.FreeString(alloc, once)
	if once != "n gb mrq" {
		t.Errorf("Map(rot13) = %s, want n gb mrq", once)
	}
	twice := strings.Map(alloc, rot13, once)
	defer mem.FreeString(alloc, twice)
	if twice != "a to zed" {
		t.Errorf("Map(rot13) twice = %s, want a to zed", twice)
	}
}

func TestMapDrop(t *testing.T) {
	// A negative result drops the code point.
	alloc := t.Allocator()
	got := strings.Map(alloc, dropNotLatin, "Hello, 세계")
	defer mem.FreeString(alloc, got)
	if got != "Hello" {
		t.Errorf("Map(dropNotLatin) = %s, want Hello", got)
	}
}

func TestMapInvalid(t *testing.T) {
	// An invalid byte reaches the mapping function as RuneError.
	alloc := t.Allocator()
	got := strings.Map(alloc, replaceNotLatin, "Hello\xffWorld")
	defer mem.FreeString(alloc, got)
	if got != "Hello\ufffdWorld" {
		t.Errorf("Map(replaceNotLatin) = %s, want Hello\ufffdWorld", got)
	}
}

func TestMapEdgeRunes(t *testing.T) {
	// The mapping writes the shortest and the longest encoding.
	alloc := t.Allocator()
	const short = "\u0080\U0010FFFF"
	const long = "\U0010FFFF\u0080"
	got := strings.Map(alloc, swapEdges, short)
	if got != long {
		t.Errorf("Map(swapEdges) gave %d bytes, want %d", len(got), len(long))
	}
	mem.FreeString(alloc, got)
	got = strings.Map(alloc, swapEdges, long)
	if got != short {
		t.Errorf("Map(swapEdges) back gave %d bytes, want %d", len(got), len(short))
	}
	mem.FreeString(alloc, got)
}

func TestMapEverywhere(t *testing.T) {
	// The mapping runs at the front, in the middle and at the back.
	alloc := t.Allocator()
	got := strings.Map(alloc, dropSpaces, "   abc    123   ")
	defer mem.FreeString(alloc, got)
	if got != "abc123" {
		t.Errorf("Map(dropSpaces) = %s, want abc123", got)
	}
}

func TestMapEmpty(t *testing.T) {
	alloc := t.Allocator()
	got := strings.Map(alloc, rot13, "")
	defer mem.FreeString(alloc, got)
	if got != "" {
		t.Errorf("Map() of an empty string = %s, want an empty string", got)
	}
}

func TestCaseConsistency(t *testing.T) {
	// The case change keeps the number of code points, and a second change
	// gives the same string.
	const numRunes = 1000
	alloc := t.Allocator()
	b := strings.NewBuilder(alloc)
	defer b.Free()
	for i := range numRunes {
		b.WriteRune(rune(i))
	}
	s := b.String()

	upper := strings.ToUpper(alloc, s)
	defer mem.FreeString(alloc, upper)
	lower := strings.ToLower(alloc, s)
	defer mem.FreeString(alloc, lower)

	if n := utf8.RuneCountInString(upper); n != numRunes {
		t.Errorf("RuneCountInString(upper) = %d, want %d", n, numRunes)
	}
	if n := utf8.RuneCountInString(lower); n != numRunes {
		t.Errorf("RuneCountInString(lower) = %d, want %d", n, numRunes)
	}

	again := strings.ToUpper(alloc, upper)
	if again != upper {
		t.Errorf("ToUpper() of an upper case string differs at %d", firstDiff(again, upper))
	}
	mem.FreeString(alloc, again)

	again = strings.ToLower(alloc, lower)
	if again != lower {
		t.Errorf("ToLower() of a lower case string differs at %d", firstDiff(again, lower))
	}
	mem.FreeString(alloc, again)
}

// firstDiff returns the index of the first byte that differs, or the length of
// the shorter string.
func firstDiff(a, b string) int {
	n := min(len(b), len(a))
	for i := range n {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}
