package bytealg_test

import (
	"solod.dev/so/bytealg"
	"solod.dev/so/testing"
)

// maxWord is the length of the longest word of a sweep.
const maxWord = 8

// maxSep is the length of the longest separator of a search sweep.
const maxSep = 4

// A sweep enumerates every word of an alphabet up to a length.
type sweep struct {
	alpha   string // the letters of the words
	maxWord int    // the length of the longest word
	probes  string // the bytes that the byte searches look for
}

// The sweeps of the tests. The first alphabet has two letters, so its words
// repeat and overlap in every way. The second alphabet has a NUL byte and a
// high byte, which the functions must treat as ordinary bytes. The last probe
// of each sweep is a byte that no word holds.
var sweeps = []sweep{
	{"ab", maxWord, "abz"},
	{"\x00a\xff", 5, "\x00a\xff\x7f"},
}

// maxPair is the length of the longest word of the compare sweep. A compare
// sweep takes every pair of words, so it uses shorter words than a search
// sweep.
const maxPair = 4

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
	for i := 0; i < n; i++ {
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

// compareBrute compares a and b byte by byte.
func compareBrute(a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return +1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return +1
	}
	return 0
}

// indexByteBrute returns the index of the first c in s, or -1.
func indexByteBrute(s []byte, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// lastIndexByteBrute returns the index of the last c in s, or -1.
func lastIndexByteBrute(s []byte, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// countByteBrute returns the number of c in s.
func countByteBrute(s []byte, c byte) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			n++
		}
	}
	return n
}

// indexBrute returns the index of the first sep in s, or -1.
func indexBrute(s, sep []byte) int {
	for i := 0; i+len(sep) <= len(s); i++ {
		if string(s[i:i+len(sep)]) == string(sep) {
			return i
		}
	}
	return -1
}

// lastIndexBrute returns the index of the last sep in s, or -1.
func lastIndexBrute(s, sep []byte) int {
	for i := len(s) - len(sep); i >= 0; i-- {
		if string(s[i:i+len(sep)]) == string(sep) {
			return i
		}
	}
	return -1
}

// A cmpCase is one case of the compare table.
type cmpCase struct {
	a, b string
	want int
}

// The compare table gives the result for the pairs that the sweep cannot make.
// A byte pair far apart shows the value of the result, because memcmp gives
// the difference of the bytes on some hosts.
var cmpCases = []cmpCase{
	{"", "", 0},
	{"", "a", -1},
	{"a", "", +1},
	{"a", "a", 0},
	{"a", "b", -1},
	{"b", "a", +1},
	{"a", "z", -1},
	{"z", "a", +1},
	{"abc", "abd", -1},
	{"abd", "abc", +1},
	{"abc", "abcd", -1},
	{"abcd", "abc", +1},
	{"\x7f", "\x80", -1},
	{"\x80", "\x7f", +1},
	{"\x00", "\xff", -1},
	{"\xff", "\x00", +1},
	{"a\x00c", "a\x00d", -1},
	{"hello world", "hello world", 0},
}

func TestCompare(t *testing.T) {
	for i, c := range cmpCases {
		if got := bytealg.Compare([]byte(c.a), []byte(c.b)); got != c.want {
			t.Errorf("Compare() case %d = %d, want %d", i, got, c.want)
		}
	}
}

func TestCompareNil(t *testing.T) {
	// A nil slice is the same as an empty slice.
	empty := []byte{}
	if got := bytealg.Compare(nil, nil); got != 0 {
		t.Errorf("Compare(nil, nil) = %d, want 0", got)
	}
	if got := bytealg.Compare(nil, empty); got != 0 {
		t.Errorf("Compare(nil, empty) = %d, want 0", got)
	}
	if got := bytealg.Compare(nil, []byte("a")); got != -1 {
		t.Errorf("Compare(nil, a) = %d, want -1", got)
	}
	if got := bytealg.Compare([]byte("a"), nil); got != +1 {
		t.Errorf("Compare(a, nil) = %d, want +1", got)
	}
}

func TestCompareSame(t *testing.T) {
	// One slice compares equal with itself, and with a view of itself.
	s := []byte("hello")
	if got := bytealg.Compare(s, s); got != 0 {
		t.Errorf("Compare(s, s) = %d, want 0", got)
	}
	if got := bytealg.Compare(s[:3], s); got != -1 {
		t.Errorf("Compare(s[:3], s) = %d, want -1", got)
	}
	if got := bytealg.Compare(s, s[:3]); got != +1 {
		t.Errorf("Compare(s, s[:3]) = %d, want +1", got)
	}
}

