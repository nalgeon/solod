package bytes_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
	"solod.dev/so/unicode/utf8"
)

// The errors of the reader and buffer tests, by number. A table holds the
// number instead of the error, because a package level table must not depend
// on the initialization order of the error values.
const (
	errNone = iota
	errEOF
	errOffset
	errWhence
	errUnread
	errNegativeRead
	errOther
)

// errCode returns the number of the error.
func errCode(err error) int {
	if err == nil {
		return errNone
	}
	if err == io.EOF {
		return errEOF
	}
	if err == io.ErrOffset {
		return errOffset
	}
	if err == io.ErrWhence {
		return errWhence
	}
	if err == io.ErrUnread {
		return errUnread
	}
	if err == io.ErrNegativeRead {
		return errNegativeRead
	}
	return errOther
}

// errName returns the name of the error with the number.
func errName(code int) string {
	switch code {
	case errNone:
		return "nil"
	case errEOF:
		return "EOF"
	case errOffset:
		return "ErrOffset"
	case errWhence:
		return "ErrWhence"
	case errUnread:
		return "ErrUnread"
	case errNegativeRead:
		return "ErrNegativeRead"
	}
	return "other"
}

// alphabet holds the lower case letters.
const alphabet = "abcdefghijklmnopqrstuvwxyz"

// testData writes the test data into buf and returns the result. The data is
// the letters of the alphabet, repeated.
func testData(buf []byte) []byte {
	for i := range buf {
		buf[i] = alphabet[i%len(alphabet)]
	}
	return buf
}

// nextRand returns the next value of a pseudo random sequence.
func nextRand(x uint32) uint32 {
	return x*1664525 + 1013904223
}

// checkBuffer verifies that the unread part of the buffer is s.
func checkBuffer(t *testing.T, name string, b *bytes.Buffer, s string) {
	if b.Len() != len(s) {
		t.Errorf("%s: Len() = %d, want %d", name, b.Len(), len(s))
		return
	}
	if len(b.Bytes()) != len(s) {
		t.Errorf("%s: len(Bytes()) = %d, want %d", name, len(b.Bytes()), len(s))
		return
	}
	if string(b.Bytes()) != s {
		t.Errorf("%s: Bytes() = %s, want %s", name, string(b.Bytes()), s)
	}
	if b.String() != s {
		t.Errorf("%s: String() = %s, want %s", name, b.String(), s)
	}
}

func TestBufferNew(t *testing.T) {
	// NewBuffer takes the slice as the contents of the buffer.
	var arr [11]byte
	s := arr[:]
	copy(s, "hello world")
	b := bytes.NewBuffer(nil, s)
	checkBuffer(t, "NewBuffer", &b, "hello world")
	if b.Cap() != 11 {
		t.Errorf("Cap() = %d, want 11", b.Cap())
	}
	if b.Available() != 0 {
		t.Errorf("Available() = %d, want 0", b.Available())
	}
}

func TestBufferNewString(t *testing.T) {
	b := bytes.NewBufferString(nil, "hello world")
	checkBuffer(t, "NewBufferString", &b, "hello world")
}

func TestBufferZero(t *testing.T) {
	// The zero buffer is empty and ready to use.
	var b bytes.Buffer
	defer b.Free()
	checkBuffer(t, "zero buffer", &b, "")
	if b.Cap() != 0 {
		t.Errorf("Cap() = %d, want 0", b.Cap())
	}
	b.WriteString("abc")
	checkBuffer(t, "after WriteString", &b, "abc")
}

func TestBufferNil(t *testing.T) {
	// String of a nil buffer names the nil pointer.
	var b *bytes.Buffer
	if b.String() != "<nil>" {
		t.Errorf("String() of a nil buffer = %s, want <nil>", b.String())
	}
}

