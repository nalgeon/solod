// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bufio_test

import (
	"solod.dev/so/bufio"
	"solod.dev/so/bytes"
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
	"solod.dev/so/unicode/utf8"
)

// readFunc reads the whole content of a reader into dst.
// It returns the number of the read bytes.
// The m argument is the size of the chunks for the readers that read chunks.
type readFunc func(a mem.Allocator, b *bufio.Reader, m int, dst []byte) int

// readChunks reads chunks of m bytes.
func readChunks(a mem.Allocator, b *bufio.Reader, m int, dst []byte) int {
	_ = a
	nb := 0
	for {
		n, err := b.Read(dst[nb : nb+m])
		nb += n
		if err == io.EOF {
			break
		}
		if err != nil && err != errTimeoutFake {
			break
		}
	}
	return nb
}

// readBytes reads one byte at a time.
func readBytes(a mem.Allocator, b *bufio.Reader, m int, dst []byte) int {
	_, _ = a, m
	nb := 0
	for {
		c, err := b.ReadByte()
		if err == io.EOF {
			break
		}
		if err == nil {
			dst[nb] = c
			nb++
			continue
		}
		if err != errTimeoutFake {
			break
		}
	}
	return nb
}

// readLines reads one line at a time. ReadString calls everything else.
func readLines(a mem.Allocator, b *bufio.Reader, m int, dst []byte) int {
	_ = m
	nb := 0
	for {
		s, err := b.ReadString('\n')
		nb += copy(dst[nb:], []byte(s))
		mem.FreeString(a, s)
		if err == io.EOF {
			break
		}
		if err != nil && err != errTimeoutFake {
			break
		}
	}
	return nb
}

// readMaker names a wrapping mode of the reader under test.
type readMaker struct {
	name string
	mode int
}

var readMakers = []readMaker{
	{"full", wrapFull},
	{"byte", wrapByte},
	{"half", wrapHalf},
	{"data+err", wrapDataErr},
	{"timeout", wrapTimeout},
}

// bufReader names a way to read the whole content of a reader.
type bufReader struct {
	name string
	m    int
	fn   readFunc
}

var bufReaders = []bufReader{
	{"1", 1, readChunks},
	{"2", 2, readChunks},
	{"3", 3, readChunks},
	{"4", 4, readChunks},
	{"5", 5, readChunks},
	{"7", 7, readChunks},
	{"bytes", 0, readBytes},
	{"lines", 0, readLines},
}

var bufSizes = []int{
	0, minReadBufferSize, 23, 32, 46, 64, 93, 128, 1024, 4096,
}

func TestReaderSimple(t *testing.T) {
	alloc := t.Allocator()
	data := "hello world"
	var dst [32]byte

	sr := strings.NewReader(data)
	b := bufio.NewReader(alloc, &sr)
	n := readBytes(alloc, &b, 0, dst[:])
	if string(dst[:n]) != "hello world" {
		t.Errorf("simple: got %s, want hello world", string(dst[:n]))
	}
	b.Free()

	sr = strings.NewReader(data)
	r13 := rot13Reader{r: &sr}
	b = bufio.NewReader(alloc, &r13)
	n = readBytes(alloc, &b, 0, dst[:])
	if string(dst[:n]) != "uryyb jbeyq" {
		t.Errorf("rot13: got %s, want uryyb jbeyq", string(dst[:n]))
	}
	b.Free()
}

func TestReader(t *testing.T) {
	// Read texts of a growing size with every wrapping mode,
	// every way to read, and every buffer size.
	const arenaSize = 16 * 1024
	alloc := t.Allocator()
	backing := mem.AllocSlice[byte](alloc, arenaSize, arenaSize)
	arena := mem.NewArena(backing)

	// texts[i] holds i letters and a newline. The last text holds all of them.
	var all [465]byte
	var texts [31]string
	off := 0
	for i := range len(texts) - 1 {
		for j := 0; j < i; j++ {
			all[off+j] = byte(j%26 + 'a')
		}
		all[off+i] = '\n'
		texts[i] = string(all[off : off+i+1])
		off += i + 1
	}
	texts[len(texts)-1] = string(all[:])

	var dst [1000]byte
	for h := range len(texts) {
		text := texts[h]
		for i := range readMakers {
			for j := range bufReaders {
				for k := range bufSizes {
					mk := readMakers[i]
					br := bufReaders[j]
					size := bufSizes[k]
					sr := strings.NewReader(text)
					wrap := wrapReader{r: &sr, mode: mk.mode}
					b := bufio.NewReaderSize(&arena, &wrap, size)
					n := br.fn(&arena, &b, br.m, dst[:])
					if string(dst[:n]) != text {
						t.Errorf("reader=%s fn=%s bufsize=%d text=%d: got %s, want %s",
							mk.name, br.name, size, h, string(dst[:n]), text)
					}
					b.Free()
					arena.Reset()
				}
			}
		}
	}

	mem.FreeSlice(alloc, backing)
}

func TestZeroReader(t *testing.T) {
	var z zeroReader
	r := bufio.NewReader(t.Allocator(), &z)
	defer r.Free()

	_, err := r.ReadByte()
	if errCode(err) != errNoProgress {
		t.Errorf("ReadByte() = %s, want ErrNoProgress", errName(errCode(err)))
	}
}

