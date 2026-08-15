package unicode_test

import (
	"solod.dev/so/testing"
	"solod.dev/so/unicode"
)

var upperTest = [...]rune{
	0x41, 0xc0, 0xd8, 0x100, 0x139, 0x14a, 0x178, 0x181,
	0x376, 0x3cf, 0x13bd, 0x1f2a, 0x2102, 0x2c00, 0x2c10, 0x2c20,
	0xa650, 0xa722, 0xff3a, 0x10400, 0x1d400, 0x1d7ca,
}

var notupperTest = [...]rune{
	0x40, 0x5b, 0x61, 0x185, 0x1b0, 0x377, 0x387, 0x2150,
	0xab7d, 0xffff, 0x10000,
}

var letterTest = [...]rune{
	0x41, 0x61, 0xaa, 0xba, 0xc8, 0xdb, 0xf9, 0x2ec,
	0x535, 0x620, 0x6e6, 0x93d, 0xa15, 0xb99, 0xdc0, 0xedd,
	0x1000, 0x1200, 0x1312, 0x1401, 0x2c00, 0xa800, 0xf900, 0xfa30,
	0xffda, 0xffdc, 0x10000, 0x10300, 0x10400, 0x20000, 0x2f800, 0x2fa1d,
}

var notletterTest = [...]rune{
	0x20, 0x35, 0x375, 0x619, 0x700, 0x1885, 0xfffe, 0x1ffff, 0x10ffff,
}

// spaceTest holds every special cased Latin-1 space character.
var spaceTest = [...]rune{
	0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x20, 0x85, 0xA0, 0x2000, 0x3000,
}

var digitTest = [...]rune{
	0x0030, 0x0039, 0x0661, 0x06F1, 0x07C9, 0x0966, 0x09EF, 0x0A66,
	0x0AEF, 0x0B66, 0x0B6F, 0x0BE6, 0x0BEF, 0x0C66, 0x0CEF, 0x0D66,
	0x0D6F, 0x0E50, 0x0E59, 0x0ED0, 0x0ED9, 0x0F20, 0x0F29, 0x1040,
	0x1049, 0x1090, 0x1091, 0x1099, 0x17E0, 0x17E9, 0x1810, 0x1819,
	0x1946, 0x194F, 0x19D0, 0x19D9, 0x1B50, 0x1B59, 0x1BB0, 0x1BB9,
	0x1C40, 0x1C49, 0x1C50, 0x1C59, 0xA620, 0xA629, 0xA8D0, 0xA8D9,
	0xA900, 0xA909, 0xAA50, 0xAA59, 0xFF10, 0xFF19, 0x104A1, 0x1D7CE,
}

// caseT is one case mapping: To(cas, in) must return out.
type caseT struct {
	cas int
	in  rune
	out rune
}

