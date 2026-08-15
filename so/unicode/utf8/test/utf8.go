package utf8_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
	"solod.dev/so/unicode"
	"solod.dev/so/unicode/utf8"
)

// utf8Case is a rune and the UTF-8 encoding of the rune.
type utf8Case struct {
	r   rune
	str string
}

var utf8map = [...]utf8Case{
	{0x0000, "\x00"},
	{0x0001, "\x01"},
	{0x007e, "\x7e"},
	{0x007f, "\x7f"},
	{0x0080, "\xc2\x80"},
	{0x0081, "\xc2\x81"},
	{0x00bf, "\xc2\xbf"},
	{0x00c0, "\xc3\x80"},
	{0x00c1, "\xc3\x81"},
	{0x00c8, "\xc3\x88"},
	{0x00d0, "\xc3\x90"},
	{0x00e0, "\xc3\xa0"},
	{0x00f0, "\xc3\xb0"},
	{0x00f8, "\xc3\xb8"},
	{0x00ff, "\xc3\xbf"},
	{0x0100, "\xc4\x80"},
	{0x07ff, "\xdf\xbf"},
	{0x0400, "\xd0\x80"},
	{0x0800, "\xe0\xa0\x80"},
	{0x0801, "\xe0\xa0\x81"},
	{0x1000, "\xe1\x80\x80"},
	{0xd000, "\xed\x80\x80"},
	{0xd7ff, "\xed\x9f\xbf"}, // the last code point before the surrogate half
	{0xe000, "\xee\x80\x80"}, // the first code point after the surrogate half
	{0xfffe, "\xef\xbf\xbe"},
	{0xffff, "\xef\xbf\xbf"},
	{0x10000, "\xf0\x90\x80\x80"},
	{0x10001, "\xf0\x90\x80\x81"},
	{0x40000, "\xf1\x80\x80\x80"},
	{0x10fffe, "\xf4\x8f\xbf\xbe"},
	{0x10ffff, "\xf4\x8f\xbf\xbf"},
	{0xFFFD, "\xef\xbf\xbd"},
}

// surrogateMap holds the code points that decode to (RuneError, 1).
var surrogateMap = [...]utf8Case{
	{0xd800, "\xed\xa0\x80"}, // the surrogate minimum
	{0xdfff, "\xed\xbf\xbf"}, // the surrogate maximum
}

var testStrings = [...]string{
	"",
	"abcd",
	"☺☻☹",
	"日a本b語ç日ð本Ê語þ日¥本¼語i日©",
	"日a本b語ç日ð本Ê語þ日¥本¼語i日©日a本b語ç日ð本Ê語þ日¥本¼語i日©日a本b語ç日ð本Ê語þ日¥本¼語i日©",
	"\x80\x80\x80\x80",
}

// badStartBytes holds the bytes that start no UTF-8 sequence.
var badStartBytes = [...]string{"\xc0", "\xc1"}

// TestConstants checks the constants that utf8 repeats from unicode.
func TestConstants(t *testing.T) {
	if utf8.MaxRune != unicode.MaxRune {
		t.Errorf("utf8.MaxRune = %X, want %X", utf8.MaxRune, unicode.MaxRune)
	}
	if utf8.RuneError != unicode.ReplacementChar {
		t.Errorf("utf8.RuneError = %X, want %X", utf8.RuneError, unicode.ReplacementChar)
	}
}

func TestFullRune(t *testing.T) {
	for _, m := range utf8map {
		b := []byte(m.str)
		if !utf8.FullRune(b) {
			t.Errorf("FullRune(U+%04X) = false, want true", m.r)
		}
		if !utf8.FullRuneInString(m.str) {
			t.Errorf("FullRuneInString(U+%04X) = false, want true", m.r)
		}
		// A sequence one byte short is not a full rune.
		b1 := b[0 : len(b)-1]
		if utf8.FullRune(b1) {
			t.Errorf("FullRune(U+%04X cut short) = true, want false", m.r)
		}
		if utf8.FullRuneInString(string(b1)) {
			t.Errorf("FullRuneInString(U+%04X cut short) = true, want false", m.r)
		}
	}
	// A byte that starts no sequence is a full rune of its own.
	for _, s := range badStartBytes {
		if !utf8.FullRune([]byte(s)) {
			t.Error("FullRune of an invalid start byte = false, want true")
		}
		if !utf8.FullRuneInString(s) {
			t.Error("FullRuneInString of an invalid start byte = false, want true")
		}
	}
}