func TestBufferBasic(t *testing.T) {
	alloc := t.Allocator()
	b := bytes.NewBuffer(alloc, nil)
	defer b.Free()
	for range 5 {
		checkBuffer(t, "at the start", &b, "")

		b.Reset()
		checkBuffer(t, "after Reset", &b, "")

		n, err := b.Write([]byte("a"))
		if n != 1 || err != nil {
			t.Errorf("Write() = %d, %s, want 1, nil", n, errName(errCode(err)))
		}
		checkBuffer(t, "after Write", &b, "a")

		b.WriteByte('b')
		checkBuffer(t, "after WriteByte", &b, "ab")

		n, err = b.WriteString("cdefghijklmnopqrstuvwxyz")
		if n != 24 || err != nil {
			t.Errorf("WriteString() = %d, %s, want 24, nil", n, errName(errCode(err)))
		}
		checkBuffer(t, "after WriteString", &b, alphabet)

		b.Reset()
		checkBuffer(t, "after the second Reset", &b, "")

		b.WriteByte('b')
		c, err := b.ReadByte()
		if c != 'b' || err != nil {
			t.Errorf("ReadByte() = %d, %s, want 98, nil", c, errName(errCode(err)))
		}
		c, err = b.ReadByte()
		if c != 0 || err != io.EOF {
			t.Errorf("ReadByte() at the end = %d, %s, want 0, EOF",
				c, errName(errCode(err)))
		}
	}
}

func TestBufferLargeWrites(t *testing.T) {
	// A write longer than the free space grows the buffer.
	var data [1000]byte
	src := testData(data[:])
	alloc := t.Allocator()
	b := bytes.NewBuffer(alloc, nil)
	defer b.Free()
	for i := range 5 {
		n, err := b.Write(src)
		if n != len(src) || err != nil {
			t.Errorf("Write() = %d, %s, want %d, nil",
				n, errName(errCode(err)), len(src))
			return
		}
		if b.Len() != (i+1)*len(src) {
			t.Errorf("Len() = %d after %d writes, want %d",
				b.Len(), i+1, (i+1)*len(src))
			return
		}
	}
	got := b.Bytes()
	for i, c := range got {
		if c != src[i%len(src)] {
			t.Errorf("byte %d = %d, want %d", i, c, src[i%len(src)])
			return
		}
	}
}

func TestBufferLargeReads(t *testing.T) {
	// A read shorter than the contents gives the bytes in order.
	var data [1000]byte
	src := testData(data[:])
	alloc := t.Allocator()
	b := bytes.NewBuffer(alloc, nil)
	defer b.Free()
	for range 5 {
		b.Write(src)
	}
	var chunk [333]byte
	read := 0
	for {
		n, err := b.Read(chunk[:])
		if n == 0 {
			if err != io.EOF {
				t.Errorf("Read() at the end error = %s, want EOF",
					errName(errCode(err)))
			}
			break
		}
		if err != nil {
			t.Errorf("Read() error = %s", errName(errCode(err)))
			return
		}
		for k := range n {
			if chunk[k] != src[(read+k)%len(src)] {
				t.Errorf("byte %d = %d, want %d",
					read+k, chunk[k], src[(read+k)%len(src)])
				return
			}
		}
		read += n
	}
	if read != 5*len(src) {
		t.Errorf("the reads gave %d bytes, want %d", read, 5*len(src))
	}
	checkBuffer(t, "after the reads", &b, "")
}

func TestBufferMixed(t *testing.T) {
	// Writes and reads of changing sizes keep the contents in order.
	var data [500]byte
	src := testData(data[:])
	var chunk [500]byte
	alloc := t.Allocator()
	b := bytes.NewBuffer(alloc, nil)
	defer b.Free()
	rnd := uint32(1)
	written, read := 0, 0
	for i := range 50 {
		rnd = nextRand(rnd)
		wlen := int(rnd>>8) % len(src)
		b.Write(src[:wlen])
		written += wlen

		rnd = nextRand(rnd)
		rlen := int(rnd>>8) % len(chunk)
		n, _ := b.Read(chunk[:rlen])
		read += n
		if b.Len() != written-read {
			t.Errorf("round %d: Len() = %d, want %d", i, b.Len(), written-read)
			return
		}
	}
}

func TestBufferCap(t *testing.T) {
	// Cap is the size of the underlying slice, and Available is the free
	// space of that slice.
	var arr [10]byte
	b := bytes.NewBuffer(nil, arr[:])
	if b.Cap() != 10 {
		t.Errorf("Cap() = %d, want 10", b.Cap())
	}
	if b.Available() != 0 {
		t.Errorf("Available() = %d, want 0", b.Available())
	}

	b2 := bytes.NewBuffer(nil, arr[:0])
	b2.Write([]byte("test"))
	if b2.Cap() != 10 {
		t.Errorf("Cap() = %d after a write, want 10", b2.Cap())
	}
	if b2.Available() != 6 {
		t.Errorf("Available() = %d after a write, want 6", b2.Available())
	}
}