func TestCompareSweep(t *testing.T) {
	for _, sw := range sweeps {
		var abuf, bbuf [maxPair]byte
		total := wordTotal(sw.alpha, maxPair)
		for ai := 0; ai < total; ai++ {
			a := wordAt(abuf[:], sw.alpha, maxPair, ai)
			for bi := 0; bi < total; bi++ {
				b := wordAt(bbuf[:], sw.alpha, maxPair, bi)
				got, want := bytealg.Compare(a, b), compareBrute(a, b)
				if got != want {
					var d1, d2 [2 * maxPair]byte
					t.Errorf("Compare(%s, %s) = %d, want %d",
						dump(d1[:], a), dump(d2[:], b), got, want)
					return
				}
			}
		}
	}
}

func TestCompareLong(t *testing.T) {
	// A long slice takes the vector path of memcmp.
	const n = 300
	var abuf, bbuf [n]byte
	for i := 0; i < n; i++ {
		abuf[i] = byte(i)
		bbuf[i] = byte(i)
	}
	a, b := abuf[:], bbuf[:]
	if got := bytealg.Compare(a, b); got != 0 {
		t.Errorf("Compare(a, a) = %d, want 0", got)
		return
	}
	for i := 0; i < n; i++ {
		// The byte 255 wraps to 0, so the reference gives the wanted sign.
		bbuf[i] = abuf[i] + 1
		if got, want := bytealg.Compare(a, b), compareBrute(a, b); got != want {
			t.Errorf("Compare above byte %d = %d, want %d", i, got, want)
			return
		}
		bbuf[i] = abuf[i] - 1
		if got, want := bytealg.Compare(a, b), compareBrute(a, b); got != want {
			t.Errorf("Compare below byte %d = %d, want %d", i, got, want)
			return
		}
		bbuf[i] = abuf[i]
	}
	// A shorter slice with the same bytes comes first.
	if got := bytealg.Compare(a[:n-1], b); got != -1 {
		t.Errorf("Compare(a[:n-1], b) = %d, want -1", got)
	}
	if got := bytealg.Compare(a, b[:n-1]); got != +1 {
		t.Errorf("Compare(a, b[:n-1]) = %d, want +1", got)
	}
}

func TestEqual(t *testing.T) {
	if !bytealg.Equal(nil, nil) {
		t.Error("Equal(nil, nil) = false")
	}
	if !bytealg.Equal(nil, []byte{}) {
		t.Error("Equal(nil, empty) = false")
	}
	if bytealg.Equal(nil, []byte("a")) {
		t.Error("Equal(nil, a) = true")
	}
	if !bytealg.Equal([]byte("a\x00b"), []byte("a\x00b")) {
		t.Error("Equal(a\\x00b, a\\x00b) = false")
	}
	if bytealg.Equal([]byte("abc"), []byte("abd")) {
		t.Error("Equal(abc, abd) = true")
	}
	if bytealg.Equal([]byte("abc"), []byte("abcd")) {
		t.Error("Equal(abc, abcd) = true")
	}
}

func TestEqualSweep(t *testing.T) {
	// Equal must agree with Compare over every pair of words.
	for _, sw := range sweeps {
		var abuf, bbuf [maxPair]byte
		total := wordTotal(sw.alpha, maxPair)
		for ai := 0; ai < total; ai++ {
			a := wordAt(abuf[:], sw.alpha, maxPair, ai)
			for bi := 0; bi < total; bi++ {
				b := wordAt(bbuf[:], sw.alpha, maxPair, bi)
				got, want := bytealg.Equal(a, b), compareBrute(a, b) == 0
				if got != want {
					var d1, d2 [2 * maxPair]byte
					t.Errorf("Equal(%s, %s) = %t, want %t",
						dump(d1[:], a), dump(d2[:], b), got, want)
					return
				}
			}
		}
	}
}

