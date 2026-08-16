package bytes_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
	"solod.dev/so/unicode/utf8"
)

// maxWord is the length of the longest word of a search sweep.
const maxWord = 8

// maxSep is the length of the longest separator of a search sweep.
const maxSep = 4

// A sweep enumerates every word of an alphabet up to a length.
type sweep struct {
	alpha   string // the letters of the words
	maxWord int    // the length of the longest word
	maxSep  int    // the length of the longest separator
}

// The sweeps of the tests. The first alphabet has two letters, so its words
// repeat and overlap in every way. The second alphabet has a NUL byte and a
// high byte, which the functions must treat as ordinary bytes.
var sweeps = []sweep{
	{"ab", maxWord, maxSep},
	{"\x00a\xff", 5, 3},
}

// hexDigits holds the digits of a hexadecimal number.
const hexDigits = "0123456789abcdef"

// dump writes s into buf as hexadecimal and returns the result. so/fmt has no
// %q verb, and a word can hold a NUL byte or a high byte, so a message gives
// the bytes of a word as hexadecimal. The result is a view of buf, so every
// word of one message needs a buffer of its own.
func dump(buf []byte, s []byte) string {
	for i, c := range s {
		buf[2*i] = hexDigits[c>>4]
		buf[2*i+1] = hexDigits[c&0xf]
	}
	return string(buf[:2*len(s)])
}

// wordCount returns the number of words of alpha with the length n.
func wordCount(alpha string, n int) int {
	count := 1
	for range n {
		count *= len(alpha)
	}
	return count
}

// wordTotal returns the number of words of alpha with a length up to max.
func wordTotal(alpha string, max int) int {
	total := 0
	for n := 0; n <= max; n++ {
		total += wordCount(alpha, n)
	}
	return total
}

// wordAt writes the word number i of alpha into buf and returns the result.
// The shorter words come first, and every word with a length up to max appears
// once. The caller must keep i below wordTotal(alpha, max).
func wordAt(buf []byte, alpha string, max, i int) []byte {
	for n := 0; n <= max; n++ {
		count := wordCount(alpha, n)
		if i < count {
			for k := 0; k < n; k++ {
				buf[k] = alpha[i%len(alpha)]
				i /= len(alpha)
			}
			return buf[:n]
		}
		i -= count
	}
	return nil
}

// The reference implementations. Every sweep checks the package against the
// simplest code that gives the wanted result.

// indexBrute returns the index of the first sep in s, or -1.
func indexBrute(s, sep []byte) int {
	for i := 0; i+len(sep) <= len(s); i++ {
		if string(s[i:i+len(sep)]) == string(sep) {
			return i
		}
	}
	return -1
}

// countBrute returns the number of separate sep in s. The separator must not
// be empty, because an empty separator counts the code points of s.
func countBrute(s, sep []byte) int {
	n := 0
	for i := 0; i+len(sep) <= len(s); {
		if string(s[i:i+len(sep)]) == string(sep) {
			n++
			i += len(sep)
			continue
		}
		i++
	}
	return n
}

// A binOp is a test case of a function that takes two byte slices.
type binOp struct {
	a    string
	b    string
	want int
}