var caseTest = [...]caseT{
	// errors
	{-1, '\n', 0xFFFD},
	{unicode.UpperCase, -1, -1},
	{unicode.UpperCase, 1 << 30, 1 << 30},

	// ASCII, which the package special cases, so test it carefully.
	{unicode.UpperCase, '\n', '\n'},
	{unicode.UpperCase, 'a', 'A'},
	{unicode.UpperCase, 'A', 'A'},
	{unicode.UpperCase, '7', '7'},
	{unicode.LowerCase, '\n', '\n'},
	{unicode.LowerCase, 'a', 'a'},
	{unicode.LowerCase, 'A', 'a'},
	{unicode.LowerCase, '7', '7'},
	{unicode.TitleCase, '\n', '\n'},
	{unicode.TitleCase, 'a', 'A'},
	{unicode.TitleCase, 'A', 'A'},
	{unicode.TitleCase, '7', '7'},

	// Latin-1
	{unicode.UpperCase, 0x80, 0x80},
	{unicode.UpperCase, 'Å', 'Å'},
	{unicode.UpperCase, 'å', 'Å'},
	{unicode.LowerCase, 0x80, 0x80},
	{unicode.LowerCase, 'Å', 'å'},
	{unicode.LowerCase, 'å', 'å'},
	{unicode.TitleCase, 0x80, 0x80},
	{unicode.TitleCase, 'Å', 'Å'},
	{unicode.TitleCase, 'å', 'Å'},

	// LATIN SMALL LETTER DOTLESS I
	{unicode.UpperCase, 0x0131, 'I'},
	{unicode.LowerCase, 0x0131, 0x0131},
	{unicode.TitleCase, 0x0131, 'I'},

	// LATIN SMALL LIGATURE IJ
	{unicode.UpperCase, 0x0133, 0x0132},
	{unicode.LowerCase, 0x0133, 0x0133},
	{unicode.TitleCase, 0x0133, 0x0132},

	// KELVIN SIGN
	{unicode.UpperCase, 0x212A, 0x212A},
	{unicode.LowerCase, 0x212A, 'k'},
	{unicode.TitleCase, 0x212A, 0x212A},

	// From an UpperLower sequence: CYRILLIC LETTER ZEMLYA
	{unicode.UpperCase, 0xA640, 0xA640},
	{unicode.LowerCase, 0xA640, 0xA641},
	{unicode.TitleCase, 0xA640, 0xA640},
	{unicode.UpperCase, 0xA641, 0xA640},
	{unicode.LowerCase, 0xA641, 0xA641},
	{unicode.TitleCase, 0xA641, 0xA640},
	// CYRILLIC LETTER NEUTRAL YER
	{unicode.UpperCase, 0xA64E, 0xA64E},
	{unicode.LowerCase, 0xA64E, 0xA64F},
	{unicode.TitleCase, 0xA64E, 0xA64E},
	// CYRILLIC LETTER YN
	{unicode.UpperCase, 0xA65F, 0xA65E},
	{unicode.LowerCase, 0xA65F, 0xA65F},
	{unicode.TitleCase, 0xA65F, 0xA65E},

	// From another UpperLower sequence: LATIN LETTER L WITH ACUTE
	{unicode.UpperCase, 0x0139, 0x0139},
	{unicode.LowerCase, 0x0139, 0x013A},
	{unicode.TitleCase, 0x0139, 0x0139},
	// LATIN LETTER L WITH MIDDLE DOT
	{unicode.UpperCase, 0x013f, 0x013f},
	{unicode.LowerCase, 0x013f, 0x0140},
	{unicode.TitleCase, 0x013f, 0x013f},
	// LATIN LETTER N WITH CARON
	{unicode.UpperCase, 0x0148, 0x0147},
	{unicode.LowerCase, 0x0148, 0x0148},
	{unicode.TitleCase, 0x0148, 0x0147},

	// The lowercase code point is below the uppercase one.
	// CHEROKEE LETTER GE
	{unicode.UpperCase, 0xab78, 0x13a8},
	{unicode.LowerCase, 0xab78, 0xab78},
	{unicode.TitleCase, 0xab78, 0x13a8},
	{unicode.UpperCase, 0x13a8, 0x13a8},
	{unicode.LowerCase, 0x13a8, 0xab78},
	{unicode.TitleCase, 0x13a8, 0x13a8},

	// The last block in the 5.1.0 table: DESERET LETTER LONG I
	{unicode.UpperCase, 0x10400, 0x10400},
	{unicode.LowerCase, 0x10400, 0x10428},
	{unicode.TitleCase, 0x10400, 0x10400},
	// DESERET LETTER EW
	{unicode.UpperCase, 0x10427, 0x10427},
	{unicode.LowerCase, 0x10427, 0x1044F},
	{unicode.TitleCase, 0x10427, 0x10427},
	{unicode.UpperCase, 0x10428, 0x10400},
	{unicode.LowerCase, 0x10428, 0x10428},
	{unicode.TitleCase, 0x10428, 0x10400},
	{unicode.UpperCase, 0x1044F, 0x10427},
	{unicode.LowerCase, 0x1044F, 0x1044F},
	{unicode.TitleCase, 0x1044F, 0x10427},

	// The first code point after the 5.1.0 table: SHAVIAN LETTER PEEP
	{unicode.UpperCase, 0x10450, 0x10450},
	{unicode.LowerCase, 0x10450, 0x10450},
	{unicode.TitleCase, 0x10450, 0x10450},

	// Non-letters with case.
	{unicode.LowerCase, 0x2161, 0x2171},
	{unicode.UpperCase, 0x0345, 0x0399},
}

