// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
	"solod.dev/so/unicode/utf8"
)

// A containsCase is a test case of Contains.
type containsCase struct {
	s, substr string
	want      bool
}

// The cases hold a word of every length that the search treats apart. Every
// length gets the substring at the start, in the middle, at the end, and
// twice with a word that only looks like a match.
var containsCases = []containsCase{
	{"abc", "bc", true},
	{"abc", "bcd", false},
	{"abc", "", true},
	{"", "a", false},

	// 2-byte substring
	{"xxxxxx", "01", false},
	{"01xxxx", "01", true},
	{"xx01xx", "01", true},
	{"xxxx01", "01", true},
	{"1xxxxx", "01", false},
	{"xxxxx0", "01", false},
	// 3-byte substring
	{"xxxxxxx", "012", false},
	{"012xxxx", "012", true},
	{"xx012xx", "012", true},
	{"xxxx012", "012", true},
	{"12xxxxx", "012", false},
	{"xxxxx01", "012", false},
	// 4-byte substring
	{"xxxxxxxx", "0123", false},
	{"0123xxxx", "0123", true},
	{"xx0123xx", "0123", true},
	{"xxxx0123", "0123", true},
	{"123xxxxx", "0123", false},
	{"xxxxx012", "0123", false},
	// 5 to 7 byte substring
	{"xxxxxxxxx", "01234", false},
	{"01234xxxx", "01234", true},
	{"xx01234xx", "01234", true},
	{"xxxx01234", "01234", true},
	{"1234xxxxx", "01234", false},
	{"xxxxx0123", "01234", false},
	// 8-byte substring
	{"xxxxxxxxxxxx", "01234567", false},
	{"01234567xxxx", "01234567", true},
	{"xx01234567xx", "01234567", true},
	{"xxxx01234567", "01234567", true},
	{"1234567xxxxx", "01234567", false},
	{"xxxxx0123456", "01234567", false},
	// 9 to 15 byte substring
	{"xxxxxxxxxxxxx", "012345678", false},
	{"012345678xxxx", "012345678", true},
	{"xx012345678xx", "012345678", true},
	{"xxxx012345678", "012345678", true},
	{"12345678xxxxx", "012345678", false},
	{"xxxxx01234567", "012345678", false},
	// 16-byte substring
	{"xxxxxxxxxxxxxxxxxxxx", "0123456789ABCDEF", false},
	{"0123456789ABCDEFxxxx", "0123456789ABCDEF", true},
	{"xx0123456789ABCDEFxx", "0123456789ABCDEF", true},
	{"xxxx0123456789ABCDEF", "0123456789ABCDEF", true},
	{"123456789ABCDEFxxxxx", "0123456789ABCDEF", false},
	{"xxxxx0123456789ABCDE", "0123456789ABCDEF", false},
	// 17 to 31 byte substring
	{"xxxxxxxxxxxxxxxxxxxxx", "0123456789ABCDEFG", false},
	{"0123456789ABCDEFGxxxx", "0123456789ABCDEFG", true},
	{"xx0123456789ABCDEFGxx", "0123456789ABCDEFG", true},
	{"xxxx0123456789ABCDEFG", "0123456789ABCDEFG", true},
	{"123456789ABCDEFGxxxxx", "0123456789ABCDEFG", false},
	{"xxxxx0123456789ABCDEF", "0123456789ABCDEFG", false},

	// A word that holds every byte of the substring except the last one.
	{"xx01x", "012", false},
	{"xx0123x", "01234", false},
	{"xx01234567x", "012345678", false},
	{"xx0123456789ABCDEFx", "0123456789ABCDEFG", false},
}

func TestContains(t *testing.T) {
	for i, tc := range containsCases {
		if got := strings.Contains(tc.s, tc.substr); got != tc.want {
			t.Errorf("case %d: Contains(%s, %s) = %t, want %t",
				i, tc.s, tc.substr, got, tc.want)
		}
	}
}