var indexCases = []binOp{
	{"", "", 0},
	{"", "a", -1},
	{"", "foo", -1},
	{"fo", "foo", -1},
	{"foo", "baz", -1},
	{"foo", "foo", 0},
	{"oofofoofooo", "f", 2},
	{"oofofoofooo", "foo", 4},
	{"barfoobarfoo", "foo", 3},
	{"foo", "", 0},
	{"foo", "o", 1},
	{"abcABCabc", "A", 3},
	// One byte separators take the IndexByte path.
	{"x", "a", -1},
	{"x", "x", 0},
	{"abc", "a", 0},
	{"abc", "b", 1},
	{"abc", "c", 2},
	{"abc", "x", -1},
	{"barfoobarfooyyyzzzyyyzzzyyyzzzyyyxxxzzzyyy", "x", 33},
	// The words below overlap, so the search restarts many times.
	{"fofofofooofoboo", "oo", 7},
	{"fofofofofofoboo", "ob", 11},
	{"fofofofofofoboo", "boo", 12},
	{"fofofofofofoboo", "oboo", 11},
	{"fofofofofoooboo", "fooo", 8},
	{"fofofofofofoboo", "foboo", 10},
	{"fofofofofofoboo", "fofob", 8},
	{"fofofofofofofoffofoobarfoo", "foffof", 12},
	{"fofofofofoofofoffofoobarfoo", "foffof", 13},
	{"fofofofofofofoffofoobarfoo", "foffofo", 12},
	{"fofofofofoofofoffofoobarfoo", "foffofoo", 13},
	{"fofofofofofofoffofoobarfoo", "foffofoob", 12},
	{"fofofofofoofofoffofoobarfoo", "foffofoobar", 13},
	{"fofofofofofofoffofoobarfoo", "foffofoobarfoo", 12},
	{"fofofofofoofofoffofoobarfoo", "ofoffofoobarfoo", 12},
	{"fofofofofofofoffofoobarfoo", "fofoffofoobarfoo", 10},
	{"fofofofofoofofoffofoobarfoo", "foobars", -1},
	{"foofyfoobarfoobar", "y", 4},
	{"oooooooooooooooooooooo", "r", -1},
	{"oxoxoxoxoxoxoxoxoxoxoxoy", "oy", 22},
	{"oxoxoxoxoxoxoxoxoxoxoxox", "oy", -1},
	// A long separator takes the Rabin-Karp path.
	{"000000000000000000000000000000000000000000000000000000000000000000000001", "0000000000000000000000000000000000000000000000000000000000000000001", 5},
	// A multibyte code point is a plain byte sequence for Index.
	{"oxoxoxoxoxoxoxoxoxoxox☺", "☺", 22},
	{"xx0123456789012345678901234567890123456789012345678901234567890120123456789012345678901234567890123456xxx\xed\x9f\xc0", "\xed\x9f\xc0", 105},
}

func TestClone(t *testing.T) {
	alloc := t.Allocator()
	src := []byte("hello")
	clone := bytes.Clone(alloc, src)
	defer mem.FreeSlice(alloc, clone)
	if string(clone) != "hello" {
		t.Errorf("Clone(hello) = %s, want hello", string(clone))
		return
	}
	// The clone owns its bytes, so a write does not reach the source.
	clone[0] = 'j'
	if string(src) != "hello" {
		t.Errorf("Clone shares the bytes of the source: %s", string(src))
	}
}

func TestCloneEmpty(t *testing.T) {
	alloc := t.Allocator()
	// A clone of an empty slice is empty, and a clone of nil is empty.
	empty := bytes.Clone(alloc, []byte{})
	defer mem.FreeSlice(alloc, empty)
	if len(empty) != 0 {
		t.Errorf("Clone(empty) has length %d, want 0", len(empty))
	}
	fromNil := bytes.Clone(alloc, nil)
	defer mem.FreeSlice(alloc, fromNil)
	if len(fromNil) != 0 {
		t.Errorf("Clone(nil) has length %d, want 0", len(fromNil))
	}
}

var compareCases = []binOp{
	{"", "", 0},
	{"a", "", +1},
	{"", "a", -1},
	{"abc", "abc", 0},
	{"abd", "abc", +1},
	{"abc", "abd", -1},
	{"ab", "abc", -1},
	{"abc", "ab", +1},
	{"x", "ab", +1},
	{"ab", "x", -1},
	{"x", "a", +1},
	{"b", "x", -1},
	{"abcdefgh", "abcdefgh", 0},
	{"abcdefghi", "abcdefghi", 0},
	{"abcdefghi", "abcdefghj", -1},
	{"abcdefghj", "abcdefghi", +1},
	// The result is -1 or +1, not the difference of the bytes.
	{"a", "z", -1},
	{"z", "a", +1},
	{"\x00", "\xff", -1},
	{"\x7f", "\x80", -1},
}