func TestIsLetter(t *testing.T) {
	for _, r := range upperTest {
		if !unicode.IsLetter(r) {
			t.Errorf("IsLetter(U+%04X) = false, want true", r)
		}
	}
	for _, r := range letterTest {
		if !unicode.IsLetter(r) {
			t.Errorf("IsLetter(U+%04X) = false, want true", r)
		}
	}
	for _, r := range notletterTest {
		if unicode.IsLetter(r) {
			t.Errorf("IsLetter(U+%04X) = true, want false", r)
		}
	}
}

func TestIsUpper(t *testing.T) {
	for _, r := range upperTest {
		if !unicode.IsUpper(r) {
			t.Errorf("IsUpper(U+%04X) = false, want true", r)
		}
	}
	for _, r := range notupperTest {
		if unicode.IsUpper(r) {
			t.Errorf("IsUpper(U+%04X) = true, want false", r)
		}
	}
	for _, r := range notletterTest {
		if unicode.IsUpper(r) {
			t.Errorf("IsUpper(U+%04X) = true, want false", r)
		}
	}
}

func TestIsDigit(t *testing.T) {
	for _, r := range digitTest {
		if !unicode.IsDigit(r) {
			t.Errorf("IsDigit(U+%04X) = false, want true", r)
		}
	}
	for _, r := range letterTest {
		if unicode.IsDigit(r) {
			t.Errorf("IsDigit(U+%04X) = true, want false", r)
		}
	}
}

func TestIsSpace(t *testing.T) {
	for _, r := range spaceTest {
		if !unicode.IsSpace(r) {
			t.Errorf("IsSpace(U+%04X) = false, want true", r)
		}
	}
	for _, r := range letterTest {
		if unicode.IsSpace(r) {
			t.Errorf("IsSpace(U+%04X) = true, want false", r)
		}
	}
}

func TestIsControl(t *testing.T) {
	for r := rune(0); r <= 0x1f; r++ {
		if !unicode.IsControl(r) {
			t.Errorf("IsControl(U+%04X) = false, want true", r)
		}
	}
	for r := rune(0x7f); r <= 0x9f; r++ {
		if !unicode.IsControl(r) {
			t.Errorf("IsControl(U+%04X) = false, want true", r)
		}
	}
	for _, r := range letterTest {
		if unicode.IsControl(r) {
			t.Errorf("IsControl(U+%04X) = true, want false", r)
		}
	}
}

func TestIn(t *testing.T) {
	if !unicode.In('5', unicode.Letter, unicode.Digit) {
		t.Error("In('5', Letter, Digit) = false, want true")
	}
	if !unicode.In('a', unicode.Letter, unicode.Digit) {
		t.Error("In('a', Letter, Digit) = false, want true")
	}
	if unicode.In(' ', unicode.Letter, unicode.Digit) {
		t.Error("In(' ', Letter, Digit) = true, want false")
	}
	// A code point above the 16-bit ranges reaches the R32 search.
	if !unicode.In(0x1D400, unicode.Letter) {
		t.Error("In(U+1D400, Letter) = false, want true")
	}
}

func TestLatin1Optimizations(t *testing.T) {
	// Checks the Latin-1 fast paths against the tables.
	// Only the Latin-1 range needs the check.
	for r := rune(0); r <= unicode.MaxLatin1; r++ {
		if unicode.Is(unicode.Letter, r) != unicode.IsLetter(r) {
			t.Errorf("IsLetter(U+%04X) disagrees with Is(Letter)", r)
		}
		if unicode.Is(unicode.Upper, r) != unicode.IsUpper(r) {
			t.Errorf("IsUpper(U+%04X) disagrees with Is(Upper)", r)
		}
		if unicode.Is(unicode.Lower, r) != unicode.IsLower(r) {
			t.Errorf("IsLower(U+%04X) disagrees with Is(Lower)", r)
		}
		if unicode.Is(unicode.Title, r) != unicode.IsTitle(r) {
			t.Errorf("IsTitle(U+%04X) disagrees with Is(Title)", r)
		}
		if unicode.Is(unicode.White_Space, r) != unicode.IsSpace(r) {
			t.Errorf("IsSpace(U+%04X) disagrees with Is(White_Space)", r)
		}
		if unicode.Is(unicode.Digit, r) != unicode.IsDigit(r) {
			t.Errorf("IsDigit(U+%04X) disagrees with Is(Digit)", r)
		}
		if unicode.To(unicode.UpperCase, r) != unicode.ToUpper(r) {
			t.Errorf("ToUpper(U+%04X) disagrees with To(UpperCase)", r)
		}
		if unicode.To(unicode.LowerCase, r) != unicode.ToLower(r) {
			t.Errorf("ToLower(U+%04X) disagrees with To(LowerCase)", r)
		}
		if unicode.To(unicode.TitleCase, r) != unicode.ToTitle(r) {
			t.Errorf("ToTitle(U+%04X) disagrees with To(TitleCase)", r)
		}
	}
}