var containsAnyCases = []containsCase{
	{"", "", false},
	{"", "a", false},
	{"", "abc", false},
	{"a", "", false},
	{"a", "a", true},
	{"aaa", "a", true},
	{"abc", "xyz", false},
	{"abc", "xcz", true},
	{"a☺b☻c☹d", "uvw☻xyz", true},
	{"aRegExp*", ".(|)*+?^$[]", true},
	{dotsThrice, " ", false},
}

func TestContainsAny(t *testing.T) {
	for i, tc := range containsAnyCases {
		if got := strings.ContainsAny(tc.s, tc.substr); got != tc.want {
			t.Errorf("case %d: ContainsAny(%s, %s) = %t, want %t",
				i, tc.s, tc.substr, got, tc.want)
		}
	}
}

// A containsRuneCase is a test case of ContainsRune and ContainsFunc.
type containsRuneCase struct {
	s    string
	r    rune
	want bool
}

var containsRuneCases = []containsRuneCase{
	{"", 'a', false},
	{"a", 'a', true},
	{"aaa", 'a', true},
	{"abc", 'y', false},
	{"abc", 'c', true},
	{"a☺b☻c☹d", 'x', false},
	{"a☺b☻c☹d", '☻', true},
	{"aRegExp*", '*', true},
}

func TestContainsRune(t *testing.T) {
	for i, tc := range containsRuneCases {
		if got := strings.ContainsRune(tc.s, tc.r); got != tc.want {
			t.Errorf("case %d: ContainsRune(%s, %d) = %t, want %t",
				i, tc.s, tc.r, got, tc.want)
		}
	}
}

// wantRune holds the code point that containsRune looks for.
var wantRune rune

// containsRune reports whether the code point is the one of the current case.
func containsRune(r rune) bool {
	return r == wantRune
}

func TestContainsFunc(t *testing.T) {
	for i, tc := range containsRuneCases {
		wantRune = tc.r
		if got := strings.ContainsFunc(tc.s, containsRune); got != tc.want {
			t.Errorf("case %d: ContainsFunc(%s, %d) = %t, want %t",
				i, tc.s, tc.r, got, tc.want)
		}
	}
}

// An indexCase is a test case of Index, LastIndex, IndexByte and IndexAny.
type indexCase struct {
	s, sep string
	want   int
}

