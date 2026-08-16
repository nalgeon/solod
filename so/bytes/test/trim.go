package bytes_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/testing"
	"solod.dev/so/unicode"
	"solod.dev/so/unicode/utf8"
)

// The trim functions.
const (
	kindTrim = iota
	kindTrimLeft
	kindTrimRight
	kindTrimPrefix
	kindTrimSuffix
)

// trimName returns the name of a trim function.
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

// applyTrim calls the trim function of the kind. Trim, TrimLeft and TrimRight
// take a cutset of code points, and TrimPrefix and TrimSuffix take a slice.
func applyTrim(kind int, s []byte, arg string) []byte {
	switch kind {
	case kindTrim:
		return bytes.Trim(s, arg)
	case kindTrimLeft:
		return bytes.TrimLeft(s, arg)
	case kindTrimRight:
		return bytes.TrimRight(s, arg)
	case kindTrimPrefix:
		return bytes.TrimPrefix(s, []byte(arg))
	}
	return bytes.TrimSuffix(s, []byte(arg))
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
		got := applyTrim(tc.kind, []byte(tc.in), tc.arg)
		if string(got) != tc.want {
			t.Errorf("%s() case %d = %s, want %s",
				trimName(tc.kind), i, string(got), tc.want)
		}
	}
}

// A trimEmptyCase is a test case that must give an empty result.
type trimEmptyCase struct {
	kind int
	in   string
	arg  string
}

var trimEmptyCases = []trimEmptyCase{
	{kindTrim, "a", "a"},
	{kindTrim, "aa", "a"},
	{kindTrim, "a", "ab"},
	{kindTrim, "ab", "ab"},
	{kindTrim, "☺", "☺"},
	{kindTrimLeft, "a", "a"},
	{kindTrimLeft, "aa", "a"},
	{kindTrimLeft, "a", "ab"},
	{kindTrimLeft, "ab", "ab"},
	{kindTrimLeft, "☺", "☺"},
	{kindTrimRight, "a", "a"},
	{kindTrimRight, "aa", "a"},
	{kindTrimRight, "a", "ab"},
	{kindTrimRight, "ab", "ab"},
	{kindTrimRight, "☺", "☺"},
	{kindTrimPrefix, "a", "a"},
	{kindTrimPrefix, "☺", "☺"},
	{kindTrimSuffix, "a", "a"},
	{kindTrimSuffix, "☺", "☺"},
}

func TestTrimEmpty(t *testing.T) {
	// A trim that removes every byte gives a slice with the length 0.
	for i, tc := range trimEmptyCases {
		got := applyTrim(tc.kind, []byte(tc.in), tc.arg)
		if len(got) != 0 {
			t.Errorf("%s() case %d has length %d, want 0",
				trimName(tc.kind), i, len(got))
		}
	}
	for kind := kindTrim; kind <= kindTrimSuffix; kind++ {
		if got := applyTrim(kind, nil, ""); len(got) != 0 {
			t.Errorf("%s(nil) has length %d, want 0", trimName(kind), len(got))
		}
		if got := applyTrim(kind, []byte{}, ""); len(got) != 0 {
			t.Errorf("%s(empty) has length %d, want 0", trimName(kind), len(got))
		}
	}
}

func TestTrimView(t *testing.T) {
	// A trim gives a view of s, not a copy.
	var buf [7]byte
	s := buf[:]
	copy(s, "xxabcxx")
	got := bytes.Trim(s, "x")
	if string(got) != "abc" {
		t.Errorf("Trim(xxabcxx, x) = %s, want abc", string(got))
		return
	}
	got[0] = 'A'
	if string(s) != "xxAbcxx" {
		t.Errorf("Trim copies the bytes: %s", string(s))
	}
}

// A spaceCase is a test case of TrimSpace.
type spaceCase struct {
	in   string
	want string
}

var spaceCases = []spaceCase{
	{"", ""},
	{"  a", "a"},
	{"b  ", "b"},
	{"abc", "abc"},
	{"\t\v\r\f\n  　abc\t\v\r\f\n  　", "abc"},
	{" ", ""},
	{"\u3000 ", ""},
	{" \u3000", ""},
	{" \t\r\n \t\t\r\r\n\n ", ""},
	{" \t\r\n x\t\t\r\r\n\n ", "x"},
	{"  \t\r\n x\t\t\r\r\ny\n 　", "x\t\t\r\r\ny"},
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
	for i, tc := range spaceCases {
		got := bytes.TrimSpace([]byte(tc.in))
		if string(got) != tc.want {
			var d1, d2 [128]byte
			t.Errorf("TrimSpace() case %d = %s, want %s",
				i, dump(d1[:], got), dump(d2[:], []byte(tc.want)))
		}
	}
}

// The predicates of the TrimFunc tests.

// isValidRune reports whether the decoding of the code point succeeded.
func isValidRune(r rune) bool {
	return r != utf8.RuneError
}

// notSpace reports whether the code point is not a space.
func notSpace(r rune) bool {
	return !unicode.IsSpace(r)
}

// notDigit reports whether the code point is not a decimal digit.
func notDigit(r rune) bool {
	return !unicode.IsDigit(r)
}

// notValidRune reports whether the decoding of the code point failed.
func notValidRune(r rune) bool {
	return r == utf8.RuneError
}

// The predicates of the trimFuncCases, by number.
const (
	predSpace = iota
	predDigit
	predUpper
	predValidRune
	predNotSpace
	predNotDigit
	predNotValidRune
)

// predicate returns the predicate of the number.
func predicate(pred int) bytes.RunePredicate {
	switch pred {
	case predSpace:
		return unicode.IsSpace
	case predDigit:
		return unicode.IsDigit
	case predUpper:
		return unicode.IsUpper
	case predValidRune:
		return isValidRune
	case predNotSpace:
		return notSpace
	case predNotDigit:
		return notDigit
	}
	return notValidRune
}

// A trimFuncCase is a test case of TrimFunc.
type trimFuncCase struct {
	pred int
	in   string
	want string
}

var trimFuncCases = []trimFuncCase{
	{predSpace, "\t\v\r\f\n  hello \t\v\r\f\n ", "hello"},
	{predDigit, "๐๒12hello34๐๑", "hello"},
	{predUpper, "ⱯⱯⱯⱯABCDhelloEFⱯⱯGHⱯⱯ", "hello"},
	{predValidRune, "ab\xc0a\xc0cd", "\xc0a\xc0"},
	{predNotSpace, "hello \t\r\n hello", " \t\r\n "},
	{predNotDigit, "hello๐๒1234๐๑helo", "๐๒1234๐๑"},
	{predNotValidRune, "\xc0a\xc0", "a"},
	{predSpace, "", ""},
	{predSpace, " ", ""},
}

func TestTrimFunc(t *testing.T) {
	for i, tc := range trimFuncCases {
		got := bytes.TrimFunc([]byte(tc.in), predicate(tc.pred))
		if string(got) != tc.want {
			var d1, d2 [128]byte
			t.Errorf("TrimFunc() case %d = %s, want %s",
				i, dump(d1[:], got), dump(d2[:], []byte(tc.want)))
		}
	}
}
