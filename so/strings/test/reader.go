// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings_test

import (
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
	"solod.dev/so/unicode/utf8"
)

func TestReaderReadAll(t *testing.T) {
	const s = "hello world"
	r := strings.NewReader(s)
	if r.Len() != len(s) {
		t.Errorf("Len() = %d before the read, want %d", r.Len(), len(s))
	}
	if r.Size() != int64(len(s)) {
		t.Errorf("Size() = %d, want %d", r.Size(), len(s))
	}

	alloc := t.Allocator()
	b, err := io.ReadAll(alloc, &r)
	if err != nil {
		t.Fatal("ReadAll failed")
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != s {
		t.Errorf("ReadAll() = %s, want %s", string(b), s)
	}
	if r.Len() != 0 {
		t.Errorf("Len() = %d after the read, want 0", r.Len())
	}
	if r.Size() != int64(len(s)) {
		t.Errorf("Size() = %d after the read, want %d", r.Size(), len(s))
	}
}

// A seekCase is a test case of Seek. The cases run one after the other on one
// reader over the ten digits.
type seekCase struct {
	off     int64
	whence  int
	n       int // the number of bytes to read after the seek
	want    string
	wantPos int64
	readErr int
	seekErr int
}

var seekCases = []seekCase{
	{0, io.SeekStart, 20, "0123456789", 0, errNone, errNone},
	{1, io.SeekStart, 1, "1", 1, errNone, errNone},
	{1, io.SeekCurrent, 2, "34", 3, errNone, errNone},
	{-1, io.SeekStart, 0, "", 0, errNone, errOffset},
	{1 << 33, io.SeekStart, 0, "", 1 << 33, errEOF, errNone},
	{1, io.SeekCurrent, 0, "", 1<<33 + 1, errEOF, errNone},
	{0, io.SeekStart, 5, "01234", 0, errNone, errNone},
	{0, io.SeekCurrent, 5, "56789", 5, errNone, errNone},
	{-1, io.SeekEnd, 1, "9", 9, errNone, errNone},
	{0, 42, 0, "", 0, errNone, errWhence},
}

func TestReaderSeek(t *testing.T) {
	r := strings.NewReader("0123456789")
	var buf [20]byte
	for i, tc := range seekCases {
		pos, err := r.Seek(tc.off, tc.whence)
		if errCode(err) != tc.seekErr {
			t.Errorf("case %d: Seek() error = %s, want %s",
				i, errName(errCode(err)), errName(tc.seekErr))
			continue
		}
		if tc.seekErr != errNone {
			continue
		}
		if pos != tc.wantPos {
			t.Errorf("case %d: Seek() = %d, want %d", i, pos, tc.wantPos)
		}
		if tc.n == 0 && tc.readErr == errNone {
			continue
		}
		n, err := r.Read(buf[:tc.n])
		if errCode(err) != tc.readErr {
			t.Errorf("case %d: Read() error = %s, want %s",
				i, errName(errCode(err)), errName(tc.readErr))
			continue
		}
		if string(buf[:n]) != tc.want {
			t.Errorf("case %d: Read() = %s, want %s", i, string(buf[:n]), tc.want)
		}
	}
}

func TestReaderSeekPastEnd(t *testing.T) {
	// A seek past the end of the string succeeds, and the next read gives EOF.
	r := strings.NewReader("0123456789")
	if _, err := r.Seek(1<<31+5, io.SeekStart); err != nil {
		t.Fatal("Seek failed")
		return
	}
	var buf [10]byte
	n, err := r.Read(buf[:])
	if n != 0 || err != io.EOF {
		t.Errorf("Read() = %d, %s, want 0, EOF", n, errName(errCode(err)))
	}
	if r.Len() != 0 {
		t.Errorf("Len() = %d, want 0", r.Len())
	}
}

// A readAtCase is a test case of ReadAt.
type readAtCase struct {
	off  int64
	n    int
	want string
	err  int
}

var readAtCases = []readAtCase{
	{0, 10, "0123456789", errNone},
	{1, 10, "123456789", errEOF},
	{1, 9, "123456789", errNone},
	{11, 10, "", errEOF},
	{0, 0, "", errNone},
	{-1, 0, "", errOffset},
}

func TestReaderReadAt(t *testing.T) {
	// ReadAt does not change the read position of the reader.
	r := strings.NewReader("0123456789")
	var buf [10]byte
	for i, tc := range readAtCases {
		n, err := r.ReadAt(buf[:tc.n], tc.off)
		if string(buf[:n]) != tc.want {
			t.Errorf("case %d: ReadAt() = %s, want %s", i, string(buf[:n]), tc.want)
		}
		if errCode(err) != tc.err {
			t.Errorf("case %d: ReadAt() error = %s, want %s",
				i, errName(errCode(err)), errName(tc.err))
		}
		if r.Len() != 10 {
			t.Errorf("case %d: ReadAt() moved the position, Len() = %d", i, r.Len())
		}
	}
}

func TestReaderReadAtAfterSeek(t *testing.T) {
	// ReadAt reads from the offset, not from the read position.
	r := strings.NewReader("0123456789")
	if _, err := r.Seek(1<<31+5, io.SeekStart); err != nil {
		t.Fatal("Seek failed")
		return
	}
	var buf [3]byte
	n, err := r.ReadAt(buf[:], 2)
	if n != 3 || err != nil {
		t.Errorf("ReadAt() = %d, %s, want 3, nil", n, errName(errCode(err)))
		return
	}
	if string(buf[:]) != "234" {
		t.Errorf("ReadAt() = %s, want 234", string(buf[:]))
	}
}

// The strings of the byte and code point tests.
var readerStrings = []string{"", abcd, faces, commas}

func TestReaderReadByte(t *testing.T) {
	alloc := t.Allocator()
	for _, s := range readerStrings {
		r := strings.NewReader(s)
		// UnreadByte at the start of the string gives an error.
		if err := r.UnreadByte(); err != io.ErrUnread {
			t.Errorf("UnreadByte() at the start = %s, want ErrUnread",
				errName(errCode(err)))
		}
		b := strings.NewBuilder(alloc)
		for i := 0; ; i++ {
			c, err := r.ReadByte()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Errorf("ReadByte() %d error = %s", i, errName(errCode(err)))
				break
			}
			b.WriteByte(c)
		}
		if b.String() != s {
			t.Errorf("ReadByte() gave %d bytes, want %d", b.Len(), len(s))
		}
		b.Free()
	}
}