func TestRuneStart(t *testing.T) {
	for _, m := range utf8map {
		b := []byte(m.str)
		if !utf8.RuneStart(b[0]) {
			t.Errorf("RuneStart(U+%04X byte 0) = false, want true", m.r)
		}
		for i := 1; i < len(b); i++ {
			if utf8.RuneStart(b[i]) {
				t.Errorf("RuneStart(U+%04X byte %d) = true, want false", m.r, i)
			}
		}
	}
}

func TestEncodeRune(t *testing.T) {
	for _, m := range utf8map {
		var buf [10]byte
		n := utf8.EncodeRune(buf[:], m.r)
		if !bytes.Equal(buf[:n], []byte(m.str)) {
			t.Errorf("EncodeRune(U+%04X) wrote %d bytes, want %d", m.r, n, len(m.str))
		}
	}
}

func TestAppendRune(t *testing.T) {
	const init = "init"
	for _, m := range utf8map {
		// AppendRune needs UTFMax bytes of spare capacity.
		buf := make([]byte, 0, utf8.UTFMax)
		if got := utf8.AppendRune(buf, m.r); string(got) != m.str {
			t.Errorf("AppendRune(empty, U+%04X) mismatch", m.r)
		}
		pre := make([]byte, 0, len(init)+utf8.UTFMax)
		pre = append(pre, init...)
		if got := utf8.AppendRune(pre, m.r); string(got) != init+m.str {
			t.Errorf("AppendRune(init, U+%04X) mismatch", m.r)
		}
	}
}

func TestDecodeRune(t *testing.T) {
	for _, m := range utf8map {
		b := []byte(m.str)
		r, size := utf8.DecodeRune(b)
		if r != m.r || size != len(b) {
			t.Errorf("DecodeRune(U+%04X) = U+%04X, %d, want U+%04X, %d", m.r, r, size, m.r, len(b))
		}
		r, size = utf8.DecodeRuneInString(m.str)
		if r != m.r || size != len(b) {
			t.Errorf("DecodeRuneInString(U+%04X) = U+%04X, %d, want U+%04X, %d", m.r, r, size, m.r, len(b))
		}

		// A trailing byte must not change the result.
		r, size = utf8.DecodeRuneInString(m.str + "\x00")
		if r != m.r || size != len(b) {
			t.Errorf("DecodeRuneInString(U+%04X + NUL) = U+%04X, %d, want U+%04X, %d", m.r, r, size, m.r, len(b))
		}

		// A sequence one byte short must fail.
		wantSize := 1
		if wantSize >= len(b) {
			wantSize = 0
		}
		r, size = utf8.DecodeRune(b[0 : len(b)-1])
		if r != utf8.RuneError || size != wantSize {
			t.Errorf("DecodeRune(U+%04X cut short) = U+%04X, %d, want RuneError, %d", m.r, r, size, wantSize)
		}
		r, size = utf8.DecodeRuneInString(m.str[0 : len(m.str)-1])
		if r != utf8.RuneError || size != wantSize {
			t.Errorf("DecodeRuneInString(U+%04X cut short) = U+%04X, %d, want RuneError, %d", m.r, r, size, wantSize)
		}

		// A bad sequence must fail. So gives a zero-copy view for
		// []byte(string), so the test copies the bytes before it writes.
		var arr [utf8.UTFMax]byte
		bad := arr[:copy(arr[:], b)]
		if len(bad) == 1 {
			bad[0] = 0x80
		} else {
			bad[len(bad)-1] = 0x7F
		}
		r, size = utf8.DecodeRune(bad)
		if r != utf8.RuneError || size != 1 {
			t.Errorf("DecodeRune(U+%04X corrupted) = U+%04X, %d, want RuneError, 1", m.r, r, size)
		}
		r, size = utf8.DecodeRuneInString(string(bad))
		if r != utf8.RuneError || size != 1 {
			t.Errorf("DecodeRuneInString(U+%04X corrupted) = U+%04X, %d, want RuneError, 1", m.r, r, size)
		}
	}
}

