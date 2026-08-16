// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings_test

import (
	"unsafe"

	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

// The trim functions, by number. A table holds the number instead of the
// function value.
const (
	kindTrim = iota
	kindTrimLeft
	kindTrimRight
	kindTrimPrefix
	kindTrimSuffix
)

// trimName returns the name of the trim function with the number.
func trimName(kind int) string {
	switch kind {
	case kindTrim:
		return "Trim"
	case kindTrimLeft:
		return "TrimLeft"
	case kindTrimRight:
		return "TrimRight"
	case kindTrimPrefix:
		return "TrimPrefix"
	}
	return "TrimSuffix"
}

// applyTrim calls the trim function with the number. Trim, TrimLeft and
// TrimRight take a cutset of code points. TrimPrefix and TrimSuffix take a
// string.
func applyTrim(kind int, s, arg string) string {
	switch kind {
	case kindTrim:
		return strings.Trim(s, arg)
	case kindTrimLeft:
		return strings.TrimLeft(s, arg)
	case kindTrimRight:
		return strings.TrimRight(s, arg)
	case kindTrimPrefix:
		return strings.TrimPrefix(s, arg)
	}
	return strings.TrimSuffix(s, arg)
}

// A trimCase is a test case of a trim function.
type trimCase struct {
	kind int
	in   string
	arg  string
	want string
}

var trimCases = []trimCase{
	{kindTrim, "abba", "a", "bb"},
	{kindTrim, "abba", "ab", ""},
	{kindTrimLeft, "abba", "ab", ""},
	{kindTrimRight, "abba", "ab", ""},
	{kindTrimLeft, "abba", "a", "bba"},
	{kindTrimLeft, "abba", "b", "abba"},
	{kindTrimRight, "abba", "a", "abb"},
	{kindTrimRight, "abba", "b", "abba"},
	{kindTrim, "<tag>", "<>", "tag"},
	{kindTrim, "* listitem", " *", "listitem"},
	{kindTrim, "\"quote\"", "\"", "quote"},
	// A cutset above ASCII takes the code point path.
	{kindTrim, "ⱯⱯɐɐⱯⱯ", "Ɐ", "ɐɐ"},
	{kindTrim, "\x80test\xff", "\xff", "test"},
	{kindTrim, " Ġ ", " ", "Ġ"},
	{kindTrim, " Ġİ0", "0 ", "Ġİ"},
	// A cutset that holds a part of a code point removes no byte of it.
	{kindTrimRight, "☺\xc0", "☺", "☺\xc0"},
	// An empty input or an empty cutset changes nothing.
	{kindTrim, "abba", "", "abba"},
	{kindTrim, "", "123", ""},
	{kindTrim, "", "", ""},
	{kindTrimLeft, "abba", "", "abba"},
	{kindTrimLeft, "", "123", ""},
	{kindTrimLeft, "", "", ""},
	{kindTrimRight, "abba", "", "abba"},
	{kindTrimRight, "", "123", ""},
	{kindTrimRight, "", "", ""},
	{kindTrimPrefix, "aabb", "a", "abb"},
	{kindTrimPrefix, "aabb", "b", "aabb"},
	{kindTrimPrefix, "aabb", "", "aabb"},
	{kindTrimSuffix, "aabb", "a", "aabb"},
	{kindTrimSuffix, "aabb", "b", "aab"},
	{kindTrimSuffix, "aabb", "", "aabb"},
}

func TestTrim(t *testing.T) {
	for i, tc := range trimCases {
		got := applyTrim(tc.kind, tc.in, tc.arg)
		if got != tc.want {
			var d1, d2 [64]byte
			t.Errorf("%s() case %d = %s, want %s",
				trimName(tc.kind), i, dump(d1[:], got), dump(d2[:], tc.want))
		}
	}
}

// The number of leading bytes that every trim function removes from
// "xxabcxx" with the cutset "x".
var trimStarts = []int{2, 2, 0, 1, 0}

func TestTrimViews(t *testing.T) {
	// A trim gives a view of the input, not a copy.
	const s = "xxabcxx"
	for kind := kindTrim; kind <= kindTrimSuffix; kind++ {
		got := applyTrim(kind, s, "x")
		start := trimStarts[kind]
		if unsafe.StringData(got) != unsafe.StringData(s[start:]) {
			t.Errorf("%s() copied the bytes", trimName(kind))
		}
	}
}

// A trimSpaceCase is a test case of TrimSpace.
type trimSpaceCase struct {
	in   string
	want string
}

var trimSpaceCases = []trimSpaceCase{
	{"", ""},
	{"abc", "abc"},
	{spaceAbc, "abc"},
	{" ", ""},
	{" \t\r\n \t\t\r\r\n\n ", ""},
	{" \t\r\n x\t\t\r\r\n\n ", "x"},
	{" \u2000\t\r\n x\t\t\r\r\ny\n \u3000", "x\t\t\r\r\ny"},
	{"1 \t\r\n2", "1 \t\r\n2"},
	// An invalid byte is not a space, and it stays in the result.
	{" x\x80", "x\x80"},
	{" x\xc0", "x\xc0"},
	{"x \xc0\xc0 ", "x \xc0\xc0"},
	{"x \xc0", "x \xc0"},
	{"x \xc0 ", "x \xc0"},
	{"x ☺\xc0\xc0 ", "x ☺\xc0\xc0"},
	{"x ☺ ", "x ☺"},
}

func TestTrimSpace(t *testing.T) {
	for i, tc := range trimSpaceCases {
		got := strings.TrimSpace(tc.in)
		if got != tc.want {
			var d1, d2 [128]byte
			t.Errorf("TrimSpace() case %d = %s, want %s",
				i, dump(d1[:], got), dump(d2[:], tc.want))
		}
	}
}

// A trimFuncCase is a test case of TrimFunc.
type trimFuncCase struct {
	pred int
	in   string
	want string
}

var trimFuncCases = []trimFuncCase{
	{predSpace, spaceHello, "hello"},
	{predDigit, "๐๒12hello34๐๑", "hello"},
	{predUpper, "ⱯⱯⱯⱯABCDhelloEFⱯⱯGHⱯⱯ", "hello"},
	{predValidRune, "ab\xc0a\xc0cd", "\xc0a\xc0"},
	{predNotSpace, helloSpace, space},
	{predNotDigit, "hello๐๒1234๐๑helo", "๐๒1234๐๑"},
	{predNotValidRune, "\xc0a\xc0", "a"},
	{predSpace, "", ""},
	{predSpace, " ", ""},
}

func TestTrimFunc(t *testing.T) {
	for i, tc := range trimFuncCases {
		got := strings.TrimFunc(tc.in, predicate(tc.pred))
		if got != tc.want {
			var d1, d2 [128]byte
			t.Errorf("TrimFunc(%s) case %d = %s, want %s",
				predName(tc.pred), i, dump(d1[:], got), dump(d2[:], tc.want))
		}
	}
}