func TestReaderUnreadByte(t *testing.T) {
	r := strings.NewReader("abc")
	if _, err := r.ReadByte(); err != nil {
		t.Errorf("ReadByte() error = %s", errName(errCode(err)))
		return
	}
	if err := r.UnreadByte(); err != nil {
		t.Errorf("UnreadByte() error = %s", errName(errCode(err)))
		return
	}
	c, err := r.ReadByte()
	if c != 'a' || err != nil {
		t.Errorf("ReadByte() after UnreadByte = %d, want 97", c)
	}
}

func TestReaderReadRune(t *testing.T) {
	for _, s := range readerStrings {
		r := strings.NewReader(s)
		// UnreadRune at the start of the string gives an error.
		if err := r.UnreadRune(); err != io.ErrUnread {
			t.Errorf("UnreadRune() at the start = %s, want ErrUnread",
				errName(errCode(err)))
		}
		pos := 0
		for pos < len(s) {
			want, wantSize := utf8.DecodeRuneInString(s[pos:])
			res := r.ReadRune()
			if res.Err != nil {
				t.Errorf("ReadRune() at %d error = %s", pos, errName(errCode(res.Err)))
				break
			}
			if res.Rune != want || res.Size != wantSize {
				t.Errorf("ReadRune() at %d = %d, %d, want %d, %d",
					pos, res.Rune, res.Size, want, wantSize)
			}
			// UnreadRune gives the same code point back.
			if err := r.UnreadRune(); err != nil {
				t.Errorf("UnreadRune() at %d error = %s", pos, errName(errCode(err)))
				break
			}
			again := r.ReadRune()
			if again.Rune != want || again.Size != wantSize {
				t.Errorf("ReadRune() after UnreadRune at %d = %d, %d, want %d, %d",
					pos, again.Rune, again.Size, want, wantSize)
			}
			pos += res.Size
		}
		res := r.ReadRune()
		if res.Rune != 0 || res.Size != 0 || res.Err != io.EOF {
			t.Errorf("ReadRune() at the end = %d, %d, %s, want 0, 0, EOF",
				res.Rune, res.Size, errName(errCode(res.Err)))
		}
	}
}