func TestCompare(t *testing.T) {
	// The second argument moves through a buffer, so the comparison runs at
	// every alignment.
	const shifts = 16
	var buf [shifts + 32]byte
	for i, tc := range compareCases {
		for offset := 0; offset <= shifts; offset++ {
			b := buf[offset : offset+len(tc.b)]
			copy(b, tc.b)
			if got := bytes.Compare([]byte(tc.a), b); got != tc.want {
				t.Errorf("Compare() case %d at offset %d = %d, want %d",
					i, offset, got, tc.want)
				return
			}
		}
	}
}

func TestCompareNil(t *testing.T) {
	// A nil argument is the same as an empty slice.
	if got := bytes.Compare(nil, nil); got != 0 {
		t.Errorf("Compare(nil, nil) = %d, want 0", got)
	}
	if got := bytes.Compare([]byte{}, nil); got != 0 {
		t.Errorf("Compare(empty, nil) = %d, want 0", got)
	}
	if got := bytes.Compare(nil, []byte("a")); got != -1 {
		t.Errorf("Compare(nil, a) = %d, want -1", got)
	}
	if got := bytes.Compare([]byte("a"), nil); got != +1 {
		t.Errorf("Compare(a, nil) = %d, want +1", got)
	}
}

func TestCompareSame(t *testing.T) {
	// One slice compares equal with itself, and after a shorter view.
	b := []byte("Hello Gophers!")
	if got := bytes.Compare(b, b); got != 0 {
		t.Errorf("Compare(b, b) = %d, want 0", got)
	}
	if got := bytes.Compare(b, b[:1]); got != +1 {
		t.Errorf("Compare(b, b[:1]) = %d, want +1", got)
	}
}

func TestCompareLengths(t *testing.T) {
	// Every length up to 128 gets the same bytes in both slices, then one byte
	// of b moves up and down.
	const n = 128
	var abuf, bbuf [n]byte
	for i := range n {
		// Data that repeats slowly, and holds no 0 or 255.
		abuf[i] = byte(1 + 31*i%254)
		bbuf[i] = abuf[i]
	}
	for size := 0; size <= n; size++ {
		a, b := abuf[:size], bbuf[:size]
		if got := bytes.Compare(a, b); got != 0 {
			t.Errorf("Compare of equal slices with length %d = %d, want 0", size, got)
			return
		}
		if size == 0 {
			continue
		}
		if got := bytes.Compare(a[:size-1], b); got != -1 {
			t.Errorf("Compare of the shorter a with length %d = %d, want -1", size, got)
			return
		}
		if got := bytes.Compare(a, b[:size-1]); got != +1 {
			t.Errorf("Compare of the shorter b with length %d = %d, want +1", size, got)
			return
		}
		for k := 0; k < size; k++ {
			bbuf[k] = abuf[k] - 1
			if got := bytes.Compare(a, b); got != +1 {
				t.Errorf("Compare with a lower byte %d of %d = %d, want +1", k, size, got)
				return
			}
			bbuf[k] = abuf[k] + 1
			if got := bytes.Compare(a, b); got != -1 {
				t.Errorf("Compare with a higher byte %d of %d = %d, want -1", k, size, got)
				return
			}
			bbuf[k] = abuf[k]
		}
	}
}