var indexCases = []indexCase{
	{"", "", 0},
	{"", "a", -1},
	{"", "foo", -1},
	{"fo", "foo", -1},
	{"foo", "foo", 0},
	{"oofofoofooo", "f", 2},
	{"oofofoofooo", "foo", 4},
	{"barfoobarfoo", "foo", 3},
	{"foo", "", 0},
	{"foo", "o", 1},
	{"abcABCabc", "A", 3},
	{"jrzm6jjhorimglljrea4w3rlgosts0w2gia17hno2td4qd1jz", "jz", 47},
	{"ekkuk5oft4eq0ocpacknhwouic1uua46unx12l37nioq9wbpnocqks6", "ks6", 52},
	{"999f2xmimunbuyew5vrkla9cpwhmxan8o98ec", "98ec", 33},
	{"9lpt9r98i04k8bz6c6dsrthb96bhi", "96bhi", 24},
	{"55u558eqfaod2r2gu42xxsu631xf0zobs5840vl", "5840vl", 33},
	// A separator of one byte takes a special path.
	{"x", "a", -1},
	{"x", "x", 0},
	{"abc", "a", 0},
	{"abc", "b", 1},
	{"abc", "c", 2},
	{"abc", "x", -1},
	// A short separator takes a special path for every length.
	{"", "ab", -1},
	{"bc", "ab", -1},
	{"ab", "ab", 0},
	{"xab", "ab", 1},
	{"xa", "ab", -1},
	{"", "abc", -1},
	{"xbc", "abc", -1},
	{"abc", "abc", 0},
	{"xabc", "abc", 1},
	{"xab", "abc", -1},
	{"xabxc", "abc", -1},
	{"", "abcd", -1},
	{"xbcd", "abcd", -1},
	{"abcd", "abcd", 0},
	{"xabcd", "abcd", 1},
	{"xyabc", "abcd", -1},
	{"xbcqq", "abcqq", -1},
	{"abcqq", "abcqq", 0},
	{"xabcqq", "abcqq", 1},
	{"xyabcq", "abcqq", -1},
	{"xabxcqq", "abcqq", -1},
	{"xabcqxq", "abcqq", -1},
	{"", "01234567", -1},
	{"32145678", "01234567", -1},
	{"01234567", "01234567", 0},
	{"x01234567", "01234567", 1},
	{"x0123456x01234567", "01234567", 9},
	{"xx0123456", "01234567", -1},
	{"", "0123456789", -1},
	{"3214567844", "0123456789", -1},
	{"0123456789", "0123456789", 0},
	{"x0123456789", "0123456789", 1},
	{"x012345678x0123456789", "0123456789", 11},
	{"xyz012345678", "0123456789", -1},
	{"x01234567x89", "0123456789", -1},
	{"", "0123456789012345", -1},
	{"3214567889012345", "0123456789012345", -1},
	{"0123456789012345", "0123456789012345", 0},
	{"x0123456789012345", "0123456789012345", 1},
	{"x012345678901234x0123456789012345", "0123456789012345", 17},
	{"", "01234567890123456789", -1},
	{"32145678890123456789", "01234567890123456789", -1},
	{"01234567890123456789", "01234567890123456789", 0},
	{"x01234567890123456789", "01234567890123456789", 1},
	{"x0123456789012345678x01234567890123456789", "01234567890123456789", 21},
	{"xyz0123456789012345678", "01234567890123456789", -1},
	{"", "0123456789012345678901234567890", -1},
	{"321456788901234567890123456789012345678911", "0123456789012345678901234567890", -1},
	{"0123456789012345678901234567890", "0123456789012345678901234567890", 0},
	{"x0123456789012345678901234567890", "0123456789012345678901234567890", 1},
	{"x012345678901234567890123456789x0123456789012345678901234567890", "0123456789012345678901234567890", 32},
	{"xyz012345678901234567890123456789", "0123456789012345678901234567890", -1},
	{"", "01234567890123456789012345678901", -1},
	{"32145678890123456789012345678901234567890211", "01234567890123456789012345678901", -1},
	{"01234567890123456789012345678901", "01234567890123456789012345678901", 0},
	{"x01234567890123456789012345678901", "01234567890123456789012345678901", 1},
	{"x0123456789012345678901234567890x01234567890123456789012345678901", "01234567890123456789012345678901", 33},
	{"xyz0123456789012345678901234567890", "01234567890123456789012345678901", -1},
	{"xxxxxx012345678901234567890123456789012345678901234567890123456789012", "012345678901234567890123456789012345678901234567890123456789012", 6},
	{"", "0123456789012345678901234567890123456789", -1},
	{"xx012345678901234567890123456789012345678901234567890123456789012", "0123456789012345678901234567890123456789", 2},
	{"xx012345678901234567890123456789012345678", "0123456789012345678901234567890123456789", -1},
	{"xx012345678901234567890123456789012345678901234567890123456789012", "0123456789012345678901234567890123456xxx", -1},
	{"xx0123456789012345678901234567890123456789012345678901234567890120123456789012345678901234567890123456xxx", "0123456789012345678901234567890123456xxx", 65},
	// A long word without a match takes the Rabin-Karp path.
	{"oxoxoxoxoxoxoxoxoxoxoxoy", "oy", 22},
	{"oxoxoxoxoxoxoxoxoxoxoxox", "oy", -1},
	// A separator of one code point takes the IndexRune path.
	{"oxoxoxoxoxoxoxoxoxoxox☺", "☺", 22},
	// An invalid UTF-8 sequence must not take the IndexRune path. The word is
	// longer than the limit of the brute force search.
	{"xx0123456789012345678901234567890123456789012345678901234567890120123456789012345678901234567890123456xxx\xed\x9f\xc0", "\xed\x9f\xc0", 105},
}

