// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unicode_test

import (
	"testing"

	. "github.com/nalgeon/solod/so/unicode"
)

var upperTest = []rune{
	0x41,
	0xc0,
	0xd8,
	0x100,
	0x139,
	0x14a,
	0x178,
	0x181,
	0x376,
	0x3cf,
	0x13bd,
	0x1f2a,
	0x2102,
	0x2c00,
	0x2c10,
	0x2c20,
	0xa650,
	0xa722,
	0xff3a,
	0x10400,
	0x1d400,
	0x1d7ca,
}

var notupperTest = []rune{
	0x40,
	0x5b,
	0x61,
	0x185,
	0x1b0,
	0x377,
	0x387,
	0x2150,
	0xab7d,
	0xffff,
	0x10000,
}

var letterTest = []rune{
	0x41,
	0x61,
	0xaa,
	0xba,
	0xc8,
	0xdb,
	0xf9,
	0x2ec,
	0x535,
	0x620,
	0x6e6,
	0x93d,
	0xa15,
	0xb99,
	0xdc0,
	0xedd,
	0x1000,
	0x1200,
	0x1312,
	0x1401,
	0x2c00,
	0xa800,
	0xf900,
	0xfa30,
	0xffda,
	0xffdc,
	0x10000,
	0x10300,
	0x10400,
	0x20000,
	0x2f800,
	0x2fa1d,
}

var notletterTest = []rune{
	0x20,
	0x35,
	0x375,
	0x619,
	0x700,
	0x1885,
	0xfffe,
	0x1ffff,
	0x10ffff,
}

// Contains all the special cased Latin-1 chars.
var spaceTest = []rune{
	0x09,
	0x0a,
	0x0b,
	0x0c,
	0x0d,
	0x20,
	0x85,
	0xA0,
	0x2000,
	0x3000,
}

type caseT struct {
	cas     int
	in, out rune
}