func TestCompareEndian(t *testing.T) {
	// The two slices differ at two byte positions next to each other, in
	// opposite directions. A comparison of whole words with the wrong byte
	// order gives the wrong result.
	const n = 512
	var abuf, bbuf [n]byte
	for i := range n {
		abuf[i] = byte(1 + 31*i%254)
		bbuf[i] = abuf[i]
	}
	a, b := abuf[:], bbuf[:]
	for size := 2; size <= n; size *= 2 {
		for j := 0; j < size-1; j++ {
			abuf[j] = bbuf[j] - 1
			abuf[j+1] = bbuf[j+1] + 1
			if got := bytes.Compare(a[:size], b[:size]); got != -1 {
				t.Errorf("Compare with a lower byte %d of %d = %d, want -1", j, size, got)
				return
			}
			abuf[j] = bbuf[j] + 1
			abuf[j+1] = bbuf[j+1] - 1
			if got := bytes.Compare(a[:size], b[:size]); got != +1 {
				t.Errorf("Compare with a higher byte %d of %d = %d, want +1", j, size, got)
				return
			}
			abuf[j] = bbuf[j]
			abuf[j+1] = bbuf[j+1]
		}
	}
}

func TestEqual(t *testing.T) {
	// Equal agrees with Compare on every compare case.
	for i, tc := range compareCases {
		got := bytes.Equal([]byte(tc.a), []byte(tc.b))
		if got != (tc.want == 0) {
			t.Errorf("Equal() case %d = %t, want %t", i, got, tc.want == 0)
		}
	}
}

func TestEqualNil(t *testing.T) {
	// A nil argument is the same as an empty slice.
	if !bytes.Equal(nil, nil) {
		t.Error("Equal(nil, nil) = false")
	}
	if !bytes.Equal(nil, []byte{}) {
		t.Error("Equal(nil, empty) = false")
	}
	if bytes.Equal(nil, []byte("a")) {
		t.Error("Equal(nil, a) = true")
	}
}

func TestNotEqual(t *testing.T) {
	// The two buffers hold zeros, and one byte of a moves to 1. Every pair of
	// views with the same length and a difference inside must not be equal.
	const size = 24
	var abuf, bbuf [size]byte
	for length := 1; length <= size; length++ {
		for x := 0; x+length <= size; x++ {
			for y := 0; y+length <= size; y++ {
				for diff := x; diff < x+length; diff++ {
					abuf[diff] = 1
					a, b := abuf[x:x+length], bbuf[y:y+length]
					if bytes.Equal(a, b) || bytes.Equal(b, a) {
						t.Errorf("Equal with length %d at %d, %d, diff %d = true",
							length, x, y, diff)
						return
					}
					abuf[diff] = 0
				}
			}
		}
	}
}

func TestContains(t *testing.T) {
	if !bytes.Contains([]byte("hello"), []byte("hel")) {
		t.Error("Contains(hello, hel) = false")
	}
	if !bytes.Contains([]byte("日本語"), []byte("日本")) {
		t.Error("Contains(日本語, 日本) = false")
	}
	if !bytes.Contains([]byte("hello"), []byte{}) {
		t.Error("Contains(hello, empty) = false")
	}
	if bytes.Contains([]byte("hello"), []byte("Hello, world")) {
		t.Error("Contains(hello, Hello, world) = true")
	}
	if bytes.Contains([]byte("東京"), []byte("京東")) {
		t.Error("Contains(東京, 京東) = true")
	}
}

func TestHasPrefix(t *testing.T) {
	b := []byte("hello")
	if !bytes.HasPrefix(b, []byte("he")) {
		t.Error("HasPrefix(hello, he) = false")
	}
	if !bytes.HasPrefix(b, []byte("hello")) {
		t.Error("HasPrefix(hello, hello) = false")
	}
	if !bytes.HasPrefix(b, []byte{}) {
		t.Error("HasPrefix(hello, empty) = false")
	}
	if !bytes.HasPrefix([]byte{}, []byte{}) {
		t.Error("HasPrefix(empty, empty) = false")
	}
	if bytes.HasPrefix(b, []byte("lo")) {
		t.Error("HasPrefix(hello, lo) = true")
	}
	if bytes.HasPrefix(b, []byte("hello!")) {
		t.Error("HasPrefix(hello, hello!) = true")
	}
}

