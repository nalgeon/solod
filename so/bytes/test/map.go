package bytes_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
	"solod.dev/so/unicode"
	"solod.dev/so/unicode/utf8"
)

// A caseCase is a test case of ToLower and ToUpper.
type caseCase struct {
	in   string
	want string
}

var lowerCases = []caseCase{
	{"", ""},
	{"abc", "abc"},
	{"AbC123", "abc123"},
	{"azAZ09_", "azaz09_"},
	{"longStrinGwitHmixofsmaLLandcAps", "longstringwithmixofsmallandcaps"},
	{"LONGⱯSTRINGⱯWITHⱯNONASCIIⱯCHARS", "longɐstringɐwithɐnonasciiɐchars"},
	// The result loses one byte per code point.
	{"ⱭⱭⱭⱭⱭ", "ɑɑɑɑɑ"},
	// The first code point above ASCII and the highest code point.
	{"A\U0010FFFF", "a\U0010FFFF"},
}

func TestToLower(t *testing.T) {
	alloc := t.Allocator()
	for i, tc := range lowerCases {
		got := bytes.ToLower(alloc, []byte(tc.in))
		if string(got) != tc.want {
			t.Errorf("ToLower() case %d = %s, want %s", i, string(got), tc.want)
		}
		mem.FreeSlice(alloc, got)
	}
}

var upperCases = []caseCase{
	{"", ""},
	{"ONLYUPPER", "ONLYUPPER"},
	{"abc", "ABC"},
	{"AbC123", "ABC123"},
	{"azAZ09_", "AZAZ09_"},
	{"longStrinGwitHmixofsmaLLandcAps", "LONGSTRINGWITHMIXOFSMALLANDCAPS"},
	{"longɐstringɐwithɐnonasciiⱯchars", "LONGⱯSTRINGⱯWITHⱯNONASCIIⱯCHARS"},
	// The result gains one byte per code point.
	{"ɐɐɐɐɐ", "ⱯⱯⱯⱯⱯ"},
	// The first code point above ASCII and the highest code point.
	{"a\U0010FFFF", "A\U0010FFFF"},
}

func TestToUpper(t *testing.T) {
	alloc := t.Allocator()
	for i, tc := range upperCases {
		got := bytes.ToUpper(alloc, []byte(tc.in))
		if string(got) != tc.want {
			t.Errorf("ToUpper() case %d = %s, want %s", i, string(got), tc.want)
		}
		mem.FreeSlice(alloc, got)
	}
}

func TestToLowerCopy(t *testing.T) {
	// ToLower gives a copy of s, even when s has no upper case letter.
	var buf [3]byte
	s := buf[:]
	copy(s, "abc")
	alloc := t.Allocator()
	got := bytes.ToLower(alloc, s)
	defer mem.FreeSlice(alloc, got)
	if string(got) != "abc" {
		t.Errorf("ToLower(abc) = %s, want abc", string(got))
		return
	}
	got[0] = 'x'
	if string(s) != "abc" {
		t.Errorf("ToLower gives a view of s: %s", string(s))
	}
}

func TestToUpperCopy(t *testing.T) {
	// ToUpper gives a copy of s, even when s has no lower case letter.
	var buf [3]byte
	s := buf[:]
	copy(s, "ABC")
	alloc := t.Allocator()
	got := bytes.ToUpper(alloc, s)
	defer mem.FreeSlice(alloc, got)
	if string(got) != "ABC" {
		t.Errorf("ToUpper(ABC) = %s, want ABC", string(got))
		return
	}
	got[0] = 'x'
	if string(s) != "ABC" {
		t.Errorf("ToUpper gives a view of s: %s", string(s))
	}
}

func TestCaseEmpty(t *testing.T) {
	// An empty slice gives an empty result.
	alloc := t.Allocator()
	lower := bytes.ToLower(alloc, nil)
	if len(lower) != 0 {
		t.Errorf("ToLower(nil) has length %d, want 0", len(lower))
	}
	mem.FreeSlice(alloc, lower)
	upper := bytes.ToUpper(alloc, []byte{})
	if len(upper) != 0 {
		t.Errorf("ToUpper(empty) has length %d, want 0", len(upper))
	}
	mem.FreeSlice(alloc, upper)
}

// The mapping functions of the Map tests.

// toMaxRune maps every code point to the highest code point.
func toMaxRune(r rune) rune {
	_ = r
	return unicode.MaxRune
}

// toLetterA maps every code point to the letter a.
func toLetterA(r rune) rune {
	_ = r
	return 'a'
}