// readRuneSegments reads the segments rune by rune and compares them to want.
func readRuneSegments(t *testing.T, segments []string, want string) {
	sr := stringReader{data: segments}
	b := bufio.NewReader(t.Allocator(), &sr)
	defer b.Free()

	var got [64]byte
	n := 0
	for {
		res := b.ReadRune()
		if res.Err != nil {
			if res.Err != io.EOF {
				t.Errorf("ReadRune() = %s, want EOF", errName(errCode(res.Err)))
				return
			}
			break
		}
		n += utf8.EncodeRune(got[n:], res.Rune)
	}
	if string(got[:n]) != want {
		t.Errorf("segments = %s, want %s", string(got[:n]), want)
	}
}

func TestReadRune(t *testing.T) {
	var segs [4]string

	readRuneSegments(t, segs[:0], "")

	segs[0] = ""
	readRuneSegments(t, segs[:1], "")

	segs[0], segs[1] = "日", "本語"
	readRuneSegments(t, segs[:2], "日本語")

	segs[0], segs[1], segs[2] = "日", "本", "語"
	readRuneSegments(t, segs[:3], "日本語")

	// The runes are split between the segments.
	segs[0], segs[1], segs[2] = "\xe6", "\x97\xa5\xe6", "\x9c\xac\xe8\xaa\x9e"
	readRuneSegments(t, segs[:3], "日本語")

	segs[0], segs[1], segs[2], segs[3] = "Hello", ", ", "World", "!"
	readRuneSegments(t, segs[:4], "Hello, World!")

	segs[0], segs[1], segs[2], segs[3] = "Hello", "", ", World", "!"
	readRuneSegments(t, segs[:4], "Hello, World!")
}

func TestUnreadRune(t *testing.T) {
	var segs [2]string
	segs[0], segs[1] = "Hello, world:", "日本語"
	sr := stringReader{data: segs[:]}
	r := bufio.NewReader(t.Allocator(), &sr)
	defer r.Free()

	var got [32]byte
	n := 0
	for {
		r1 := r.ReadRune()
		if r1.Err != nil {
			if r1.Err != io.EOF {
				t.Errorf("ReadRune() = %s, want EOF", errName(errCode(r1.Err)))
			}
			break
		}
		n += utf8.EncodeRune(got[n:], r1.Rune)
		// Put the rune back and read it again.
		if err := r.UnreadRune(); err != nil {
			t.Errorf("UnreadRune() = %s, want nil", errName(errCode(err)))
			return
		}
		r2 := r.ReadRune()
		if r2.Err != nil {
			t.Errorf("ReadRune() after UnreadRune = %s, want nil", errName(errCode(r2.Err)))
			return
		}
		if r1.Rune != r2.Rune {
			t.Errorf("rune after unread = %c, want %c", r2.Rune, r1.Rune)
			return
		}
	}
	if string(got[:n]) != "Hello, world:日本語" {
		t.Errorf("got %s, want Hello, world:日本語", string(got[:n]))
	}
}

func TestNoUnreadRuneAfterPeek(t *testing.T) {
	sr := strings.NewReader("example")
	br := bufio.NewReader(t.Allocator(), &sr)
	defer br.Free()

	br.ReadRune()
	br.Peek(1)
	if err := br.UnreadRune(); err == nil {
		t.Error("UnreadRune() = nil after Peek, want an error")
	}
}

func TestNoUnreadByteAfterPeek(t *testing.T) {
	sr := strings.NewReader("example")
	br := bufio.NewReader(t.Allocator(), &sr)
	defer br.Free()

	br.ReadByte()
	br.Peek(1)
	if err := br.UnreadByte(); err == nil {
		t.Error("UnreadByte() = nil after Peek, want an error")
	}
}

func TestNoUnreadRuneAfterDiscard(t *testing.T) {
	sr := strings.NewReader("example")
	br := bufio.NewReader(t.Allocator(), &sr)
	defer br.Free()

	br.ReadRune()
	br.Discard(1)
	if err := br.UnreadRune(); err == nil {
		t.Error("UnreadRune() = nil after Discard, want an error")
	}
}

func TestNoUnreadByteAfterDiscard(t *testing.T) {
	sr := strings.NewReader("example")
	br := bufio.NewReader(t.Allocator(), &sr)
	defer br.Free()

	br.ReadByte()
	br.Discard(1)
	if err := br.UnreadByte(); err == nil {
		t.Error("UnreadByte() = nil after Discard, want an error")
	}
}

func TestNoUnreadRuneAfterWriteTo(t *testing.T) {
	sr := strings.NewReader("example")
	br := bufio.NewReader(t.Allocator(), &sr)
	defer br.Free()

	var w io.DiscardWriter
	br.WriteTo(&w)
	if err := br.UnreadRune(); err == nil {
		t.Error("UnreadRune() = nil after WriteTo, want an error")
	}
}

func TestNoUnreadByteAfterWriteTo(t *testing.T) {
	sr := strings.NewReader("example")
	br := bufio.NewReader(t.Allocator(), &sr)
	defer br.Free()

	var w io.DiscardWriter
	br.WriteTo(&w)
	if err := br.UnreadByte(); err == nil {
		t.Error("UnreadByte() = nil after WriteTo, want an error")
	}
}