func TestIndex(t *testing.T) {
	for i, tc := range indexCases {
		if got := strings.Index(tc.s, tc.sep); got != tc.want {
			t.Errorf("case %d: Index(%s, %s) = %d, want %d",
				i, tc.s, tc.sep, got, tc.want)
		}
	}
}

func TestIndexByte(t *testing.T) {
	// A separator of one byte gives the same result as IndexByte.
	for i, tc := range indexCases {
		if len(tc.sep) != 1 {
			continue
		}
		if got := strings.IndexByte(tc.s, tc.sep[0]); got != tc.want {
			t.Errorf("case %d: IndexByte(%s, %d) = %d, want %d",
				i, tc.s, tc.sep[0], got, tc.want)
		}
	}
}

var lastIndexCases = []indexCase{
	{"", "", 0},
	{"", "a", -1},
	{"", "foo", -1},
	{"fo", "foo", -1},
	{"foo", "foo", 0},
	{"foo", "f", 0},
	{"oofofoofooo", "f", 7},
	{"oofofoofooo", "foo", 7},
	{"barfoobarfoo", "foo", 9},
	{"foo", "", 3},
	{"foo", "o", 2},
	{"abcABCabc", "A", 3},
	{"abcABCabc", "a", 6},
}

func TestLastIndex(t *testing.T) {
	for i, tc := range lastIndexCases {
		if got := strings.LastIndex(tc.s, tc.sep); got != tc.want {
			t.Errorf("case %d: LastIndex(%s, %s) = %d, want %d",
				i, tc.s, tc.sep, got, tc.want)
		}
	}
}

var lastIndexByteCases = []indexCase{
	{"", "q", -1},
	{"abcdef", "q", -1},
	{"abcdefabcdef", "a", 6},  // in the middle
	{"abcdefabcdef", "f", 11}, // the last byte
	{"zabcdefabcdef", "z", 0}, // the first byte
	{"a☺b☻c☹d", "b", 4},       // after a code point of several bytes
}

func TestLastIndexByte(t *testing.T) {
	for i, tc := range lastIndexByteCases {
		if got := strings.LastIndexByte(tc.s, tc.sep[0]); got != tc.want {
			t.Errorf("case %d: LastIndexByte(%s, %d) = %d, want %d",
				i, tc.s, tc.sep[0], got, tc.want)
		}
	}
}

var indexAnyCases = []indexCase{
	{"", "", -1},
	{"", "a", -1},
	{"", "abc", -1},
	{"a", "", -1},
	{"a", "a", 0},
	{"\x80", "\xffb", 0},
	{"aaa", "a", 0},
	{"abc", "xyz", -1},
	{"abc", "xcz", 2},
	{"ab☺c", "x☺yz", 2},
	{"a☺b☻c☹d", "cx", 8},
	{"a☺b☻c☹d", "uvw☻xyz", 5},
	{"aRegExp*", ".(|)*+?^$[]", 7},
	{dotsThrice, " ", -1},
	{"012abcba210", "\xffb", 4},
	{"012\x80bcb\x80210", "\xffb", 3},
	{"0123456\xcf\x80abc", "\xcfb\x80", 10},
}

func TestIndexAny(t *testing.T) {
	for i, tc := range indexAnyCases {
		if got := strings.IndexAny(tc.s, tc.sep); got != tc.want {
			var sbuf, sepbuf [64]byte
			t.Errorf("case %d: IndexAny(%s, %s) = %d, want %d",
				i, dump(sbuf[:], tc.s), dump(sepbuf[:], tc.sep), got, tc.want)
		}
	}
}

// An indexRuneCase is a test case of IndexRune.
type indexRuneCase struct {
	s    string
	r    rune
	want int
}

