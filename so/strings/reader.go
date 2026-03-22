// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"solod.dev/so/io"
	"solod.dev/so/unicode/utf8"
)

// A Reader implements the [io.Reader], [io.ReaderAt], [io.ByteReader], [io.ByteScanner],
// [io.RuneReader], [io.RuneScanner], [io.Seeker], and [io.WriterTo] interfaces by reading
// from a string.
// The zero value for Reader operates like a Reader of an empty string.
type Reader struct {
	s        string
	i        int64 // current reading index
	prevRune int   // index of previous rune; or < 0
}

// Len returns the number of bytes of the unread portion of the
// string.
func (r *Reader) Len() int {
	if r.i >= int64(len(r.s)) {
		return 0
	}
	return int(int64(len(r.s)) - r.i)
}

// Size returns the original length of the underlying string.
// Size is the number of bytes available for reading via [Reader.ReadAt].
// The returned value is always the same and is not affected by calls
// to any other method.
func (r *Reader) Size() int64 { return int64(len(r.s)) }

// Read implements the [io.Reader] interface.
func (r *Reader) Read(b []byte) (int, error) {
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	r.prevRune = -1
	n := copy(b, r.s[r.i:])
	r.i += int64(n)
	return n, nil
}

// ReadAt implements the [io.ReaderAt] interface.
func (r *Reader) ReadAt(b []byte, off int64) (int, error) {
	// cannot modify state - see io.ReaderAt
	if off < 0 {
		return 0, io.ErrOffset
	}
	if off >= int64(len(r.s)) {
		return 0, io.EOF
	}
	n := copy(b, r.s[off:])
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}

// ReadByte implements the [io.ByteReader] interface.
func (r *Reader) ReadByte() (byte, error) {
	r.prevRune = -1
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	b := r.s[r.i]
	r.i++
	return b, nil
}

// UnreadByte implements the [io.ByteScanner] interface.
func (r *Reader) UnreadByte() error {
	if r.i <= 0 {
		return io.ErrUnread
	}
	r.prevRune = -1
	r.i--
	return nil
}

// ReadRune implements the [io.RuneReader] interface.
func (r *Reader) ReadRune() io.RuneSizeResult {
	if r.i >= int64(len(r.s)) {
		r.prevRune = -1
		return io.RuneSizeResult{Rune: 0, Size: 0, Err: io.EOF}
	}
	r.prevRune = int(r.i)
	if c := r.s[r.i]; c < utf8.RuneSelf {
		r.i++
		return io.RuneSizeResult{Rune: rune(c), Size: 1, Err: nil}
	}
	ch, size := utf8.DecodeRuneInString(r.s[r.i:])
	r.i += int64(size)
	return io.RuneSizeResult{Rune: ch, Size: size, Err: nil}
}

// UnreadRune implements the [io.RuneScanner] interface.
func (r *Reader) UnreadRune() error {
	if r.i <= 0 {
		return io.ErrUnread
	}
	if r.prevRune < 0 {
		return io.ErrUnread
	}
	r.i = int64(r.prevRune)
	r.prevRune = -1
	return nil
}

// Seek implements the [io.Seeker] interface.
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	r.prevRune = -1
	var abs int64
	if whence == io.SeekStart {
		abs = offset
	} else if whence == io.SeekCurrent {
		abs = r.i + offset
	} else if whence == io.SeekEnd {
		abs = int64(len(r.s)) + offset
	} else {
		return 0, io.ErrWhence
	}
	if abs < 0 {
		return 0, io.ErrOffset
	}
	r.i = abs
	return abs, nil
}

// WriteTo implements the [io.WriterTo] interface.
func (r *Reader) WriteTo(w io.Writer) (int64, error) {
	var err error
	r.prevRune = -1
	if r.i >= int64(len(r.s)) {
		return 0, nil
	}
	s := r.s[r.i:]
	m, err := io.WriteString(w, s)
	if m > len(s) {
		panic("strings.Reader.WriteTo: invalid WriteString count")
	}
	r.i += int64(m)
	n := int64(m)
	if m != len(s) && err == nil {
		err = io.ErrShortWrite
	}
	return n, err
}

// Reset resets the [Reader] to be reading from s.
func (r *Reader) Reset(s string) { *r = Reader{s: s, prevRune: -1} }

// NewReader returns a new [Reader] reading from s.
// It is similar to [bytes.NewBufferString] but more efficient and non-writable.
func NewReader(s string) Reader { return Reader{s: s, prevRune: -1} }