func TestUnreadByte(t *testing.T) {
	var segs [2]string
	segs[0], segs[1] = "Hello, ", "world"
	sr := stringReader{data: segs[:]}
	r := bufio.NewReader(t.Allocator(), &sr)
	defer r.Free()

	var got [32]byte
	n := 0
	for {
		b1, err := r.ReadByte()
		if err != nil {
			if err != io.EOF {
				t.Errorf("ReadByte() = %s, want EOF", errName(errCode(err)))
			}
			break
		}
		got[n] = b1
		n++
		// Put the byte back and read it again.
		if err := r.UnreadByte(); err != nil {
			t.Errorf("UnreadByte() = %s, want nil", errName(errCode(err)))
			return
		}
		b2, err := r.ReadByte()
		if err != nil {
			t.Errorf("ReadByte() after UnreadByte = %s, want nil", errName(errCode(err)))
			return
		}
		if b1 != b2 {
			t.Errorf("byte after unread = %c, want %c", b2, b1)
			return
		}
	}
	if string(got[:n]) != "Hello, world" {
		t.Errorf("got %s, want Hello, world", string(got[:n]))
	}
}

func TestUnreadByteMultiple(t *testing.T) {
	const data = "Hello, world"
	var segs [2]string
	segs[0], segs[1] = "Hello, ", "world"

	for n := 0; n <= len(data); n++ {
		sr := stringReader{data: segs[:]}
		r := bufio.NewReader(t.Allocator(), &sr)

		// Read n bytes.
		for i := 0; i < n; i++ {
			b, err := r.ReadByte()
			if err != nil {
				t.Errorf("n = %d: ReadByte() = %s, want nil", n, errName(errCode(err)))
				break
			}
			if b != data[i] {
				t.Errorf("n = %d: ReadByte() = %c, want %c", n, b, data[i])
				break
			}
		}
		// Unread one byte if there is one.
		if n > 0 {
			if err := r.UnreadByte(); err != nil {
				t.Errorf("n = %d: UnreadByte() = %s, want nil", n, errName(errCode(err)))
			}
		}
		// The reader unreads one byte only.
		if err := r.UnreadByte(); err == nil {
			t.Errorf("n = %d: second UnreadByte() = nil, want an error", n)
		}
		r.Free()
	}
}

// readDelim reads until the delimiter and returns the read bytes.
type readDelim func(r *bufio.Reader, delim byte) ([]byte, error)

// readStringAsBytes reads a string until the delimiter and returns its bytes.
func readStringAsBytes(r *bufio.Reader, delim byte) ([]byte, error) {
	data, err := r.ReadString(delim)
	return []byte(data), err
}

func TestUnreadByteOthers(t *testing.T) {
	// Read with the other read methods and unread a byte between the reads.
	// ReadLine has no test here: it calls ReadSlice, and ReadSlice handles
	// the last byte.
	readers := []readDelim{
		(*bufio.Reader).ReadBytes,
		(*bufio.Reader).ReadSlice,
		readStringAsBytes,
	}

	const arenaSize = 2048
	alloc := t.Allocator()
	backing := mem.AllocSlice[byte](alloc, arenaSize, arenaSize)
	arena := mem.NewArena(backing)

	// The input is longer than the minimum buffer of a reader.
	const n = 10
	var data [n * 7]byte
	for i := range n {
		copy(data[i*7:], []byte("abcdefg"))
	}

	for rno := range readers {
		read := readers[rno]
		br := bytes.NewReader(data[:])
		r := bufio.NewReaderSize(&arena, &br, minReadBufferSize)

		// Read the data with occasional UnreadByte calls.
		for range n {
			readTo(t, read, &r, rno, 'd', "abcd")
			for range 3 {
				if err := r.UnreadByte(); err != nil {
					t.Errorf("#%d: UnreadByte() = %s, want nil", rno, errName(errCode(err)))
					return
				}
				readTo(t, read, &r, rno, 'd', "d")
			}
			readTo(t, read, &r, rno, 'g', "efg")
		}

		// The reader read all the data.
		_, err := r.ReadByte()
		if errCode(err) != errEOF {
			t.Errorf("#%d: ReadByte() = %s, want EOF", rno, errName(errCode(err)))
		}
		arena.Reset()
	}

	mem.FreeSlice(alloc, backing)
}

// readTo reads until the delimiter and compares the result to want.
func readTo(t *testing.T, read readDelim, r *bufio.Reader, rno int, delim byte, want string) {
	data, err := read(r, delim)
	if err != nil {
		t.Errorf("#%d: read to %c = %s, want nil", rno, delim, errName(errCode(err)))
		return
	}
	if string(data) != want {
		t.Errorf("#%d: read to %c = %s, want %s", rno, delim, string(data), want)
	}
}