func TestHasSuffix(t *testing.T) {
	b := []byte("hello")
	if !bytes.HasSuffix(b, []byte("lo")) {
		t.Error("HasSuffix(hello, lo) = false")
	}
	if !bytes.HasSuffix(b, []byte("hello")) {
		t.Error("HasSuffix(hello, hello) = false")
	}
	if !bytes.HasSuffix(b, []byte{}) {
		t.Error("HasSuffix(hello, empty) = false")
	}
	if bytes.HasSuffix(b, []byte("he")) {
		t.Error("HasSuffix(hello, he) = true")
	}
	if bytes.HasSuffix(b, []byte("ohello")) {
		t.Error("HasSuffix(hello, ohello) = true")
	}
}

func TestIndex(t *testing.T) {
	for i, tc := range indexCases {
		if got := bytes.Index([]byte(tc.a), []byte(tc.b)); got != tc.want {
			t.Errorf("Index() case %d = %d, want %d", i, got, tc.want)
		}
	}
}

func TestIndexSweep(t *testing.T) {
	// Index must agree with a brute force search for every pair of words.
	for _, sw := range sweeps {
		var sbuf, sepbuf [maxWord]byte
		sTotal := wordTotal(sw.alpha, sw.maxWord)
		sepTotal := wordTotal(sw.alpha, sw.maxSep)
		for i := range sTotal {
			s := wordAt(sbuf[:], sw.alpha, sw.maxWord, i)
			for j := range sepTotal {
				sep := wordAt(sepbuf[:], sw.alpha, sw.maxSep, j)
				got, want := bytes.Index(s, sep), indexBrute(s, sep)
				if got != want {
					var d1, d2 [2 * maxWord]byte
					t.Errorf("Index(%s, %s) = %d, want %d",
						dump(d1[:], s), dump(d2[:], sep), got, want)
					return
				}
			}
		}
	}
}

func TestIndexRabinKarp(t *testing.T) {
	// A search that fails many times gives up on IndexByte and switches to
	// Rabin-Karp. The separator below matches only at the end of s, so every
	// earlier position fails.
	const n = 200
	const sepLen = 31
	var buf [n]byte
	var sepbuf [sepLen]byte
	for i := range n {
		buf[i] = 'a'
	}
	for i := range sepLen - 1 {
		sepbuf[i] = 'a'
	}
	sepbuf[sepLen-1] = 'b'
	s, sep := buf[:], sepbuf[:]
	if got := bytes.Index(s, sep); got != -1 {
		t.Errorf("Index without a match = %d, want -1", got)
		return
	}
	for at := sepLen - 1; at < n; at++ {
		buf[at] = 'b'
		got, want := bytes.Index(s, sep), at-sepLen+1
		buf[at] = 'a'
		if got != want {
			t.Errorf("Index with a match at %d = %d, want %d", at, got, want)
			return
		}
	}
}

func TestIndexByte(t *testing.T) {
	for i, tc := range indexCases {
		if len(tc.b) != 1 {
			continue
		}
		if got := bytes.IndexByte([]byte(tc.a), tc.b[0]); got != tc.want {
			t.Errorf("IndexByte() case %d = %d, want %d", i, got, tc.want)
		}
	}
}

func TestIndexByteNil(t *testing.T) {
	if got := bytes.IndexByte(nil, 'a'); got != -1 {
		t.Errorf("IndexByte(nil, a) = %d, want -1", got)
	}
	if got := bytes.IndexByte([]byte{}, 'a'); got != -1 {
		t.Errorf("IndexByte(empty, a) = %d, want -1", got)
	}
	if got := bytes.IndexByte([]byte("a\x00b"), 0); got != 1 {
		t.Errorf("IndexByte(a\\x00b, NUL) = %d, want 1", got)
	}
}