var indexRuneCases = []indexRuneCase{
	{"", 'a', -1},
	{"", '☺', -1},
	{"foo", '☹', -1},
	{"foo", 'o', 1},
	{"foo☺bar", '☺', 3},
	{"foo☺☻☹bar", '☹', 9},
	{"a A x", 'A', 2},
	{"some_text=some_value", '=', 9},
	{"☺a", 'a', 3},
	{"a☻☺b", '☺', 4},

	// RuneError matches every invalid UTF-8 sequence.
	{"\ufffd", utf8.RuneError, 0},
	{"\xff", utf8.RuneError, 0},
	{"☻x\ufffd", utf8.RuneError, 4},
	{"☻x\xe2\x98", utf8.RuneError, 4},
	{"☻x\xe2\x98\ufffd", utf8.RuneError, 4},
	{"☻x\xe2\x98x", utf8.RuneError, 4},

	// An invalid code point never matches.
	{"a☺b☻c☹d\xe2\x98\ufffd\xff\ufffd\xed\xa0\x80", -1, -1},
	{"a☺b☻c☹d\xe2\x98\ufffd\xff\ufffd\xed\xa0\x80", 0xD800, -1},
	{"a☺b☻c☹d\xe2\x98\ufffd\xff\ufffd\xed\xa0\x80", utf8.MaxRune + 1, -1},

	// A code point of 2 bytes.
	{"ӆ", 'ӆ', 0},
	{"a", 'ӆ', -1},
	{"  ӆ", 'ӆ', 2},
	{"  a", 'ӆ', -1},

	// A code point of 3 bytes.
	{"Ꚁ", 'Ꚁ', 0},
	{"a", 'Ꚁ', -1},
	{"  Ꚁ", 'Ꚁ', 2},
	{"  a", 'Ꚁ', -1},

	// A code point of 4 bytes.
	{"𡌀", '𡌀', 0},
	{"a", '𡌀', -1},
	{"  𡌀", '𡌀', 2},
	{"  a", '𡌀', -1},

	// The search changes to a byte search in the middle of a code point that
	// holds runs of equal bytes.
	{"aaaaaKKKK\U000bc104", '\U000bc104', 17},
	{"aaaaaKKKK鄄", '鄄', 17},
	{"aaKKKKKa\U000bc104", '\U000bc104', 18},
	{"aaKKKKKa鄄", '鄄', 18},
}

func TestIndexRune(t *testing.T) {
	for i, tc := range indexRuneCases {
		if got := strings.IndexRune(tc.s, tc.r); got != tc.want {
			t.Errorf("case %d: IndexRune(%s, %d) = %d, want %d",
				i, tc.s, tc.r, got, tc.want)
		}
	}
}

// A cutoverCase is a test case of IndexRune with a word that is long enough
// to change the search from a code point search to a byte search.
type cutoverCase struct {
	fill string // the code point that repeats 64 times
	tail string // the code point after the repeated part
	r    rune   // the code point to look for
	want int
}

var cutoverCases = []cutoverCase{
	{"ц", "ӆ", 'ӆ', 128},
	{"Ꙁ", "Ꚁ", '䚀', -1}, // Ꚁ and 䚀 share the last two bytes
	{"Ꙁ", "Ꚁ", 'Ꚁ', 192},
	{"𡋀", "𡌀", '𣌀', -1}, // 𡌀 and 𣌀 share the last two bytes
	{"𡋀", "𡌀", '𡌀', 256},
	{"𡋀", "", '𡌀', -1},
}

func TestIndexRuneCutover(t *testing.T) {
	alloc := t.Allocator()
	for i, tc := range cutoverCases {
		fill := strings.Repeat(alloc, tc.fill, 64)
		s := strings.Join(alloc, []string{fill, tc.tail}, "")
		if got := strings.IndexRune(s, tc.r); got != tc.want {
			t.Errorf("case %d: IndexRune() = %d, want %d", i, got, tc.want)
		}
		mem.FreeString(alloc, s)
		mem.FreeString(alloc, fill)
	}
}

// An indexFuncCase is a test case of IndexFunc.
type indexFuncCase struct {
	s    string
	pred int
	want int
}