func TestBufferGrow(t *testing.T) {
	// Grow gives space for n more bytes and keeps the contents.
	alloc := t.Allocator()
	b := bytes.NewBuffer(alloc, nil)
	defer b.Free()
	b.WriteString("hello")
	b.Grow(100)
	checkBuffer(t, "after Grow", &b, "hello")
	if b.Available() < 100 {
		t.Errorf("Available() = %d after Grow(100), want 100 or more", b.Available())
	}
	// A write up to the new space does not move the contents.
	before := b.Cap()
	var data [100]byte
	b.Write(testData(data[:]))
	if b.Cap() != before {
		t.Errorf("Cap() = %d after the write, want %d", b.Cap(), before)
	}
	// Grow(0) changes nothing.
	b.Grow(0)
	if b.Len() != 105 {
		t.Errorf("Len() = %d after Grow(0), want 105", b.Len())
	}
}

func TestBufferGrowth(t *testing.T) {
	// A buffer that is written and read again and again slides the contents
	// down instead of growing without an end.
	var data [1024]byte
	src := testData(data[:])
	var chunk [1024]byte
	alloc := t.Allocator()
	b := bytes.NewBuffer(alloc, nil)
	defer b.Free()
	b.Write(src[:1])
	cap0 := 0
	for i := range 1024 {
		b.Write(src)
		b.Read(chunk[:])
		if i == 0 {
			cap0 = b.Cap()
		}
	}
	// grow allows twice the capacity before it slides the contents, so the
	// test accepts three times the first capacity.
	if b.Cap() > cap0*3 {
		t.Errorf("Cap() = %d, too big (it grew from %d)", b.Cap(), cap0)
	}
}

func TestBufferReset(t *testing.T) {
	// Reset empties the buffer and keeps the space.
	alloc := t.Allocator()
	b := bytes.NewBuffer(alloc, nil)
	defer b.Free()
	b.WriteString("hello world")
	before := b.Cap()
	b.Reset()
	checkBuffer(t, "after Reset", &b, "")
	if b.Cap() != before {
		t.Errorf("Cap() = %d after Reset, want %d", b.Cap(), before)
	}
}

func TestBufferFree(t *testing.T) {
	// Free releases the space, and the buffer works again after Free.
	alloc := t.Allocator()
	b := bytes.NewBuffer(alloc, nil)
	b.WriteString("hello world")
	b.Free()
	checkBuffer(t, "after Free", &b, "")
	if b.Cap() != 0 {
		t.Errorf("Cap() = %d after Free, want 0", b.Cap())
	}
	b.WriteString("again")
	checkBuffer(t, "after the second write", &b, "again")
	b.Free()
}

func TestBufferReadEmptyAtEOF(t *testing.T) {
	// A read of no bytes from an empty buffer is not an error.
	var b bytes.Buffer
	var arr [1]byte
	n, err := b.Read(arr[:0])
	if n != 0 || err != nil {
		t.Errorf("Read() = %d, %s, want 0, nil", n, errName(errCode(err)))
	}
}

func TestBufferNext(t *testing.T) {
	// Next gives the next n bytes, or the rest of the contents.
	var arr [5]byte
	for i := range arr {
		arr[i] = byte(i)
	}
	var tmp [5]byte
	for i := 0; i <= 5; i++ {
		for j := i; j <= 5; j++ {
			for k := 0; k <= 6; k++ {
				b := bytes.NewBuffer(nil, arr[0:j])
				n, _ := b.Read(tmp[0:i])
				if n != i {
					t.Errorf("Read(%d) = %d", i, n)
					return
				}
				got := b.Next(k)
				want := min(k, j-i)
				if len(got) != want {
					t.Errorf("i=%d j=%d: len(Next(%d)) = %d, want %d",
						i, j, k, len(got), want)
					return
				}
				for l, v := range got {
					if v != byte(l+i) {
						t.Errorf("i=%d j=%d: Next(%d)[%d] = %d, want %d",
							i, j, k, l, v, l+i)
						return
					}
				}
			}
		}
	}
}

// A readBytesCase is a test case of ReadBytes and ReadString.
type readBytesCase struct {
	buffer string
	delim  byte
	want   string // the wanted lines, joined with partSep
	parts  int
	err    int // the error of the last read
}