func TestUnreadRuneError(t *testing.T) {
	// Check that UnreadRune fails if the operation before it was not a ReadRune.
	var buf [3]byte // all the runes of this test are 3 bytes long
	var segs [1]string
	segs[0] = "日本語日本語日本語"
	sr := stringReader{data: segs[:]}
	r := bufio.NewReader(t.Allocator(), &sr)
	defer r.Free()

	if r.UnreadRune() == nil {
		t.Error("UnreadRune() = nil on a fresh buffer, want an error")
	}
	res := r.ReadRune()
	if res.Err != nil {
		t.Errorf("ReadRune() (1) = %s, want nil", errName(errCode(res.Err)))
	}
	if err := r.UnreadRune(); err != nil {
		t.Errorf("UnreadRune() (1) = %s, want nil", errName(errCode(err)))
	}
	if r.UnreadRune() == nil {
		t.Error("second UnreadRune() (1) = nil, want an error")
	}

	// The error after Read.
	res = r.ReadRune() // reset the state
	if res.Err != nil {
		t.Errorf("ReadRune() (2) = %s, want nil", errName(errCode(res.Err)))
	}
	_, err := r.Read(buf[:])
	if err != nil {
		t.Errorf("Read() (2) = %s, want nil", errName(errCode(err)))
	}
	if r.UnreadRune() == nil {
		t.Error("UnreadRune() after Read = nil, want an error")
	}

	// The error after ReadByte.
	res = r.ReadRune() // reset the state
	if res.Err != nil {
		t.Errorf("ReadRune() (3) = %s, want nil", errName(errCode(res.Err)))
	}
	for range len(buf) {
		_, err = r.ReadByte()
		if err != nil {
			t.Errorf("ReadByte() (3) = %s, want nil", errName(errCode(err)))
		}
	}
	if r.UnreadRune() == nil {
		t.Error("UnreadRune() after ReadByte = nil, want an error")
	}

	// The error after UnreadByte.
	res = r.ReadRune() // reset the state
	if res.Err != nil {
		t.Errorf("ReadRune() (4) = %s, want nil", errName(errCode(res.Err)))
	}
	_, err = r.ReadByte()
	if err != nil {
		t.Errorf("ReadByte() (4) = %s, want nil", errName(errCode(err)))
	}
	err = r.UnreadByte()
	if err != nil {
		t.Errorf("UnreadByte() (4) = %s, want nil", errName(errCode(err)))
	}
	if r.UnreadRune() == nil {
		t.Error("UnreadRune() after UnreadByte = nil, want an error")
	}

	// The error after ReadSlice.
	res = r.ReadRune() // reset the state
	if res.Err != nil {
		t.Errorf("ReadRune() (5) = %s, want nil", errName(errCode(res.Err)))
	}
	_, err = r.ReadSlice(0)
	if errCode(err) != errEOF {
		t.Errorf("ReadSlice() (5) = %s, want EOF", errName(errCode(err)))
	}
	if r.UnreadRune() == nil {
		t.Error("UnreadRune() after ReadSlice = nil, want an error")
	}
}

func TestUnreadRuneAtEOF(t *testing.T) {
	// UnreadRune and ReadRune must fail at EOF.
	sr := strings.NewReader("x")
	r := bufio.NewReader(t.Allocator(), &sr)
	defer r.Free()

	r.ReadRune()
	r.ReadRune()
	r.UnreadRune()
	res := r.ReadRune()
	if errCode(res.Err) != errEOF {
		t.Errorf("ReadRune() at EOF = %s, want EOF", errName(errCode(res.Err)))
	}
}

func TestBufferFull(t *testing.T) {
	const longString = "And now, hello, world! It is the time for all good men to come to the aid of their party"
	sr := strings.NewReader(longString)
	buf := bufio.NewReaderSize(t.Allocator(), &sr, minReadBufferSize)
	defer buf.Free()

	line, err := buf.ReadSlice('!')
	if string(line) != "And now, hello, " || errCode(err) != errBufferFull {
		t.Errorf("first ReadSlice() = %s, %s, want And now, hello, , ErrBufferFull",
			string(line), errName(errCode(err)))
	}
	line, err = buf.ReadSlice('!')
	if string(line) != "world!" || err != nil {
		t.Errorf("second ReadSlice() = %s, %s, want world!, nil",
			string(line), errName(errCode(err)))
	}
}

func TestPeek(t *testing.T) {
	alloc := t.Allocator()
	var p [10]byte
	// The string is minReadBufferSize long.
	sr := strings.NewReader("abcdefghijklmnop")
	buf := bufio.NewReaderSize(alloc, &sr, minReadBufferSize)

	if s, err := buf.Peek(1); string(s) != "a" || err != nil {
		t.Errorf("Peek(1) = %s, %s, want a, nil", string(s), errName(errCode(err)))
	}
	if s, err := buf.Peek(4); string(s) != "abcd" || err != nil {
		t.Errorf("Peek(4) = %s, %s, want abcd, nil", string(s), errName(errCode(err)))
	}
	if _, err := buf.Peek(-1); errCode(err) != errNegativeCount {
		t.Errorf("Peek(-1) = %s, want ErrNegativeCount", errName(errCode(err)))
	}
	if s, err := buf.Peek(32); string(s) != "abcdefghijklmnop" || errCode(err) != errBufferFull {
		t.Errorf("Peek(32) = %s, %s, want abcdefghijklmnop, ErrBufferFull",
			string(s), errName(errCode(err)))
	}
	if _, err := buf.Read(p[0:3]); string(p[0:3]) != "abc" || err != nil {
		t.Errorf("Read(3) = %s, %s, want abc, nil", string(p[0:3]), errName(errCode(err)))
	}
	if s, err := buf.Peek(1); string(s) != "d" || err != nil {
		t.Errorf("Peek(1) = %s, %s, want d, nil", string(s), errName(errCode(err)))
	}
	if s, err := buf.Peek(2); string(s) != "de" || err != nil {
		t.Errorf("Peek(2) = %s, %s, want de, nil", string(s), errName(errCode(err)))
	}
	if _, err := buf.Read(p[0:3]); string(p[0:3]) != "def" || err != nil {
		t.Errorf("Read(3) = %s, %s, want def, nil", string(p[0:3]), errName(errCode(err)))
	}
	if s, err := buf.Peek(4); string(s) != "ghij" || err != nil {
		t.Errorf("Peek(4) = %s, %s, want ghij, nil", string(s), errName(errCode(err)))
	}
	if _, err := buf.Read(p[0:]); string(p[0:]) != "ghijklmnop" || err != nil {
		t.Errorf("Read(10) = %s, %s, want ghijklmnop, nil", string(p[0:]), errName(errCode(err)))
	}
	if s, err := buf.Peek(0); string(s) != "" || err != nil {
		t.Errorf("Peek(0) = %s, %s, want the empty string, nil", string(s), errName(errCode(err)))
	}
	if _, err := buf.Peek(1); errCode(err) != errEOF {
		t.Errorf("Peek(1) at EOF = %s, want EOF", errName(errCode(err)))
	}
	buf.Free()

	// A successful Peek hides the error of the reader.
	dr := dataAndEOFReader{s: "abcd"}
	buf = bufio.NewReaderSize(alloc, &dr, 32)
	defer buf.Free()

	if s, err := buf.Peek(2); string(s) != "ab" || err != nil {
		t.Errorf("Peek(2) on abcd+EOF = %s, %s, want ab, nil", string(s), errName(errCode(err)))
	}
	if s, err := buf.Peek(4); string(s) != "abcd" || err != nil {
		t.Errorf("Peek(4) on abcd+EOF = %s, %s, want abcd, nil", string(s), errName(errCode(err)))
	}
	if n, err := buf.Read(p[0:5]); string(p[0:n]) != "abcd" || err != nil {
		t.Errorf("Read after Peek = %s, %s, want abcd, nil", string(p[0:n]), errName(errCode(err)))
	}
	if n, err := buf.Read(p[0:1]); n != 0 || errCode(err) != errEOF {
		t.Errorf("second Read after Peek = %d, %s, want 0, EOF", n, errName(errCode(err)))
	}
}