func TestReaderReadRuneInvalid(t *testing.T) {
	// An invalid byte gives RuneError with the size 1.
	r := strings.NewReader("a\xffb")
	r.ReadRune()
	res := r.ReadRune()
	if res.Rune != utf8.RuneError || res.Size != 1 || res.Err != nil {
		t.Errorf("ReadRune() of an invalid byte = %d, %d, want 65533, 1",
			res.Rune, res.Size)
	}
}

// The calls that clear the last read code point, by number.
const (
	afterRead = iota
	afterReadByte
	afterUnreadRune
	afterSeek
	afterWriteTo
)

// afterName returns the name of the call with the number.
func afterName(kind int) string {
	switch kind {
	case afterRead:
		return "Read"
	case afterReadByte:
		return "ReadByte"
	case afterUnreadRune:
		return "UnreadRune"
	case afterSeek:
		return "Seek"
	}
	return "WriteTo"
}

func TestReaderUnreadRuneError(t *testing.T) {
	// A call that is not ReadRune clears the last read code point, so the
	// next UnreadRune gives an error.
	alloc := t.Allocator()
	for kind := afterRead; kind <= afterWriteTo; kind++ {
		r := strings.NewReader("0123456789")
		if res := r.ReadRune(); res.Err != nil {
			t.Errorf("ReadRune() error = %s", errName(errCode(res.Err)))
			return
		}
		var buf [1]byte
		switch kind {
		case afterRead:
			r.Read(buf[:])
		case afterReadByte:
			r.ReadByte()
		case afterUnreadRune:
			r.UnreadRune()
		case afterSeek:
			r.Seek(0, io.SeekCurrent)
		case afterWriteTo:
			w := strings.NewBuilder(alloc)
			r.WriteTo(&w)
			w.Free()
		}
		if err := r.UnreadRune(); err == nil {
			t.Errorf("UnreadRune() after %s = nil, want an error", afterName(kind))
		}
	}
}

func TestReaderWriteTo(t *testing.T) {
	alloc := t.Allocator()
	src := "0123456789abcdefghijklmnopqrstuvwxyz"
	for i := 0; i < 30; i += 3 {
		n := 0
		if i > 0 {
			n = len(src) / i
		}
		r := strings.NewReader(src[:n])
		w := strings.NewBuilder(alloc)
		got, err := r.WriteTo(&w)
		if got != int64(n) {
			t.Errorf("WriteTo() = %d, want %d", got, n)
		}
		if err != nil {
			t.Errorf("WriteTo() error = %s", errName(errCode(err)))
		}
		if w.String() != src[:n] {
			t.Errorf("WriteTo() wrote the wrong bytes for the length %d", n)
		}
		if r.Len() != 0 {
			t.Errorf("Len() = %d after WriteTo, want 0", r.Len())
		}
		w.Free()
	}
}

// A plainReader gives the Read method of a reader and hides every other
// method. io.Copy from a plainReader cannot take the WriteTo path.
type plainReader struct {
	r io.Reader
}

func (p *plainReader) Read(b []byte) (int, error) {
	return p.r.Read(b)
}

func TestReaderCopyNothing(t *testing.T) {
	// A copy from an empty reader gives the same result with and without the
	// WriteTo method.
	var w io.DiscardWriter
	r1 := strings.NewReader("")
	withN, withErr := io.Copy(&w, &r1)
	r2 := strings.NewReader("")
	plain := plainReader{r: &r2}
	withoutN, withoutErr := io.Copy(&w, &plain)
	if withN != withoutN {
		t.Errorf("Copy() = %d with WriteTo, %d without", withN, withoutN)
	}
	if errCode(withErr) != errCode(withoutErr) {
		t.Errorf("Copy() error = %s with WriteTo, %s without",
			errName(errCode(withErr)), errName(errCode(withoutErr)))
	}
	if withN != 0 || withErr != nil {
		t.Errorf("Copy() = %d, %s, want 0, nil", withN, errName(errCode(withErr)))
	}
}