var readBytesCases = []readBytesCase{
	{"", 0, "", 1, errEOF},
	{"a\x00", 0, "a\x00", 1, errNone},
	{"abbbaaaba", 'b', "ab\x02b\x02b\x02aaab", 4, errNone},
	{"hello\x01world", 1, "hello\x01", 1, errNone},
	{"foo\nbar", 0, "foo\nbar", 1, errEOF},
	{"alpha\nbeta\ngamma\n", '\n', "alpha\n\x02beta\n\x02gamma\n", 3, errNone},
	{"alpha\nbeta\ngamma", '\n', "alpha\n\x02beta\n\x02gamma", 3, errEOF},
}

func TestBufferReadBytes(t *testing.T) {
	alloc := t.Allocator()
	for i, tc := range readBytesCases {
		b := bytes.NewBufferString(alloc, tc.buffer)
		code := errNone
		for n := 0; n < tc.parts; n++ {
			line, err := b.ReadBytes(tc.delim)
			if string(line) != partAt(tc.want, n) {
				t.Errorf("case %d: line %d = %s, want %s",
					i, n, string(line), partAt(tc.want, n))
			}
			mem.FreeSlice(alloc, line)
			code = errCode(err)
			if err != nil {
				break
			}
		}
		if code != tc.err {
			t.Errorf("case %d: error = %s, want %s",
				i, errName(code), errName(tc.err))
		}
	}
}

func TestBufferReadString(t *testing.T) {
	alloc := t.Allocator()
	for i, tc := range readBytesCases {
		b := bytes.NewBufferString(alloc, tc.buffer)
		code := errNone
		for n := 0; n < tc.parts; n++ {
			line, err := b.ReadString(tc.delim)
			if line != partAt(tc.want, n) {
				t.Errorf("case %d: line %d = %s, want %s",
					i, n, line, partAt(tc.want, n))
			}
			mem.FreeString(alloc, line)
			code = errCode(err)
			if err != nil {
				break
			}
		}
		if code != tc.err {
			t.Errorf("case %d: error = %s, want %s",
				i, errName(code), errName(tc.err))
		}
	}
}

// A peekCase is a test case of Peek.
type peekCase struct {
	buffer string
	skip   int
	n      int
	want   string
	err    int
}

var peekCases = []peekCase{
	{"", 0, 0, "", errNone},
	{"aaa", 0, 3, "aaa", errNone},
	{"foobar", 0, 2, "fo", errNone},
	{"a", 0, 2, "a", errEOF},
	{"helloworld", 4, 3, "owo", errNone},
	{"helloworld", 5, 5, "world", errNone},
	{"helloworld", 5, 6, "world", errEOF},
	{"helloworld", 10, 1, "", errEOF},
}

func TestBufferPeek(t *testing.T) {
	// Peek gives the next n bytes and does not advance the buffer.
	for i, tc := range peekCases {
		b := bytes.NewBufferString(nil, tc.buffer)
		b.Next(tc.skip)
		got, err := b.Peek(tc.n)
		if string(got) != tc.want {
			t.Errorf("case %d: Peek() = %s, want %s", i, string(got), tc.want)
		}
		if errCode(err) != tc.err {
			t.Errorf("case %d: Peek() error = %s, want %s",
				i, errName(errCode(err)), errName(tc.err))
		}
		if b.Len() != len(tc.buffer)-tc.skip {
			t.Errorf("case %d: Len() = %d after Peek, want %d",
				i, b.Len(), len(tc.buffer)-tc.skip)
		}
	}
}

func TestBufferReadFrom(t *testing.T) {
	alloc := t.Allocator()
	var data [300]byte
	src := testData(data[:])
	for i := 3; i < 30; i += 3 {
		n := len(src) / i
		r := bytes.NewReader(src[:n])
		b := bytes.NewBuffer(alloc, nil)
		got, err := b.ReadFrom(&r)
		if got != int64(n) || err != nil {
			t.Errorf("ReadFrom() = %d, %s, want %d, nil",
				got, errName(errCode(err)), n)
		}
		if !bytes.Equal(b.Bytes(), src[:n]) {
			t.Errorf("ReadFrom() read the wrong bytes for the length %d", n)
		}
		b.Free()
	}
}

// A negativeReader gives a negative count from Read, which no reader may do.
type negativeReader struct{}

func (*negativeReader) Read(p []byte) (int, error) {
	_ = p
	return -1, nil
}

func TestBufferReadFromNegative(t *testing.T) {
	// A negative count from the reader gives ErrNegativeRead.
	alloc := t.Allocator()
	b := bytes.NewBuffer(alloc, nil)
	defer b.Free()
	var r negativeReader
	n, err := b.ReadFrom(&r)
	if err != io.ErrNegativeRead {
		t.Errorf("ReadFrom() error = %s, want ErrNegativeRead", errName(errCode(err)))
	}
	if n != 0 {
		t.Errorf("ReadFrom() = %d, want 0", n)
	}
}