var caseTest = []caseT{
	// errors
	{-1, '\n', 0xFFFD},
	{UpperCase, -1, -1},
	{UpperCase, 1 << 30, 1 << 30},

	// ASCII (special-cased so test carefully)
	{UpperCase, '\n', '\n'},
	{UpperCase, 'a', 'A'},
	{UpperCase, 'A', 'A'},
	{UpperCase, '7', '7'},
	{LowerCase, '\n', '\n'},
	{LowerCase, 'a', 'a'},
	{LowerCase, 'A', 'a'},
	{LowerCase, '7', '7'},
	{TitleCase, '\n', '\n'},
	{TitleCase, 'a', 'A'},
	{TitleCase, 'A', 'A'},
	{TitleCase, '7', '7'},

	// Latin-1: easy to read the tests!
	{UpperCase, 0x80, 0x80},
	{UpperCase, 'Å', 'Å'},
	{UpperCase, 'å', 'Å'},
	{LowerCase, 0x80, 0x80},
	{LowerCase, 'Å', 'å'},
	{LowerCase, 'å', 'å'},
	{TitleCase, 0x80, 0x80},
	{TitleCase, 'Å', 'Å'},
	{TitleCase, 'å', 'Å'},

	// 0131;LATIN SMALL LETTER DOTLESS I;Ll;0;L;;;;;N;;;0049;;0049
	{UpperCase, 0x0131, 'I'},
	{LowerCase, 0x0131, 0x0131},
	{TitleCase, 0x0131, 'I'},

	// 0133;LATIN SMALL LIGATURE IJ;Ll;0;L;<compat> 0069 006A;;;;N;LATIN SMALL LETTER I J;;0132;;0132
	{UpperCase, 0x0133, 0x0132},
	{LowerCase, 0x0133, 0x0133},
	{TitleCase, 0x0133, 0x0132},

	// 212A;KELVIN SIGN;Lu;0;L;004B;;;;N;DEGREES KELVIN;;;006B;
	{UpperCase, 0x212A, 0x212A},
	{LowerCase, 0x212A, 'k'},
	{TitleCase, 0x212A, 0x212A},

	// From an UpperLower sequence
	// A640;CYRILLIC CAPITAL LETTER ZEMLYA;Lu;0;L;;;;;N;;;;A641;
	{UpperCase, 0xA640, 0xA640},
	{LowerCase, 0xA640, 0xA641},
	{TitleCase, 0xA640, 0xA640},
	// A641;CYRILLIC SMALL LETTER ZEMLYA;Ll;0;L;;;;;N;;;A640;;A640
	{UpperCase, 0xA641, 0xA640},
	{LowerCase, 0xA641, 0xA641},
	{TitleCase, 0xA641, 0xA640},
	// A64E;CYRILLIC CAPITAL LETTER NEUTRAL YER;Lu;0;L;;;;;N;;;;A64F;
	{UpperCase, 0xA64E, 0xA64E},
	{LowerCase, 0xA64E, 0xA64F},
	{TitleCase, 0xA64E, 0xA64E},
	// A65F;CYRILLIC SMALL LETTER YN;Ll;0;L;;;;;N;;;A65E;;A65E
	{UpperCase, 0xA65F, 0xA65E},
	{LowerCase, 0xA65F, 0xA65F},
	{TitleCase, 0xA65F, 0xA65E},

	// From another UpperLower sequence
	// 0139;LATIN CAPITAL LETTER L WITH ACUTE;Lu;0;L;004C 0301;;;;N;LATIN CAPITAL LETTER L ACUTE;;;013A;
	{UpperCase, 0x0139, 0x0139},
	{LowerCase, 0x0139, 0x013A},
	{TitleCase, 0x0139, 0x0139},
	// 013F;LATIN CAPITAL LETTER L WITH MIDDLE DOT;Lu;0;L;<compat> 004C 00B7;;;;N;;;;0140;
	{UpperCase, 0x013f, 0x013f},
	{LowerCase, 0x013f, 0x0140},
	{TitleCase, 0x013f, 0x013f},
	// 0148;LATIN SMALL LETTER N WITH CARON;Ll;0;L;006E 030C;;;;N;LATIN SMALL LETTER N HACEK;;0147;;0147
	{UpperCase, 0x0148, 0x0147},
	{LowerCase, 0x0148, 0x0148},
	{TitleCase, 0x0148, 0x0147},

	// Lowercase lower than uppercase.
	// AB78;CHEROKEE SMALL LETTER GE;Ll;0;L;;;;;N;;;13A8;;13A8
	{UpperCase, 0xab78, 0x13a8},
	{LowerCase, 0xab78, 0xab78},
	{TitleCase, 0xab78, 0x13a8},
	{UpperCase, 0x13a8, 0x13a8},
	{LowerCase, 0x13a8, 0xab78},
	{TitleCase, 0x13a8, 0x13a8},

	// Last block in the 5.1.0 table
	// 10400;DESERET CAPITAL LETTER LONG I;Lu;0;L;;;;;N;;;;10428;
	{UpperCase, 0x10400, 0x10400},
	{LowerCase, 0x10400, 0x10428},
	{TitleCase, 0x10400, 0x10400},
	// 10427;DESERET CAPITAL LETTER EW;Lu;0;L;;;;;N;;;;1044F;
	{UpperCase, 0x10427, 0x10427},
	{LowerCase, 0x10427, 0x1044F},
	{TitleCase, 0x10427, 0x10427},
	// 10428;DESERET SMALL LETTER LONG I;Ll;0;L;;;;;N;;;10400;;10400
	{UpperCase, 0x10428, 0x10400},
	{LowerCase, 0x10428, 0x10428},
	{TitleCase, 0x10428, 0x10400},
	// 1044F;DESERET SMALL LETTER EW;Ll;0;L;;;;;N;;;10427;;10427
	{UpperCase, 0x1044F, 0x10427},
	{LowerCase, 0x1044F, 0x1044F},
	{TitleCase, 0x1044F, 0x10427},

	// First one not in the 5.1.0 table
	// 10450;SHAVIAN LETTER PEEP;Lo;0;L;;;;;N;;;;;
	{UpperCase, 0x10450, 0x10450},
	{LowerCase, 0x10450, 0x10450},
	{TitleCase, 0x10450, 0x10450},

	// Non-letters with case.
	{LowerCase, 0x2161, 0x2171},
	{UpperCase, 0x0345, 0x0399},
}

func TestIsLetter(t *testing.T) {
	for _, r := range upperTest {
		if !IsLetter(r) {
			t.Errorf("IsLetter(U+%04X) = false, want true", r)
		}
	}
	for _, r := range letterTest {
		if !IsLetter(r) {
			t.Errorf("IsLetter(U+%04X) = false, want true", r)
		}
	}
	for _, r := range notletterTest {
		if IsLetter(r) {
			t.Errorf("IsLetter(U+%04X) = true, want false", r)
		}
	}
}

func TestIsUpper(t *testing.T) {
	for _, r := range upperTest {
		if !IsUpper(r) {
			t.Errorf("IsUpper(U+%04X) = false, want true", r)
		}
	}
	for _, r := range notupperTest {
		if IsUpper(r) {
			t.Errorf("IsUpper(U+%04X) = true, want false", r)
		}
	}
	for _, r := range notletterTest {
		if IsUpper(r) {
			t.Errorf("IsUpper(U+%04X) = true, want false", r)
		}
	}
}