func TestIndexByteAlign(t *testing.T) {
	// The search runs at every start and end alignment of a buffer.
	const n = 130
	var buf [n]byte
	for start := range n {
		b := buf[start:]
		for j := range b {
			b[j] = 'x'
			got := bytes.IndexByte(b, 'x')
			b[j] = 0
			if got != j {
				t.Errorf("IndexByte from %d at %d = %d, want %d", start, j, got, j)
				return
			}
			if got := bytes.IndexByte(b, 'x'); got != -1 {
				t.Errorf("IndexByte from %d after the reset = %d, want -1", start, got)
				return
			}
		}
	}
	for end := range n {
		b := buf[:end]
		for j := range b {
			b[j] = 'x'
			got := bytes.IndexByte(b, 'x')
			b[j] = 0
			if got != j {
				t.Errorf("IndexByte up to %d at %d = %d, want %d", end, j, got, j)
				return
			}
		}
	}
}

func TestIndexByteWindow(t *testing.T) {
	// A byte outside the window must not match. The window moves through a
	// buffer that holds the wanted byte everywhere else.
	const n = 512
	const window = 15
	var buf [n]byte
	for i := range n {
		buf[i] = 'x'
	}
	for at := 0; at+window <= n; at++ {
		b := buf[at : at+window]
		for j := range window {
			b[j] = 'y'
		}
		if got := bytes.IndexByte(b, 'x'); got != -1 {
			t.Errorf("IndexByte outside the window at %d = %d, want -1", at, got)
			return
		}
		for j := range window {
			b[j] = 'x'
		}
	}
}

func TestCount(t *testing.T) {
	if got := bytes.Count([]byte("cheese"), []byte("e")); got != 3 {
		t.Errorf("Count(cheese, e) = %d, want 3", got)
	}
	if got := bytes.Count([]byte("cheese"), []byte("x")); got != 0 {
		t.Errorf("Count(cheese, x) = %d, want 0", got)
	}
	if got := bytes.Count([]byte("banana"), []byte("ana")); got != 1 {
		t.Errorf("Count(banana, ana) = %d, want 1", got)
	}
	if got := bytes.Count([]byte("aaaa"), []byte("aa")); got != 2 {
		t.Errorf("Count(aaaa, aa) = %d, want 2", got)
	}
	if got := bytes.Count(nil, []byte("a")); got != 0 {
		t.Errorf("Count(nil, a) = %d, want 0", got)
	}
}

func TestCountEmptySep(t *testing.T) {
	// An empty separator counts the code points of s, plus one.
	if got := bytes.Count([]byte(""), []byte("")); got != 1 {
		t.Errorf("Count(empty, empty) = %d, want 1", got)
	}
	if got := bytes.Count([]byte("abc"), []byte("")); got != 4 {
		t.Errorf("Count(abc, empty) = %d, want 4", got)
	}
	if got := bytes.Count([]byte("日本語"), []byte("")); got != 4 {
		t.Errorf("Count(日本語, empty) = %d, want 4", got)
	}
	if got := bytes.Count([]byte("hello world"), []byte("")); got != 12 {
		t.Errorf("Count(hello world, empty) = %d, want 12", got)
	}
}