func TestDecodeSurrogateRune(t *testing.T) {
	for _, m := range surrogateMap {
		r, size := utf8.DecodeRune([]byte(m.str))
		if r != utf8.RuneError || size != 1 {
			t.Errorf("DecodeRune(U+%04X) = U+%04X, %d, want RuneError, 1", m.r, r, size)
		}
		r, size = utf8.DecodeRuneInString(m.str)
		if r != utf8.RuneError || size != 1 {
			t.Errorf("DecodeRuneInString(U+%04X) = U+%04X, %d, want RuneError, 1", m.r, r, size)
		}
	}
}

var invalidSequences = [...]string{
	"\xed\xa0\x80\x80", // the surrogate minimum
	"\xed\xbf\xbf\x80", // the surrogate maximum

	// xx
	"\x91\x80\x80\x80",

	// s1
	"\xC2\x7F\x80\x80",
	"\xC2\xC0\x80\x80",
	"\xDF\x7F\x80\x80",
	"\xDF\xC0\x80\x80",

	// s2
	"\xE0\x9F\xBF\x80",
	"\xE0\xA0\x7F\x80",
	"\xE0\xBF\xC0\x80",
	"\xE0\xC0\x80\x80",

	// s3
	"\xE1\x7F\xBF\x80",
	"\xE1\x80\x7F\x80",
	"\xE1\xBF\xC0\x80",
	"\xE1\xC0\x80\x80",

	// s4
	"\xED\x7F\xBF\x80",
	"\xED\x80\x7F\x80",
	"\xED\x9F\xC0\x80",
	"\xED\xA0\x80\x80",

	// s5
	"\xF0\x8F\xBF\xBF",
	"\xF0\x90\x7F\xBF",
	"\xF0\x90\x80\x7F",
	"\xF0\xBF\xBF\xC0",
	"\xF0\xBF\xC0\x80",
	"\xF0\xC0\x80\x80",

	// s6
	"\xF1\x7F\xBF\xBF",
	"\xF1\x80\x7F\xBF",
	"\xF1\x80\x80\x7F",
	"\xF1\xBF\xBF\xC0",
	"\xF1\xBF\xC0\x80",
	"\xF1\xC0\x80\x80",

	// s7
	"\xF4\x7F\xBF\xBF",
	"\xF4\x80\x7F\xBF",
	"\xF4\x80\x80\x7F",
	"\xF4\x8F\xBF\xC0",
	"\xF4\x8F\xC0\x80",
	"\xF4\x90\x80\x80",
}

func TestDecodeInvalidSequence(t *testing.T) {
	for i, s := range invalidSequences {
		r1, _ := utf8.DecodeRune([]byte(s))
		if r1 != utf8.RuneError {
			t.Errorf("DecodeRune(invalidSequences[%d]) = U+%04X, want RuneError", i, r1)
			return
		}
		r2, _ := utf8.DecodeRuneInString(s)
		if r2 != utf8.RuneError {
			t.Errorf("DecodeRuneInString(invalidSequences[%d]) = U+%04X, want RuneError", i, r2)
			return
		}
		// A range loop over a string decodes with the code the compiler
		// emits, not with this package. The two must agree.
		if r3 := rangeDecodeRune(s); r3 != r2 {
			t.Errorf("range over invalidSequences[%d] = U+%04X, want U+%04X", i, r3, r2)
			return
		}
	}
}

func TestSequencing(t *testing.T) {
	// Check that DecodeRune and DecodeLastRune walk a string in
	// the same order as a range loop.
	for _, ts := range testStrings {
		for _, m := range utf8map {
			checkSequence(t, ts, m.str, "")
			checkSequence(t, "", m.str, ts)
			checkSequence(t, ts, m.str, ts)
		}
	}
}

func TestRuneConversion(t *testing.T) {
	// Check that a range loop and a []rune conversion visit the same runes
	// as RuneCountInString. The compiler emits the code for both, so the test
	// covers the compiler, not only this package.
	for _, ts := range testStrings {
		checkRunes(t, ts)
	}
}

func TestNegativeRune(t *testing.T) {
	// Check that a negative rune encodes as U+FFFD.
	var wantBuf, gotBuf [utf8.UTFMax]byte
	want := wantBuf[:utf8.EncodeRune(wantBuf[:], utf8.RuneError)]
	got := gotBuf[:utf8.EncodeRune(gotBuf[:], -1)]
	if !bytes.Equal(got, want) {
		t.Error("EncodeRune(-1) does not encode RuneError")
	}
}

// badRunes holds the runes that UTF-8 cannot represent.
var badRunes = [...]rune{-1, 0x7fffffff, 0xd800, 0xdfff, utf8.MaxRune + 1}