func TestIndexByte(t *testing.T) {
	for _, sw := range sweeps {
		var buf [maxWord]byte
		total := wordTotal(sw.alpha, sw.maxWord)
		for i := 0; i < total; i++ {
			s := wordAt(buf[:], sw.alpha, sw.maxWord, i)
			for k := 0; k < len(sw.probes); k++ {
				c := sw.probes[k]
				want := indexByteBrute(s, c)
				var d1, d2 [2 * maxWord]byte
				if got := bytealg.IndexByte(s, c); got != want {
					t.Errorf("IndexByte(%s, %s) = %d, want %d",
						dump(d1[:], s), dump(d2[:2], []byte{c}), got, want)
					return
				}
				if got := bytealg.IndexByteString(string(s), c); got != want {
					t.Errorf("IndexByteString(%s, %s) = %d, want %d",
						dump(d1[:], s), dump(d2[:2], []byte{c}), got, want)
					return
				}
			}
		}
	}
}

func TestIndexByteNil(t *testing.T) {
	// An empty slice can have a NULL pointer, which memchr rejects.
	if got := bytealg.IndexByte(nil, 'a'); got != -1 {
		t.Errorf("IndexByte(nil, a) = %d, want -1", got)
	}
	if got := bytealg.IndexByte([]byte{}, 'a'); got != -1 {
		t.Errorf("IndexByte(empty, a) = %d, want -1", got)
	}
	if got := bytealg.IndexByteString("", 'a'); got != -1 {
		t.Errorf("IndexByteString(empty, a) = %d, want -1", got)
	}
	if got := bytealg.LastIndexByte(nil, 'a'); got != -1 {
		t.Errorf("LastIndexByte(nil, a) = %d, want -1", got)
	}
	if got := bytealg.LastIndexByteString("", 'a'); got != -1 {
		t.Errorf("LastIndexByteString(empty, a) = %d, want -1", got)
	}
}

func TestIndexByteLong(t *testing.T) {
	// A long slice takes the vector path of memchr.
	const n = 300
	var buf [n]byte
	for i := 0; i < n; i++ {
		buf[i] = 'a'
	}
	s := buf[:]
	if got := bytealg.IndexByte(s, 'b'); got != -1 {
		t.Errorf("IndexByte(long, b) = %d, want -1", got)
		return
	}
	for i := 0; i < n; i++ {
		buf[i] = 'b'
		if got := bytealg.IndexByte(s, 'b'); got != i {
			t.Errorf("IndexByte at byte %d = %d, want %d", i, got, i)
			return
		}
		if got := bytealg.LastIndexByte(s, 'b'); got != i {
			t.Errorf("LastIndexByte at byte %d = %d, want %d", i, got, i)
			return
		}
		buf[i] = 'a'
	}
}

func TestLastIndexByte(t *testing.T) {
	for _, sw := range sweeps {
		var buf [maxWord]byte
		total := wordTotal(sw.alpha, sw.maxWord)
		for i := 0; i < total; i++ {
			s := wordAt(buf[:], sw.alpha, sw.maxWord, i)
			for k := 0; k < len(sw.probes); k++ {
				c := sw.probes[k]
				want := lastIndexByteBrute(s, c)
				var d1, d2 [2 * maxWord]byte
				if got := bytealg.LastIndexByte(s, c); got != want {
					t.Errorf("LastIndexByte(%s, %s) = %d, want %d",
						dump(d1[:], s), dump(d2[:2], []byte{c}), got, want)
					return
				}
				if got := bytealg.LastIndexByteString(string(s), c); got != want {
					t.Errorf("LastIndexByteString(%s, %s) = %d, want %d",
						dump(d1[:], s), dump(d2[:2], []byte{c}), got, want)
					return
				}
			}
		}
	}
}

func TestCount(t *testing.T) {
	for _, sw := range sweeps {
		var buf [maxWord]byte
		total := wordTotal(sw.alpha, sw.maxWord)
		for i := 0; i < total; i++ {
			s := wordAt(buf[:], sw.alpha, sw.maxWord, i)
			for k := 0; k < len(sw.probes); k++ {
				c := sw.probes[k]
				want := countByteBrute(s, c)
				var d1, d2 [2 * maxWord]byte
				if got := bytealg.Count(s, c); got != want {
					t.Errorf("Count(%s, %s) = %d, want %d",
						dump(d1[:], s), dump(d2[:2], []byte{c}), got, want)
					return
				}
				if got := bytealg.CountString(string(s), c); got != want {
					t.Errorf("CountString(%s, %s) = %d, want %d",
						dump(d1[:], s), dump(d2[:2], []byte{c}), got, want)
					return
				}
			}
		}
	}
}