func TestPeekThenUnreadRune(t *testing.T) {
	// This sequence used to crash.
	sr := strings.NewReader("x")
	r := bufio.NewReader(t.Allocator(), &sr)
	defer r.Free()

	r.ReadRune()
	r.Peek(1)
	r.UnreadRune()
	r.ReadRune() // used to panic here
}

const testOutput = "0123456789abcdefghijklmnopqrstuvwxy"
const testInput = "012\n345\n678\n9ab\ncde\nfgh\nijk\nlmn\nopq\nrst\nuvw\nxy"
const testInputrn = "012\r\n345\r\n678\r\n9ab\r\ncde\r\nfgh\r\nijk\r\nlmn\r\nopq\r\nrst\r\nuvw\r\nxy\r\n\n\r\n"

// testReadLine reads the input line by line and compares it to testOutput.
func testReadLine(t *testing.T, input string) {
	reader := strideReader{data: []byte(input), stride: 1}
	l := bufio.NewReaderSize(t.Allocator(), &reader, len(input)+1)
	defer l.Free()

	done := 0
	for {
		res := l.ReadLine()
		if len(res.Line) > 0 && res.Err != nil {
			t.Errorf("ReadLine() = %s, %s, want a line or an error",
				string(res.Line), errName(errCode(res.Err)))
		}
		if res.IsPrefix {
			t.Error("ReadLine() = a prefix, want a whole line")
		}
		if res.Err != nil {
			if res.Err != io.EOF {
				t.Errorf("ReadLine() = %s, want EOF", errName(errCode(res.Err)))
				return
			}
			break
		}
		want := testOutput[done : done+len(res.Line)]
		if string(res.Line) != want {
			t.Errorf("ReadLine() = %s, want %s", string(res.Line), want)
		}
		done += len(res.Line)
	}
	if done != len(testOutput) {
		t.Errorf("ReadLine() read %d bytes, want %d", done, len(testOutput))
	}
}

func TestReadLine(t *testing.T) {
	testReadLine(t, testInput)
	testReadLine(t, testInputrn)
}

func TestLineTooLong(t *testing.T) {
	var data [minReadBufferSize * 5 / 2]byte
	for i := range len(data) {
		data[i] = '0' + byte(i%10)
	}
	buf := bytes.NewReader(data[:])
	l := bufio.NewReaderSize(t.Allocator(), &buf, minReadBufferSize)
	defer l.Free()

	rest := data[:]
	res := l.ReadLine()
	if !res.IsPrefix || string(res.Line) != string(rest[:minReadBufferSize]) || res.Err != nil {
		t.Errorf("first ReadLine() = %s, %t, %s, want %s, true, nil",
			string(res.Line), res.IsPrefix, errName(errCode(res.Err)),
			string(rest[:minReadBufferSize]))
	}
	rest = rest[len(res.Line):]
	res = l.ReadLine()
	if !res.IsPrefix || string(res.Line) != string(rest[:minReadBufferSize]) || res.Err != nil {
		t.Errorf("second ReadLine() = %s, %t, %s, want %s, true, nil",
			string(res.Line), res.IsPrefix, errName(errCode(res.Err)),
			string(rest[:minReadBufferSize]))
	}
	rest = rest[len(res.Line):]
	res = l.ReadLine()
	if res.IsPrefix || string(res.Line) != string(rest[:minReadBufferSize/2]) || res.Err != nil {
		t.Errorf("third ReadLine() = %s, %t, %s, want %s, false, nil",
			string(res.Line), res.IsPrefix, errName(errCode(res.Err)),
			string(rest[:minReadBufferSize/2]))
	}
	res = l.ReadLine()
	if res.IsPrefix || res.Err == nil {
		t.Errorf("fourth ReadLine() = %s, %t, %s, want the empty line, false, an error",
			string(res.Line), res.IsPrefix, errName(errCode(res.Err)))
	}
}