var indexFuncCases = []indexFuncCase{
	{"", predValidRune, -1},
	{"abc", predDigit, -1},
	{"0123", predDigit, 0},
	{"a1b", predDigit, 1},
	{space, predSpace, 0},
	{"๐๒12hello34๐๑", predDigit, 0},
	{"ⱯⱯⱯⱯABCDhelloEFⱯⱯGHⱯⱯ", predUpper, 0},
	{"12๐๒hello34๐๑", predNotDigit, 8},

	// Invalid UTF-8.
	{"\x801", predDigit, 1},
	{"\x80abc", predDigit, -1},
	{"\xc0a\xc0", predValidRune, 1},
	{"\xc0a\xc0", predNotValidRune, 0},
	{"\xc0☺\xc0", predNotValidRune, 0},
	{"\xc0☺\xc0\xc0", predNotValidRune, 0},
	{"ab\xc0a\xc0cd", predNotValidRune, 2},
	{"a\xe0\x80cd", predNotValidRune, 1},
	{"\x80\x80\x80\x80", predNotValidRune, 0},
}

func TestIndexFunc(t *testing.T) {
	for i, tc := range indexFuncCases {
		got := strings.IndexFunc(tc.s, predicate(tc.pred))
		if got != tc.want {
			t.Errorf("case %d: IndexFunc(%s, %s) = %d, want %d",
				i, tc.s, predName(tc.pred), got, tc.want)
		}
	}
}

func TestIndexSweep(t *testing.T) {
	// Index and LastIndex agree with the reference over every short word.
	var sbuf, sepbuf [maxWord]byte
	for _, sw := range sweeps {
		words := wordTotal(sw.alpha, sw.maxWord)
		seps := wordTotal(sw.alpha, sw.maxSep)
		for i := range words {
			s := wordAt(sbuf[:], sw.alpha, sw.maxWord, i)
			for j := range seps {
				sep := wordAt(sepbuf[:], sw.alpha, sw.maxSep, j)
				var d1, d2 [2 * maxWord]byte
				if got := strings.Index(s, sep); got != indexBrute(s, sep) {
					t.Errorf("Index(%s, %s) = %d, want %d",
						dump(d1[:], s), dump(d2[:], sep), got, indexBrute(s, sep))
					return
				}
				if got := strings.LastIndex(s, sep); got != lastIndexBrute(s, sep) {
					t.Errorf("LastIndex(%s, %s) = %d, want %d",
						dump(d1[:], s), dump(d2[:], sep), got, lastIndexBrute(s, sep))
					return
				}
				if got := strings.Contains(s, sep); got != (indexBrute(s, sep) >= 0) {
					t.Errorf("Contains(%s, %s) = %t",
						dump(d1[:], s), dump(d2[:], sep), got)
					return
				}
			}
		}
	}
}

func TestIndexLong(t *testing.T) {
	// A word of many letters and a separator of many letters go through the
	// paths for the long inputs. The letters of the word repeat, so a partial
	// match happens at almost every position.
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	var sbuf [140]byte
	var sepbuf [141]byte
	x := uint32(1)
	for size := 5; size < 140; size += 10 {
		for i := range size {
			x = nextRand(x)
			sbuf[i] = chars[int(x>>16)%len(chars)]
		}
		s := string(sbuf[:size])
		for k := range 50 {
			x = nextRand(x)
			begin := int(x>>16) % (len(s) + 1)
			x = nextRand(x)
			end := begin + int(x>>16)%(len(s)+1-begin)
			sep := s[begin:end]
			if k%4 == 0 {
				// One letter goes into the separator, so the search fails at
				// a position that almost matches.
				x = nextRand(x)
				pos := int(x>>16) % (len(sep) + 1)
				copy(sepbuf[:], sep[:pos])
				sepbuf[pos] = 'A'
				copy(sepbuf[pos+1:], sep[pos:])
				sep = string(sepbuf[:len(sep)+1])
			}
			want := indexBrute(s, sep)
			if got := strings.Index(s, sep); got != want {
				t.Errorf("Index(%s, %s) = %d, want %d", s, sep, got, want)
				return
			}
		}
	}
}