// Independently check that the special "Is" functions work
// in the Latin-1 range through the property table.

func TestIsLetterLatin1(t *testing.T) {
	for i := rune(0); i <= MaxLatin1; i++ {
		got := IsLetter(i)
		want := Is(Letter, i)
		if got != want {
			t.Errorf("%U incorrect: got %t; want %t", i, got, want)
		}
	}
}

func TestIsUpperLatin1(t *testing.T) {
	for i := rune(0); i <= MaxLatin1; i++ {
		got := IsUpper(i)
		want := Is(Upper, i)
		if got != want {
			t.Errorf("%U incorrect: got %t; want %t", i, got, want)
		}
	}
}

func TestIsLowerLatin1(t *testing.T) {
	for i := rune(0); i <= MaxLatin1; i++ {
		got := IsLower(i)
		want := Is(Lower, i)
		if got != want {
			t.Errorf("%U incorrect: got %t; want %t", i, got, want)
		}
	}
}

func TestIsSpaceLatin1(t *testing.T) {
	for i := rune(0); i <= MaxLatin1; i++ {
		got := IsSpace(i)
		want := Is(White_Space, i)
		if got != want {
			t.Errorf("%U incorrect: got %t; want %t", i, got, want)
		}
	}
}

var testDigit = []rune{
	0x0030,
	0x0039,
	0x0661,
	0x06F1,
	0x07C9,
	0x0966,
	0x09EF,
	0x0A66,
	0x0AEF,
	0x0B66,
	0x0B6F,
	0x0BE6,
	0x0BEF,
	0x0C66,
	0x0CEF,
	0x0D66,
	0x0D6F,
	0x0E50,
	0x0E59,
	0x0ED0,
	0x0ED9,
	0x0F20,
	0x0F29,
	0x1040,
	0x1049,
	0x1090,
	0x1091,
	0x1099,
	0x17E0,
	0x17E9,
	0x1810,
	0x1819,
	0x1946,
	0x194F,
	0x19D0,
	0x19D9,
	0x1B50,
	0x1B59,
	0x1BB0,
	0x1BB9,
	0x1C40,
	0x1C49,
	0x1C50,
	0x1C59,
	0xA620,
	0xA629,
	0xA8D0,
	0xA8D9,
	0xA900,
	0xA909,
	0xAA50,
	0xAA59,
	0xFF10,
	0xFF19,
	0x104A1,
	0x1D7CE,
}

var testLetter = []rune{
	0x0041,
	0x0061,
	0x00AA,
	0x00BA,
	0x00C8,
	0x00DB,
	0x00F9,
	0x02EC,
	0x0535,
	0x06E6,
	0x093D,
	0x0A15,
	0x0B99,
	0x0DC0,
	0x0EDD,
	0x1000,
	0x1200,
	0x1312,
	0x1401,
	0x1885,
	0x2C00,
	0xA800,
	0xF900,
	0xFA30,
	0xFFDA,
	0xFFDC,
	0x10000,
	0x10300,
	0x10400,
	0x20000,
	0x2F800,
	0x2FA1D,
}

func TestDigit(t *testing.T) {
	for _, r := range testDigit {
		if !IsDigit(r) {
			t.Errorf("IsDigit(U+%04X) = false, want true", r)
		}
	}
	for _, r := range testLetter {
		if IsDigit(r) {
			t.Errorf("IsDigit(U+%04X) = true, want false", r)
		}
	}
}

// Test that the special case in IsDigit agrees with the table
func TestDigitOptimization(t *testing.T) {
	for i := rune(0); i <= MaxLatin1; i++ {
		if Is(Digit, i) != IsDigit(i) {
			t.Errorf("IsDigit(U+%04X) disagrees with Is(Digit)", i)
		}
	}
}

func caseString(c int) string {
	switch c {
	case UpperCase:
		return "UpperCase"
	case LowerCase:
		return "LowerCase"
	case TitleCase:
		return "TitleCase"
	}
	return "ErrorCase"
}

func TestTo(t *testing.T) {
	for _, c := range caseTest {
		r := To(c.cas, c.in)
		if c.out != r {
			t.Errorf("To(U+%04X, %s) = U+%04X want U+%04X", c.in, caseString(c.cas), r, c.out)
		}
	}
}

func TestToUpperCase(t *testing.T) {
	for _, c := range caseTest {
		if c.cas != UpperCase {
			continue
		}
		r := ToUpper(c.in)
		if c.out != r {
			t.Errorf("ToUpper(U+%04X) = U+%04X want U+%04X", c.in, r, c.out)
		}
	}
}