func TestReadAfterLines(t *testing.T) {
	const line1 = "this is line1"
	const restData = "this is line2\nthis is line 3\n"
	inbuf := bytes.NewReader([]byte(line1 + "\n" + restData))
	var outArr [64]byte
	outbuf := bytes.NewBuffer(mem.NoAlloc, outArr[:0])
	maxLineLength := len(line1) + len(restData)/2
	l := bufio.NewReaderSize(t.Allocator(), &inbuf, maxLineLength)
	defer l.Free()

	res := l.ReadLine()
	if res.IsPrefix || res.Err != nil || string(res.Line) != line1 {
		t.Errorf("first ReadLine() = %s, %t, %s, want %s, false, nil",
			string(res.Line), res.IsPrefix, errName(errCode(res.Err)), line1)
	}
	n, err := io.Copy(&outbuf, &l)
	if int(n) != len(restData) || err != nil {
		t.Errorf("Copy() = %d, %s, want %d, nil", n, errName(errCode(err)), len(restData))
	}
	if outbuf.String() != restData {
		t.Errorf("Copy() wrote %s, want %s", outbuf.String(), restData)
	}
}

func TestReadEmptyBuffer(t *testing.T) {
	var arr [1]byte
	buf := bytes.NewBuffer(mem.NoAlloc, arr[:0])
	l := bufio.NewReaderSize(t.Allocator(), &buf, minReadBufferSize)
	defer l.Free()

	res := l.ReadLine()
	if errCode(res.Err) != errEOF {
		t.Errorf("ReadLine() = %s, %t, %s, want EOF",
			string(res.Line), res.IsPrefix, errName(errCode(res.Err)))
	}
}

func TestLinesAfterRead(t *testing.T) {
	alloc := t.Allocator()
	br := bytes.NewReader([]byte("foo"))
	l := bufio.NewReaderSize(alloc, &br, minReadBufferSize)
	defer l.Free()

	all, err := io.ReadAll(alloc, &l)
	if err != nil {
		t.Errorf("ReadAll() = %s, want nil", errName(errCode(err)))
		return
	}
	mem.FreeSlice(alloc, all)

	res := l.ReadLine()
	if errCode(res.Err) != errEOF {
		t.Errorf("ReadLine() = %s, %t, %s, want EOF",
			string(res.Line), res.IsPrefix, errName(errCode(res.Err)))
	}
}

func TestReadLineNonNilLineOrError(t *testing.T) {
	sr := strings.NewReader("line 1\n")
	r := bufio.NewReader(t.Allocator(), &sr)
	defer r.Free()

	for i := range 2 {
		res := r.ReadLine()
		if res.Line != nil && res.Err != nil {
			t.Errorf("line %d: ReadLine() = %s, %s, want a line or an error, not both",
				i+1, string(res.Line), errName(errCode(res.Err)))
		}
	}
}

// readLineResult is a wanted result of a ReadLine call.
type readLineResult struct {
	line     string
	isPrefix bool
	err      int
}

// The wanted results of the first input of the ReadLine newlines test.
var newlinesResults1 = []readLineResult{
	{"012345678901234", true, errNone},
	{"", false, errNone},
	{"012345678901234", true, errNone},
	{"", false, errNone},
	{"", false, errEOF},
}

// The wanted results of the second input of the ReadLine newlines test.
var newlinesResults2 = []readLineResult{
	{"0123456789012345", true, errNone},
	{"\r012345678901234", true, errNone},
	{"\r", false, errNone},
	{"", false, errEOF},
}

func TestReadLineNewlines(t *testing.T) {
	testReadLineNewlines(t, "012345678901234\r\n012345678901234\r\n", newlinesResults1)
	testReadLineNewlines(t, "0123456789012345\r012345678901234\r", newlinesResults2)
}

func testReadLineNewlines(t *testing.T, input string, expect []readLineResult) {
	sr := strings.NewReader(input)
	b := bufio.NewReaderSize(t.Allocator(), &sr, minReadBufferSize)
	defer b.Free()

	for i := range expect {
		e := expect[i]
		res := b.ReadLine()
		if string(res.Line) != e.line {
			t.Errorf("%s call %d: line = %s, want %s", input, i, string(res.Line), e.line)
			return
		}
		if res.IsPrefix != e.isPrefix {
			t.Errorf("%s call %d: isPrefix = %t, want %t", input, i, res.IsPrefix, e.isPrefix)
			return
		}
		if errCode(res.Err) != e.err {
			t.Errorf("%s call %d: error = %s, want %s",
				input, i, errName(errCode(res.Err)), errName(e.err))
			return
		}
	}
}

// fillTestInput fills the buffer with a sequence that rarely repeats.
// 101 and 251 are arbitrary prime numbers.
func fillTestInput(input []byte) {
	for i := range input {
		input[i] = byte(i % 251)
		if i%101 == 0 {
			input[i] ^= byte(i / 101)
		}
	}
}