func TestTo(t *testing.T) {
	for _, c := range caseTest {
		if got := unicode.To(c.cas, c.in); got != c.out {
			t.Errorf("To(%d, U+%04X) = U+%04X, want U+%04X", c.cas, c.in, got, c.out)
		}
	}
}

func TestToUpper(t *testing.T) {
	for _, c := range caseTest {
		if c.cas != unicode.UpperCase {
			continue
		}
		if got := unicode.ToUpper(c.in); got != c.out {
			t.Errorf("ToUpper(U+%04X) = U+%04X, want U+%04X", c.in, got, c.out)
		}
	}
}

func TestToLower(t *testing.T) {
	for _, c := range caseTest {
		if c.cas != unicode.LowerCase {
			continue
		}
		if got := unicode.ToLower(c.in); got != c.out {
			t.Errorf("ToLower(U+%04X) = U+%04X, want U+%04X", c.in, got, c.out)
		}
	}
}

func TestToTitle(t *testing.T) {
	for _, c := range caseTest {
		if c.cas != unicode.TitleCase {
			continue
		}
		if got := unicode.ToTitle(c.in); got != c.out {
			t.Errorf("ToTitle(U+%04X) = U+%04X, want U+%04X", c.in, got, c.out)
		}
	}
}

// nonLatin1 holds one code point of each category that TestNegativeRune covers.
var nonLatin1 = [...]uint32{
	0x0100, // Lu: LATIN CAPITAL LETTER A WITH MACRON
	0x0101, // Ll: LATIN SMALL LETTER A WITH MACRON
	0x01C5, // Lt: LATIN CAPITAL LETTER D WITH SMALL LETTER Z WITH CARON
	0x0300, // M: COMBINING GRAVE ACCENT
	0x0660, // Nd: ARABIC-INDIC DIGIT ZERO
	0x037E, // P: GREEK QUESTION MARK
	0x02C2, // S: MODIFIER LETTER LEFT ARROWHEAD
	0x1680, // Z: OGHAM SPACE MARK
}

func TestNegativeRune(t *testing.T) {
	// Checks the negative runes that look like a valid rune after a conversion
	// to uint8 or uint16. The package has Latin-1 fast paths, so the test
	// covers all of Latin-1 and one code point of each non-Latin-1 category.
	for i := 0; i < unicode.MaxLatin1+len(nonLatin1); i++ {
		base := uint32(i)
		if i >= unicode.MaxLatin1 {
			base = nonLatin1[i-unicode.MaxLatin1]
		}

		// r is negative, but uint8(r) == uint8(base) and uint16(r) == uint16(base).
		r := rune(base - 1<<31)
		if unicode.Is(unicode.Letter, r) {
			t.Errorf("Is(Letter, 0x%x - 1<<31) = true, want false", base)
		}
		if unicode.IsControl(r) {
			t.Errorf("IsControl(0x%x - 1<<31) = true, want false", base)
		}
		if unicode.IsDigit(r) {
			t.Errorf("IsDigit(0x%x - 1<<31) = true, want false", base)
		}
		if unicode.IsLetter(r) {
			t.Errorf("IsLetter(0x%x - 1<<31) = true, want false", base)
		}
		if unicode.IsLower(r) {
			t.Errorf("IsLower(0x%x - 1<<31) = true, want false", base)
		}
		if unicode.IsSpace(r) {
			t.Errorf("IsSpace(0x%x - 1<<31) = true, want false", base)
		}
		if unicode.IsTitle(r) {
			t.Errorf("IsTitle(0x%x - 1<<31) = true, want false", base)
		}
		if unicode.IsUpper(r) {
			t.Errorf("IsUpper(0x%x - 1<<31) = true, want false", base)
		}
	}
}