func TestRuneString(t *testing.T) {
	// Check the string conversion of a rune. The compiler emits the code for
	// the conversion, so the test covers the compiler, not only this package.
	for _, tt := range utf8map {
		checkRuneString(t, tt.r, tt.str)
	}
	for _, r := range badRunes {
		checkRuneString(t, r, "\xef\xbf\xbd")
	}
}

type countCase struct {
	in  string
	out int
}

var runeCountTests = [...]countCase{
	{"abcd", 4},
	{"☺☻☹", 3},
	{"1,2,3,4", 7},
	{"\xe2\x00", 2},
	{"\xe2\x80", 2},
	{"a\xe2\x80", 3},
}

func TestRuneCount(t *testing.T) {
	for i, tt := range runeCountTests {
		if got := utf8.RuneCountInString(tt.in); got != tt.out {
			t.Errorf("RuneCountInString(runeCountTests[%d]) = %d, want %d", i, got, tt.out)
		}
		if got := utf8.RuneCount([]byte(tt.in)); got != tt.out {
			t.Errorf("RuneCount(runeCountTests[%d]) = %d, want %d", i, got, tt.out)
		}
	}
}

type lenCase struct {
	r    rune
	size int
}

var runeLenTests = [...]lenCase{
	{0, 1},
	{'e', 1},
	{'é', 2},
	{'☺', 3},
	{utf8.RuneError, 3},
	{utf8.MaxRune, 4},
	{0xD800, -1},
	{0xDFFF, -1},
	{utf8.MaxRune + 1, -1},
	{-1, -1},
}

func TestRuneLen(t *testing.T) {
	for _, tt := range runeLenTests {
		if got := utf8.RuneLen(tt.r); got != tt.size {
			t.Errorf("RuneLen(U+%04X) = %d, want %d", tt.r, got, tt.size)
		}
	}
}

type validCase struct {
	in  string
	out bool
}

var validTests = [...]validCase{
	{"", true},
	{"a", true},
	{"abc", true},
	{"Ж", true},
	{"ЖЖ", true},
	{"брэд-ЛГТМ", true},
	{"☺☻☹", true},
	{"aa\xe2", false},
	{"\x42\xfa", false},
	{"\x42\xfa\x43", false},
	{"a�b", true},
	{"\xF4\x8F\xBF\xBF", true},      // U+10FFFF
	{"\xF4\x90\x80\x80", false},     // U+10FFFF+1, outside the range
	{"\xF7\xBF\xBF\xBF", false},     // 0x1FFFFF, outside the range
	{"\xFB\xBF\xBF\xBF\xBF", false}, // 0x3FFFFFF, outside the range
	{"\xc0\x80", false},             // U+0000 in two bytes, which is incorrect
	{"\xed\xa0\x80", false},         // U+D800, a high surrogate
	{"\xed\xbf\xbf", false},         // U+DFFF, a low surrogate
}

func TestValid(t *testing.T) {
	for i, tt := range validTests {
		if utf8.Valid([]byte(tt.in)) != tt.out {
			t.Errorf("Valid(validTests[%d]) = %t, want %t", i, !tt.out, tt.out)
		}
		if utf8.ValidString(tt.in) != tt.out {
			t.Errorf("ValidString(validTests[%d]) = %t, want %t", i, !tt.out, tt.out)
		}
	}
	// Strings of a growing length reach every alignment of the scan.
	for i := 0; i < 100; i++ {
		astr := strings.Repeat(mem.System, "a", i)
		checkValidRepeat(t, i, astr)
		mem.FreeString(mem.System, astr)
	}
}

type validRuneCase struct {
	r  rune
	ok bool
}

var validRuneTests = [...]validRuneCase{
	{0, true},
	{'e', true},
	{'é', true},
	{'☺', true},
	{utf8.RuneError, true},
	{utf8.MaxRune, true},
	{0xD7FF, true},
	{0xD800, false},
	{0xDFFF, false},
	{0xE000, true},
	{utf8.MaxRune + 1, false},
	{-1, false},
}

func TestValidRune(t *testing.T) {
	for _, tt := range validRuneTests {
		if got := utf8.ValidRune(tt.r); got != tt.ok {
			t.Errorf("ValidRune(U+%04X) = %t, want %t", tt.r, got, tt.ok)
		}
	}
}