func TestReaderWriteTo(t *testing.T) {
	const size = 8192
	alloc := t.Allocator()
	input := mem.AllocSlice[byte](alloc, size, size)
	fillTestInput(input)

	br := bytes.NewReader(input)
	r := bufio.NewReader(alloc, &br)
	w := bytes.NewBuffer(alloc, nil)

	n, err := r.WriteTo(&w)
	if err != nil || int(n) != len(input) {
		t.Errorf("WriteTo() = %d, %s, want %d, nil", n, errName(errCode(err)), len(input))
		return
	}
	written := w.Bytes()
	if string(written) != string(input) {
		t.Error("WriteTo() wrote the wrong bytes")
	}

	w.Free()
	r.Free()
	mem.FreeSlice(alloc, input)
}

// writerToCase is a case of the WriteTo errors test.
type writerToCase struct {
	rn, wn   int
	rerr     int
	werr     int
	expected int
}

var writerToCases = []writerToCase{
	{1, 0, errNone, errClosedPipe, errClosedPipe},
	{0, 1, errClosedPipe, errNone, errClosedPipe},
	{0, 0, errUnexpectedEOF, errClosedPipe, errUnexpectedEOF},
	{0, 1, errEOF, errNone, errNone},
}

func TestReaderWriteToErrors(t *testing.T) {
	for i := range writerToCases {
		tc := writerToCases[i]
		rw := errorReadWriter{rn: tc.rn, wn: tc.wn, rerr: tc.rerr, werr: tc.werr}
		r := bufio.NewReader(t.Allocator(), &rw)
		_, err := r.WriteTo(&rw)
		if errCode(err) != tc.expected {
			t.Errorf("case %d: WriteTo() = %s, want %s",
				i, errName(errCode(err)), errName(tc.expected))
		}
		r.Free()
	}
}

func TestReaderClearError(t *testing.T) {
	r := errorThenGoodReader{}
	b := bufio.NewReader(t.Allocator(), &r)
	defer b.Free()

	var buf [1]byte
	if _, err := b.Read(nil); err != nil {
		t.Errorf("first empty Read() = %s, want nil", errName(errCode(err)))
	}
	if _, err := b.Read(buf[:]); errCode(err) != errFake {
		t.Errorf("first Read() = %s, want the fake error", errName(errCode(err)))
	}
	if _, err := b.Read(nil); err != nil {
		t.Errorf("second empty Read() = %s, want nil", errName(errCode(err)))
	}
	if _, err := b.Read(buf[:]); err != nil {
		t.Errorf("second Read() = %s, want nil", errName(errCode(err)))
	}
	if r.nread != 2 {
		t.Errorf("reads = %d, want 2", r.nread)
	}
}

// readZeroWant reads from the reader and compares the result to want.
func readZeroWant(t *testing.T, br *bufio.Reader, size int, want string, wantErr int) {
	var p [50]byte
	n, err := br.Read(p[:])
	if errCode(err) != wantErr || n != len(want) || string(p[:n]) != want {
		t.Errorf("bufsize=%d: Read() = %s, %s, want %s, %s",
			size, string(p[:n]), errName(errCode(err)), want, errName(wantErr))
	}
}

func TestReadZero(t *testing.T) {
	// Check that a read of no bytes from the
	// reader gives no bytes and no error.
	sizes := []int{100, 2}
	for i := range sizes {
		size := sizes[i]
		r1 := strings.NewReader("abc")
		r2 := strings.NewReader("def")
		etr := emptyThenNonEmptyReader{r: &r2, n: 1}
		mr := io.NewMultiReader(&r1, &etr)
		br := bufio.NewReaderSize(t.Allocator(), &mr, size)
		readZeroWant(t, &br, size, "abc", errNone)
		readZeroWant(t, &br, size, "", errNone)
		readZeroWant(t, &br, size, "def", errNone)
		readZeroWant(t, &br, size, "", errEOF)
		br.Free()
	}
}

// checkReadAll reads the whole content of the reader and compares it to want.
func checkReadAll(t *testing.T, a mem.Allocator, r *bufio.Reader, want string) {
	all, err := io.ReadAll(a, r)
	if err != nil {
		t.Errorf("ReadAll() = %s, want nil", errName(errCode(err)))
		return
	}
	if string(all) != want {
		t.Errorf("ReadAll() = %s, want %s", string(all), want)
	}
	mem.FreeSlice(a, all)
}

func TestReaderReset(t *testing.T) {
	alloc := t.Allocator()
	sr := strings.NewReader("foo")
	r := bufio.NewReader(alloc, &sr)

	var buf [3]byte
	r.Read(buf[:])
	if string(buf[:]) != "foo" {
		t.Errorf("Read() = %s, want foo", string(buf[:]))
	}

	sr = strings.NewReader("bar bar")
	r.Reset(&sr)
	checkReadAll(t, alloc, &r, "bar bar")
	r.Free()

	// Reset on the zero value of a reader allocates the internal buffer.
	r = bufio.Reader{}
	sr = strings.NewReader("bar bar")
	r.Reset(&sr)
	checkReadAll(t, alloc, &r, "bar bar")

	// A reset of a reader to itself does nothing.
	sr = strings.NewReader("recur")
	r.Reset(&sr)
	r.Reset(&r)
	checkReadAll(t, alloc, &r, "recur")

	r.Free()
}

// The readers of the Discard test.
const (
	discardAlphabet    = iota // gives the lowercase letters
	discardFiveThenErr        // gives five bytes and an error at every read
	discardUnused             // expects no read
)

// discardCase is a case of the Discard test.
type discardCase struct {
	name         string
	reader       int
	bufSize      int // 0 means minReadBufferSize
	peekSize     int
	n            int // the input of Discard
	want         int // the result of Discard
	wantErr      int
	wantBuffered int
}