func TestBufferWriteTo(t *testing.T) {
	alloc := t.Allocator()
	var data [300]byte
	src := testData(data[:])
	for i := 3; i < 30; i += 3 {
		n := len(src) / i
		b := bytes.NewBuffer(alloc, nil)
		b.Write(src[:n])
		w := bytes.NewBuffer(alloc, nil)
		got, err := b.WriteTo(&w)
		if got != int64(n) || err != nil {
			t.Errorf("WriteTo() = %d, %s, want %d, nil",
				got, errName(errCode(err)), n)
		}
		if !bytes.Equal(w.Bytes(), src[:n]) {
			t.Errorf("WriteTo() wrote the wrong bytes for the length %d", n)
		}
		if b.Len() != 0 {
			t.Errorf("Len() = %d after WriteTo, want 0", b.Len())
		}
		w.Free()
		b.Free()
	}
}

// The code points of the WriteRune test. The list holds the first and the last
// code point of every width of the UTF-8 encoding.
var runeCases = []rune{
	0, 'a', 0x7f, 0x80, 0x7ff, 0x800, 0xfffd, 0xffff, 0x10000, utf8.MaxRune,
}

func TestBufferRuneIO(t *testing.T) {
	// WriteRune writes the UTF-8 encoding, and ReadRune gives the code point
	// back.
	alloc := t.Allocator()
	b := bytes.NewBuffer(alloc, nil)
	defer b.Free()
	total := 0
	for _, r := range runeCases {
		n, err := b.WriteRune(r)
		if err != nil {
			t.Errorf("WriteRune(%x) error = %s", r, errName(errCode(err)))
			return
		}
		if n != utf8.RuneLen(r) {
			t.Errorf("WriteRune(%x) = %d, want %d", r, n, utf8.RuneLen(r))
		}
		total += n
	}
	if b.Len() != total {
		t.Errorf("Len() = %d, want %d", b.Len(), total)
	}
	for _, r := range runeCases {
		res := b.ReadRune()
		if res.Err != nil {
			t.Errorf("ReadRune() error = %s", errName(errCode(res.Err)))
			return
		}
		if res.Rune != r || res.Size != utf8.RuneLen(r) {
			t.Errorf("ReadRune() = %x, %d, want %x, %d",
				res.Rune, res.Size, r, utf8.RuneLen(r))
		}
	}
	res := b.ReadRune()
	if res.Err != io.EOF {
		t.Errorf("ReadRune() at the end error = %s, want EOF", errName(errCode(res.Err)))
	}
}

func TestBufferWriteInvalidRune(t *testing.T) {
	// An invalid code point becomes RuneError, and a negative code point too.
	alloc := t.Allocator()
	bad := []rune{-1, 0xd800, utf8.MaxRune + 1}
	for _, r := range bad {
		b := bytes.NewBuffer(alloc, nil)
		n, err := b.WriteRune(r)
		if err != nil {
			t.Errorf("WriteRune(%x) error = %s", r, errName(errCode(err)))
			b.Free()
			return
		}
		if n != 3 {
			t.Errorf("WriteRune(%x) = %d, want 3", r, n)
		}
		checkBuffer(t, "after WriteRune", &b, "�")
		b.Free()
	}
}

func TestBufferReadRuneInvalid(t *testing.T) {
	// An invalid byte of the contents gives RuneError with the size 1.
	b := bytes.NewBufferString(nil, "\xffa")
	res := b.ReadRune()
	if res.Rune != utf8.RuneError || res.Size != 1 || res.Err != nil {
		t.Errorf("ReadRune() = %x, %d, want fffd, 1", res.Rune, res.Size)
	}
	res = b.ReadRune()
	if res.Rune != 'a' || res.Size != 1 {
		t.Errorf("ReadRune() = %x, %d, want 61, 1", res.Rune, res.Size)
	}
}

func TestBufferBytesView(t *testing.T) {
	// Bytes gives a view of the contents, not a copy.
	alloc := t.Allocator()
	b := bytes.NewBuffer(alloc, nil)
	defer b.Free()
	b.WriteString("hello")
	got := b.Bytes()
	got[0] = 'H'
	if b.String() != "Hello" {
		t.Errorf("Bytes() gives a copy: String() = %s", b.String())
	}
}