// rangeDecodeRune returns the first rune that a range loop gives for s.
func rangeDecodeRune(s string) rune {
	for _, r := range s {
		return r
	}
	return -1
}

// checkSequence joins the three parts and checks the sequence of the result.
// The concatenation allocates on the stack, so it happens in a function that
// returns before the next case runs.
func checkSequence(t *testing.T, a string, b string, c string) {
	s := a + b + c
	seqIndex := make([]int, len(s)+1)
	seqRune := make([]rune, len(s)+1)

	// Walk forward with a range loop and record every rune.
	n := 0
	pos := 0
	for i, r := range s {
		if pos != i {
			t.Errorf("range index = %d, want %d", i, pos)
			return
		}
		seqIndex[n] = i
		seqRune[n] = r
		n++
		r1, size1 := utf8.DecodeRune([]byte(s)[i:])
		if r != r1 {
			t.Errorf("DecodeRune at %d = U+%04X, want U+%04X", i, r1, r)
			return
		}
		r2, size2 := utf8.DecodeRuneInString(s[i:])
		if r != r2 {
			t.Errorf("DecodeRuneInString at %d = U+%04X, want U+%04X", i, r2, r)
			return
		}
		if size1 != size2 {
			t.Errorf("size at %d: DecodeRune = %d, DecodeRuneInString = %d", i, size1, size2)
			return
		}
		pos += size1
	}

	// Walk backward with DecodeLastRune and compare against the record.
	n--
	for pos = len(s); pos > 0; n-- {
		r1, size1 := utf8.DecodeLastRune([]byte(s)[0:pos])
		r2, size2 := utf8.DecodeLastRuneInString(s[0:pos])
		if size1 != size2 {
			t.Errorf("last size at %d: DecodeLastRune = %d, DecodeLastRuneInString = %d", pos, size1, size2)
			return
		}
		if r1 != seqRune[n] {
			t.Errorf("DecodeLastRune at %d = U+%04X, want U+%04X", pos, r1, seqRune[n])
			return
		}
		if r2 != seqRune[n] {
			t.Errorf("DecodeLastRuneInString at %d = U+%04X, want U+%04X", pos, r2, seqRune[n])
			return
		}
		pos -= size1
		if pos != seqIndex[n] {
			t.Errorf("DecodeLastRune index = %d, want %d", pos, seqIndex[n])
			return
		}
	}
	if pos != 0 {
		t.Errorf("the backward walk finished at %d, want 0", pos)
	}
}

// checkRuneString checks the string conversion of r against want. The
// conversion allocates on the stack, so it happens in a function that returns
// before the next rune runs.
func checkRuneString(t *testing.T, r rune, want string) {
	if got := string(r); got != want {
		t.Errorf("string(U+%04X) = %s, want %s", r, got, want)
	}
}

// checkRunes compares a range loop and a []rune conversion of s against
// RuneCountInString. The conversion allocates on the stack, so it happens in a
// function that returns before the next string runs.
func checkRunes(t *testing.T, s string) {
	count := utf8.RuneCountInString(s)
	runes := []rune(s)
	if len(runes) != count {
		t.Errorf("[]rune has length %d, want %d", len(runes), count)
		return
	}
	i := 0
	for _, r := range s {
		if r != runes[i] {
			t.Errorf("rune %d = U+%04X, want U+%04X", i, r, runes[i])
			return
		}
		i++
	}
	if i != count {
		t.Errorf("the range loop visited %d runes, want %d", i, count)
	}
}

// checkValidRepeat checks the validity of astr with a suffix and a wrapper.
// The concatenations allocate on the stack, so they happen in a function that
// returns before the next length runs.
func checkValidRepeat(t *testing.T, n int, astr string) {
	checkValid(t, n, astr, true)
	checkValid(t, n, astr+"Ж", true)
	checkValid(t, n, astr+"\xe2", false)
	checkValid(t, n, astr+"Ж"+astr, true)
	checkValid(t, n, astr+"\xe2"+astr, false)
}

func checkValid(t *testing.T, n int, s string, want bool) {
	if utf8.Valid([]byte(s)) != want {
		t.Errorf("Valid of a %d byte repeat = %t, want %t", n, !want, want)
	}
	if utf8.ValidString(s) != want {
		t.Errorf("ValidString of a %d byte repeat = %t, want %t", n, !want, want)
	}
}