func TestCountSweep(t *testing.T) {
	// Count must agree with a brute force count for every pair of words. The
	// sweep starts at the separator number 1, because the separator number 0
	// is the empty word, which counts code points.
	for _, sw := range sweeps {
		var sbuf, sepbuf [maxWord]byte
		sTotal := wordTotal(sw.alpha, sw.maxWord)
		sepTotal := wordTotal(sw.alpha, sw.maxSep)
		for i := range sTotal {
			s := wordAt(sbuf[:], sw.alpha, sw.maxWord, i)
			for j := 1; j < sepTotal; j++ {
				sep := wordAt(sepbuf[:], sw.alpha, sw.maxSep, j)
				got, want := bytes.Count(s, sep), countBrute(s, sep)
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

func TestCountWindow(t *testing.T) {
	// A byte outside the window must not count. The window moves through a
	// buffer that holds the counted byte everywhere else.
	const n = 512
	const window = 15
	var buf [n]byte
	for i := range n {
		buf[i] = 'x'
	}
	for at := 0; at+window <= n; at++ {
		b := buf[at : at+window]
		for j := range window {
			b[j] = 'y'
		}
		if got := bytes.Count(b, []byte("x")); got != 0 {
			t.Errorf("Count outside the window at %d = %d, want 0", at, got)
			return
		}
		for j := range window {
			b[j] = 'x'
		}
	}
}

// A cutCase is a test case of Cut.
type cutCase struct {
	s      string
	sep    string
	before string
	after  string
	found  bool
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
	for i, tc := range cutCases {
		got := bytes.Cut([]byte(tc.s), []byte(tc.sep))
		if string(got.Before) != tc.before {
			t.Errorf("Cut() case %d before = %s, want %s", i, string(got.Before), tc.before)
		}
		if string(got.After) != tc.after {
			t.Errorf("Cut() case %d after = %s, want %s", i, string(got.After), tc.after)
		}
		if got.Found != tc.found {
			t.Errorf("Cut() case %d found = %t, want %t", i, got.Found, tc.found)
		}
	}
}

func TestCutView(t *testing.T) {
	// Cut returns views of s, not copies. The bytes of s come from an array,
	// because a string literal is read-only.
	var buf [9]byte
	s := buf[:]
	copy(s, "go is fun")
	got := bytes.Cut(s, []byte(" is "))
	got.Before[0] = 'G'
	got.After[0] = 'F'
	if string(s) != "Go is Fun" {
		t.Errorf("Cut copies the bytes: %s", string(s))
	}
}

func TestJoin(t *testing.T) {
	alloc := t.Allocator()
	parts := [][]byte{[]byte("go"), []byte("is"), []byte("fun")}
	joined := bytes.Join(alloc, parts, []byte(" "))
	defer mem.FreeSlice(alloc, joined)
	if string(joined) != "go is fun" {
		t.Errorf("Join with a space = %s, want go is fun", string(joined))
	}
	empty := bytes.Join(alloc, parts, nil)
	defer mem.FreeSlice(alloc, empty)
	if string(empty) != "goisfun" {
		t.Errorf("Join with no separator = %s, want goisfun", string(empty))
	}
}

func TestJoinShort(t *testing.T) {
	alloc := t.Allocator()
	// No element gives an empty result.
	none := bytes.Join(alloc, nil, []byte(","))
	defer mem.FreeSlice(alloc, none)
	if len(none) != 0 {
		t.Errorf("Join of no element has length %d, want 0", len(none))
	}
	// One element gives a copy of that element.
	src := []byte("go")
	one := [][]byte{src}
	got := bytes.Join(alloc, one, []byte(","))
	defer mem.FreeSlice(alloc, got)
	if string(got) != "go" {
		t.Errorf("Join of one element = %s, want go", string(got))
		return
	}
	got[0] = 'G'
	if string(src) != "go" {
		t.Errorf("Join of one element shares the bytes of the source: %s", string(src))
	}
}

// A repeatCase is a test case of Repeat.
type repeatCase struct {
	in    string
	count int
	want  string
}

var repeatCases = []repeatCase{
	{"", 0, ""},
	{"", 1, ""},
	{"", 2, ""},
	{"-", 0, ""},
	{"-", 1, "-"},
	{"-", 10, "----------"},
	{"abc ", 3, "abc abc abc "},
	{"\x00\xff", 2, "\x00\xff\x00\xff"},
}

func TestRepeat(t *testing.T) {
	alloc := t.Allocator()
	for i, tc := range repeatCases {
		got := bytes.Repeat(alloc, []byte(tc.in), tc.count)
		ok := string(got) == tc.want
		mem.FreeSlice(alloc, got)
		if !ok {
			t.Errorf("Repeat() case %d gives the wrong result", i)
			return
		}
	}
}

func TestRepeatLong(t *testing.T) {
	// A result above the chunk limit of 8KB reuses one chunk of the source.
	alloc := t.Allocator()
	const count = 12 * 1024
	got := bytes.Repeat(alloc, []byte("ab"), count)
	defer mem.FreeSlice(alloc, got)
	if len(got) != 2*count {
		t.Errorf("Repeat above the chunk limit has length %d, want %d", len(got), 2*count)
		return
	}
	for i := range got {
		want := byte('a')
		if i%2 == 1 {
			want = 'b'
		}
		if got[i] != want {
			t.Errorf("Repeat above the chunk limit gives %c at %d, want %c",
				got[i], i, want)
			return
		}
	}
}

// A replaceCase is a test case of Replace.
type replaceCase struct {
	in   string
	old  string
	new  string
	n    int
	want string
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
	{"☺☻☹", "", "<>", -1, "<>☺<>☻<>☹<>"},
}

func TestReplace(t *testing.T) {
	alloc := t.Allocator()
	for i, tc := range replaceCases {
		got := bytes.Replace(alloc, []byte(tc.in), []byte(tc.old), []byte(tc.new), tc.n)
		ok := string(got) == tc.want
		mem.FreeSlice(alloc, got)
		if !ok {
			t.Errorf("Replace() case %d gives the wrong result", i)
			return
		}
	}
}

func TestReplaceCopy(t *testing.T) {
	// Replace always gives a new slice, also when it replaces nothing.
	alloc := t.Allocator()
	src := []byte("hello")
	got := bytes.Replace(alloc, src, []byte("x"), []byte("y"), -1)
	defer mem.FreeSlice(alloc, got)
	if string(got) != "hello" {
		t.Errorf("Replace without a match = %s, want hello", string(got))
		return
	}
	got[0] = 'j'
	if string(src) != "hello" {
		t.Errorf("Replace shares the bytes of the source: %s", string(src))
	}
}

func TestRunes(t *testing.T) {
	alloc := t.Allocator()
	got := bytes.Runes(alloc, []byte("a日本語z"))
	defer mem.FreeSlice(alloc, got)
	if len(got) != 5 {
		t.Errorf("Runes(a日本語z) has length %d, want 5", len(got))
		return
	}
	want := []rune{'a', 0x65e5, 0x672c, 0x8a9e, 'z'}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Runes(a日本語z) gives %x at %d, want %x", got[i], i, want[i])
		}
	}
}

func TestRunesEmpty(t *testing.T) {
	alloc := t.Allocator()
	got := bytes.Runes(alloc, nil)
	defer mem.FreeSlice(alloc, got)
	if len(got) != 0 {
		t.Errorf("Runes(nil) has length %d, want 0", len(got))
	}
}

func TestRunesInvalid(t *testing.T) {
	// An invalid byte gives one RuneError.
	alloc := t.Allocator()
	got := bytes.Runes(alloc, []byte("ab\x80c"))
	defer mem.FreeSlice(alloc, got)
	if len(got) != 4 {
		t.Errorf("Runes(ab\\x80c) has length %d, want 4", len(got))
		return
	}
	if got[2] != utf8.RuneError {
		t.Errorf("Runes(ab\\x80c) gives %x at 2, want %x", got[2], utf8.RuneError)
	}
}

func TestString(t *testing.T) {
	alloc := t.Allocator()
	var buf [5]byte
	src := buf[:]
	copy(src, "hello")
	s := bytes.String(alloc, src)
	defer mem.FreeString(alloc, s)
	if s != "hello" {
		t.Errorf("String(hello) = %s, want hello", s)
		return
	}
	// The string owns its bytes, so a write to the source does not reach it.
	src[0] = 'j'
	if s != "hello" {
		t.Errorf("String shares the bytes of the source: %s", s)
	}
}