// An error of the fill does not show up until the reader passes the valid
// bytes. The fill cases give five valid bytes with an error and check that
// Discard hides the error.
var discardCases = []discardCase{
	{"normal case", discardAlphabet, 0, 16, 6, 6, errNone, 10},
	{"discard causing read", discardAlphabet, 0, 0, 6, 6, errNone, 10},
	{"discard all without peek", discardAlphabet, 0, 0, 26, 26, errNone, 0},
	{"discard more than end", discardAlphabet, 0, 0, 27, 26, errEOF, 0},
	{"fill error, discard less", discardFiveThenErr, 0, 0, 4, 4, errNone, 1},
	{"fill error, discard equal", discardFiveThenErr, 0, 0, 5, 5, errNone, 0},
	{"fill error, discard more", discardFiveThenErr, 0, 0, 6, 5, errFive, 0},
	// A discard of no bytes does not read.
	{"discard zero", discardUnused, 0, 0, 0, 0, errNone, 0},
	{"discard negative", discardUnused, 0, 0, -1, 0, errNegativeCount, 0},
}

func TestReaderDiscard(t *testing.T) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	for i := range discardCases {
		tc := discardCases[i]

		var sr strings.Reader
		var five fiveThenErrReader
		var unused unusedReader
		var rd io.Reader
		switch tc.reader {
		case discardAlphabet:
			sr = strings.NewReader(alphabet)
			rd = &sr
		case discardFiveThenErr:
			rd = &five
		default:
			rd = &unused
		}

		br := bufio.NewReaderSize(t.Allocator(), rd, tc.bufSize)
		if tc.peekSize > 0 {
			peekBuf, err := br.Peek(tc.peekSize)
			if err != nil {
				t.Errorf("%s: Peek(%d) = %s, want nil",
					tc.name, tc.peekSize, errName(errCode(err)))
				br.Free()
				continue
			}
			if len(peekBuf) != tc.peekSize {
				t.Errorf("%s: len(Peek(%d)) = %d, want %d",
					tc.name, tc.peekSize, len(peekBuf), tc.peekSize)
				br.Free()
				continue
			}
		}

		discarded, err := br.Discard(tc.n)
		if discarded != tc.want || errCode(err) != tc.wantErr {
			t.Errorf("%s: Discard(%d) = %d, %s, want %d, %s",
				tc.name, tc.n, discarded, errName(errCode(err)),
				tc.want, errName(tc.wantErr))
			br.Free()
			continue
		}
		if bn := br.Buffered(); bn != tc.wantBuffered {
			t.Errorf("%s: Buffered() after Discard = %d, want %d",
				tc.name, bn, tc.wantBuffered)
		}
		if five.small {
			t.Errorf("%s: the fill read is too small", tc.name)
		}
		if unused.used {
			t.Errorf("%s: Discard(%d) read from the reader", tc.name, tc.n)
		}
		br.Free()
	}
}

func TestReaderSize(t *testing.T) {
	alloc := t.Allocator()
	sr := strings.NewReader("")

	r := bufio.NewReader(alloc, &sr)
	if got := r.Size(); got != bufio.DefaultBufSize {
		t.Errorf("NewReader: Size() = %d, want %d", got, bufio.DefaultBufSize)
	}
	r.Free()

	r = bufio.NewReaderSize(alloc, &sr, 1234)
	if got := r.Size(); got != 1234 {
		t.Errorf("NewReaderSize: Size() = %d, want 1234", got)
	}
	r.Free()
}

func TestNewReaderSizeIdempotent(t *testing.T) {
	const bufSize = 1000
	alloc := t.Allocator()
	sr := strings.NewReader("hello world")
	b := bufio.NewReaderSize(alloc, &sr, bufSize)

	// The reader must recognize itself.
	b1 := bufio.NewReaderSize(alloc, &b, bufSize)
	if b1.Size() != b.Size() {
		t.Errorf("NewReaderSize: Size() = %d, want %d", b1.Size(), b.Size())
	}

	// The reader must wrap a buffer that is too small.
	b2 := bufio.NewReaderSize(alloc, &b, 2*bufSize)
	if b2.Size() == b.Size() {
		t.Errorf("NewReaderSize: Size() = %d, want a larger buffer", b2.Size())
	}

	b2.Free()
	b.Free()
}

func TestPartialReadEOF(t *testing.T) {
	var src [10]byte
	eofR := eofReader{buf: src[:]}
	r := bufio.NewReader(t.Allocator(), &eofR)
	defer r.Free()

	// Read 5 of the 10 available bytes.
	var dest [5]byte
	read, err := r.Read(dest[:])
	if err != nil {
		t.Errorf("Read() = %s, want nil", errName(errCode(err)))
		return
	}
	if read != len(dest) {
		t.Errorf("Read() = %d bytes, want %d", read, len(dest))
		return
	}

	// The reader buffered all the content of the source.
	if n := len(eofR.buf); n != 0 {
		t.Errorf("the source has %d bytes left, want 0", n)
	}
	// There are still 5 bytes to read.
	if n := r.Buffered(); n != 5 {
		t.Errorf("Buffered() = %d, want 5", n)
	}

	// The second read is a read of no bytes.
	var empty [0]byte
	read, err = r.Read(empty[:])
	if err != nil {
		t.Errorf("second Read() = %s, want nil", errName(errCode(err)))
		return
	}
	if read != 0 {
		t.Errorf("second Read() = %d bytes, want 0", read)
	}
}