func TestCountNil(t *testing.T) {
	if got := bytealg.Count(nil, 'a'); got != 0 {
		t.Errorf("Count(nil, a) = %d, want 0", got)
	}
	if got := bytealg.CountString("", 'a'); got != 0 {
		t.Errorf("CountString(empty, a) = %d, want 0", got)
	}
}

func TestIndexRabinKarp(t *testing.T) {
	// The separator must fit s, so the sweep skips the longer separators.
	for _, sw := range sweeps {
		var sbuf, pbuf [maxWord]byte
		sTotal := wordTotal(sw.alpha, sw.maxWord)
		pTotal := wordTotal(sw.alpha, maxSep)
		for si := 0; si < sTotal; si++ {
			s := wordAt(sbuf[:], sw.alpha, sw.maxWord, si)
			for pi := 0; pi < pTotal; pi++ {
				sep := wordAt(pbuf[:], sw.alpha, maxSep, pi)
				if len(sep) > len(s) {
					continue
				}
				got, want := bytealg.IndexRabinKarp(s, sep), indexBrute(s, sep)
				if got != want {
					var d1, d2 [2 * maxWord]byte
					t.Errorf("IndexRabinKarp(%s, %s) = %d, want %d",
						dump(d1[:], s), dump(d2[:], sep), got, want)
					return
				}
			}
		}
	}
}

func TestLastIndexRabinKarp(t *testing.T) {
	for _, sw := range sweeps {
		var sbuf, pbuf [maxWord]byte
		sTotal := wordTotal(sw.alpha, sw.maxWord)
		pTotal := wordTotal(sw.alpha, maxSep)
		for si := 0; si < sTotal; si++ {
			s := wordAt(sbuf[:], sw.alpha, sw.maxWord, si)
			for pi := 0; pi < pTotal; pi++ {
				sep := wordAt(pbuf[:], sw.alpha, maxSep, pi)
				if len(sep) > len(s) {
					continue
				}
				got, want := bytealg.LastIndexRabinKarp(s, sep), lastIndexBrute(s, sep)
				if got != want {
					var d1, d2 [2 * maxWord]byte
					t.Errorf("LastIndexRabinKarp(%s, %s) = %d, want %d",
						dump(d1[:], s), dump(d2[:], sep), got, want)
					return
				}
			}
		}
	}
}

func TestRabinKarpLong(t *testing.T) {
	// A long text with one match, at every position. The text repeats one
	// letter, so the hash of every window but one is the same.
	const n = 300
	var buf [n]byte
	for i := 0; i < n; i++ {
		buf[i] = 'a'
	}
	s := buf[:]
	sep := []byte("aba")
	if got := bytealg.IndexRabinKarp(s, sep); got != -1 {
		t.Errorf("IndexRabinKarp(long, aba) = %d, want -1", got)
		return
	}
	for i := 0; i <= n-len(sep); i++ {
		buf[i+1] = 'b'
		if got := bytealg.IndexRabinKarp(s, sep); got != i {
			t.Errorf("IndexRabinKarp at byte %d = %d, want %d", i, got, i)
			return
		}
		if got := bytealg.LastIndexRabinKarp(s, sep); got != i {
			t.Errorf("LastIndexRabinKarp at byte %d = %d, want %d", i, got, i)
			return
		}
		buf[i+1] = 'a'
	}
}

func TestHashStrRev(t *testing.T) {
	// HashStrRev gives the hash of the reversed word and the factor
	// PrimeRK to the power of the length.
	for _, sw := range sweeps {
		var buf [maxWord]byte
		total := wordTotal(sw.alpha, sw.maxWord)
		for i := 0; i < total; i++ {
			s := wordAt(buf[:], sw.alpha, sw.maxWord, i)
			var wantHash, wantPow uint32 = 0, 1
			for k := len(s) - 1; k >= 0; k-- {
				wantHash = wantHash*bytealg.PrimeRK + uint32(s[k])
			}
			for k := 0; k < len(s); k++ {
				wantPow *= bytealg.PrimeRK
			}
			hash, pow := bytealg.HashStrRev(s)
			if hash != wantHash || pow != wantPow {
				var d [2 * maxWord]byte
				t.Errorf("HashStrRev(%s) = %x, %x, want %x, %x",
					dump(d[:], s), hash, pow, wantHash, wantPow)
				return
			}
		}
	}
}