func TestToLowerCase(t *testing.T) {
	for _, c := range caseTest {
		if c.cas != LowerCase {
			continue
		}
		r := ToLower(c.in)
		if c.out != r {
			t.Errorf("ToLower(U+%04X) = U+%04X want U+%04X", c.in, r, c.out)
		}
	}
}

func TestToTitleCase(t *testing.T) {
	for _, c := range caseTest {
		if c.cas != TitleCase {
			continue
		}
		r := ToTitle(c.in)
		if c.out != r {
			t.Errorf("ToTitle(U+%04X) = U+%04X want U+%04X", c.in, r, c.out)
		}
	}
}

func TestIsSpace(t *testing.T) {
	for _, c := range spaceTest {
		if !IsSpace(c) {
			t.Errorf("IsSpace(U+%04X) = false; want true", c)
		}
	}
	for _, c := range letterTest {
		if IsSpace(c) {
			t.Errorf("IsSpace(U+%04X) = true; want false", c)
		}
	}
}

// Check that the optimizations for IsLetter etc. agree with the tables.
// We only need to check the Latin-1 range.
func TestLetterOptimizations(t *testing.T) {
	for i := rune(0); i <= MaxLatin1; i++ {
		if Is(Letter, i) != IsLetter(i) {
			t.Errorf("IsLetter(U+%04X) disagrees with Is(Letter)", i)
		}
		if Is(Upper, i) != IsUpper(i) {
			t.Errorf("IsUpper(U+%04X) disagrees with Is(Upper)", i)
		}
		if Is(Lower, i) != IsLower(i) {
			t.Errorf("IsLower(U+%04X) disagrees with Is(Lower)", i)
		}
		if Is(Title, i) != IsTitle(i) {
			t.Errorf("IsTitle(U+%04X) disagrees with Is(Title)", i)
		}
		if Is(White_Space, i) != IsSpace(i) {
			t.Errorf("IsSpace(U+%04X) disagrees with Is(White_Space)", i)
		}
		if To(UpperCase, i) != ToUpper(i) {
			t.Errorf("ToUpper(U+%04X) disagrees with To(Upper)", i)
		}
		if To(LowerCase, i) != ToLower(i) {
			t.Errorf("ToLower(U+%04X) disagrees with To(Lower)", i)
		}
		if To(TitleCase, i) != ToTitle(i) {
			t.Errorf("ToTitle(U+%04X) disagrees with To(Title)", i)
		}
	}
}

func TestNegativeRune(t *testing.T) {
	// Issue 43254
	// These tests cover negative rune handling by testing values which,
	// when cast to uint8 or uint16, look like a particular valid rune.
	// This package has Latin-1-specific optimizations, so we test all of
	// Latin-1 and representative non-Latin-1 values in the character
	// categories covered by IsGraphic, etc.
	nonLatin1 := []uint32{
		// Lu: LATIN CAPITAL LETTER A WITH MACRON
		0x0100,
		// Ll: LATIN SMALL LETTER A WITH MACRON
		0x0101,
		// Lt: LATIN CAPITAL LETTER D WITH SMALL LETTER Z WITH CARON
		0x01C5,
		// M: COMBINING GRAVE ACCENT
		0x0300,
		// Nd: ARABIC-INDIC DIGIT ZERO
		0x0660,
		// P: GREEK QUESTION MARK
		0x037E,
		// S: MODIFIER LETTER LEFT ARROWHEAD
		0x02C2,
		// Z: OGHAM SPACE MARK
		0x1680,
	}
	for i := 0; i < MaxLatin1+len(nonLatin1); i++ {
		base := uint32(i)
		if i >= MaxLatin1 {
			base = nonLatin1[i-MaxLatin1]
		}

		// Note r is negative, but uint8(r) == uint8(base) and
		// uint16(r) == uint16(base).
		r := rune(base - 1<<31)
		if Is(Letter, r) {
			t.Errorf("Is(Letter, 0x%x - 1<<31) = true, want false", base)
		}
		if IsControl(r) {
			t.Errorf("IsControl(0x%x - 1<<31) = true, want false", base)
		}
		if IsDigit(r) {
			t.Errorf("IsDigit(0x%x - 1<<31) = true, want false", base)
		}
		if IsLetter(r) {
			t.Errorf("IsLetter(0x%x - 1<<31) = true, want false", base)
		}
		if IsLower(r) {
			t.Errorf("IsLower(0x%x - 1<<31) = true, want false", base)
		}
		if IsSpace(r) {
			t.Errorf("IsSpace(0x%x - 1<<31) = true, want false", base)
		}
		if IsTitle(r) {
			t.Errorf("IsTitle(0x%x - 1<<31) = true, want false", base)
		}
		if IsUpper(r) {
			t.Errorf("IsUpper(0x%x - 1<<31) = true, want false", base)
		}
	}
}