// rot13 moves a letter 13 places along the alphabet. rot13 is its own inverse.
func rot13(r rune) rune {
	const step = 13
	if r >= 'a' && r <= 'z' {
		return ((r - 'a' + step) % 26) + 'a'
	}
	if r >= 'A' && r <= 'Z' {
		return ((r - 'A' + step) % 26) + 'A'
	}
	return r
}

// dropNotUpper drops every code point that is not an upper case letter.
func dropNotUpper(r rune) rune {
	if unicode.Is(unicode.Upper, r) {
		return r
	}
	return -1
}

// toInvalidRune maps every code point to a value above the highest code point.
func toInvalidRune(r rune) rune {
	_ = r
	return utf8.MaxRune + 1
}

// tenRunes writes the code point ten times into buf and returns the result.
func tenRunes(buf []byte, r rune) []byte {
	n := 0
	for range 10 {
		n += utf8.EncodeRune(buf[n:], r)
	}
	return buf[:n]
}

func TestMapGrow(t *testing.T) {
	// The result of the mapping is longer than the input, so Map grows the
	// result more than once.
	var inBuf, wantBuf [10 * utf8.UTFMax]byte
	in := tenRunes(inBuf[:], 'a')
	want := tenRunes(wantBuf[:], unicode.MaxRune)
	alloc := t.Allocator()
	got := bytes.Map(alloc, toMaxRune, in)
	defer mem.FreeSlice(alloc, got)
	if !bytes.Equal(got, want) {
		var d1, d2 [2 * 10 * utf8.UTFMax]byte
		t.Errorf("Map(toMaxRune) = %s, want %s",
			dump(d1[:], got), dump(d2[:], want))
	}
}

func TestMapShrink(t *testing.T) {
	// The result of the mapping is shorter than the input.
	var inBuf, wantBuf [10 * utf8.UTFMax]byte
	in := tenRunes(inBuf[:], unicode.MaxRune)
	want := tenRunes(wantBuf[:], 'a')
	alloc := t.Allocator()
	got := bytes.Map(alloc, toLetterA, in)
	defer mem.FreeSlice(alloc, got)
	if !bytes.Equal(got, want) {
		var d1, d2 [2 * 10 * utf8.UTFMax]byte
		t.Errorf("Map(toLetterA) = %s, want %s",
			dump(d1[:], got), dump(d2[:], want))
	}
}

func TestMapRot13(t *testing.T) {
	alloc := t.Allocator()
	got := bytes.Map(alloc, rot13, []byte("a to zed"))
	defer mem.FreeSlice(alloc, got)
	if string(got) != "n gb mrq" {
		t.Errorf("Map(rot13) = %s, want n gb mrq", string(got))
		return
	}
	// rot13 twice gives the input back.
	back := bytes.Map(alloc, rot13, got)
	defer mem.FreeSlice(alloc, back)
	if string(back) != "a to zed" {
		t.Errorf("Map(rot13) twice = %s, want a to zed", string(back))
	}
}

func TestMapDrop(t *testing.T) {
	// A negative result of the mapping drops the code point.
	alloc := t.Allocator()
	got := bytes.Map(alloc, dropNotUpper, []byte("Hello, WORLD"))
	defer mem.FreeSlice(alloc, got)
	if string(got) != "HWORLD" {
		t.Errorf("Map(dropNotUpper) = %s, want HWORLD", string(got))
	}
}

func TestMapInvalidRune(t *testing.T) {
	// A result of the mapping above the highest code point becomes RuneError.
	alloc := t.Allocator()
	got := bytes.Map(alloc, toInvalidRune, []byte("x"))
	defer mem.FreeSlice(alloc, got)
	if string(got) != "�" {
		var d [16]byte
		t.Errorf("Map(toInvalidRune) = %s, want efbfbd", dump(d[:], got))
	}
}

func TestMapEmpty(t *testing.T) {
	alloc := t.Allocator()
	got := bytes.Map(alloc, rot13, nil)
	if len(got) != 0 {
		t.Errorf("Map(nil) has length %d, want 0", len(got))
	}
	mem.FreeSlice(alloc, got)
}

func TestMapInvalidInput(t *testing.T) {
	// An invalid byte of the input becomes RuneError, one per byte.
	alloc := t.Allocator()
	got := bytes.Map(alloc, rot13, []byte("a\xffb"))
	defer mem.FreeSlice(alloc, got)
	if string(got) != "n�o" {
		var d [16]byte
		t.Errorf("Map(rot13, a-ff-b) = %s, want 6eefbfbd6f", dump(d[:], got))
	}
}