func TestReaderReadAfterBigSeek(t *testing.T) {
	// A seek far past the end of the string leaves nothing to read.
	r := strings.NewReader("0123456789")
	if _, err := r.Seek(1<<31+5, io.SeekStart); err != nil {
		t.Fatal("Seek failed")
		return
	}
	var w io.DiscardWriter
	if n, err := io.CopyN(&w, &r, 1); n != 0 || err != io.EOF {
		t.Errorf("CopyN() = %d, %s, want 0, EOF", n, errName(errCode(err)))
	}
}

func TestReaderLen(t *testing.T) {
	// A read lowers Len, and Size stays the same.
	const s = "hello world"
	r := strings.NewReader(s)
	var buf [10]byte
	if n, err := r.Read(buf[:]); err != nil || n != 10 {
		t.Errorf("Read() = %d, %s, want 10, nil", n, errName(errCode(err)))
		return
	}
	if r.Len() != 1 {
		t.Errorf("Len() = %d, want 1", r.Len())
	}
	if r.Size() != int64(len(s)) {
		t.Errorf("Size() = %d, want %d", r.Size(), len(s))
	}
	if n, err := r.Read(buf[:1]); err != nil || n != 1 {
		t.Errorf("Read() = %d, %s, want 1, nil", n, errName(errCode(err)))
		return
	}
	if r.Len() != 0 {
		t.Errorf("Len() = %d, want 0", r.Len())
	}
}

func TestReaderReset(t *testing.T) {
	r := strings.NewReader("世界")
	if res := r.ReadRune(); res.Err != nil {
		t.Errorf("ReadRune() error = %s", errName(errCode(res.Err)))
		return
	}
	const want = "abcdef"
	r.Reset(want)
	// Reset clears the last read code point.
	if err := r.UnreadRune(); err == nil {
		t.Errorf("UnreadRune() after Reset = nil, want an error")
	}
	if r.Len() != len(want) {
		t.Errorf("Len() = %d after Reset, want %d", r.Len(), len(want))
	}
	alloc := t.Allocator()
	got, err := io.ReadAll(alloc, &r)
	if err != nil {
		t.Errorf("ReadAll() error = %s", errName(errCode(err)))
		return
	}
	defer mem.FreeSlice(alloc, got)
	if string(got) != want {
		t.Errorf("ReadAll() after Reset = %s, want %s", string(got), want)
	}
}

func TestReaderZero(t *testing.T) {
	// The zero reader works like a reader over an empty string.
	var r strings.Reader
	if r.Len() != 0 {
		t.Errorf("Len() = %d, want 0", r.Len())
	}
	if r.Size() != 0 {
		t.Errorf("Size() = %d, want 0", r.Size())
	}
	var buf [1]byte
	if n, err := r.Read(buf[:]); n != 0 || err != io.EOF {
		t.Errorf("Read() = %d, %s, want 0, EOF", n, errName(errCode(err)))
	}
	if n, err := r.ReadAt(buf[:], 11); n != 0 || err != io.EOF {
		t.Errorf("ReadAt() = %d, %s, want 0, EOF", n, errName(errCode(err)))
	}
	if c, err := r.ReadByte(); c != 0 || err != io.EOF {
		t.Errorf("ReadByte() = %d, %s, want 0, EOF", c, errName(errCode(err)))
	}
	res := r.ReadRune()
	if res.Rune != 0 || res.Size != 0 || res.Err != io.EOF {
		t.Errorf("ReadRune() = %d, %d, %s, want 0, 0, EOF",
			res.Rune, res.Size, errName(errCode(res.Err)))
	}
	if off, err := r.Seek(11, io.SeekStart); off != 11 || err != nil {
		t.Errorf("Seek() = %d, %s, want 11, nil", off, errName(errCode(err)))
	}
	r.Reset("")
	if r.UnreadByte() == nil {
		t.Errorf("UnreadByte() = nil, want an error")
	}
	if r.UnreadRune() == nil {
		t.Errorf("UnreadRune() = nil, want an error")
	}
	var w io.DiscardWriter
	if n, err := r.WriteTo(&w); n != 0 || err != nil {
		t.Errorf("WriteTo() = %d, %s, want 0, nil", n, errName(errCode(err)))
	}
}
