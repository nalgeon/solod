// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

import (
	"solod.dev/so/io"
	"solod.dev/so/unicode/utf8"
)

// A Reader implements the [io.Reader], [io.ReaderAt], [io.WriterTo], [io.Seeker],
// [io.ByteScanner], and [io.RuneScanner] interfaces by reading from
// a byte slice.
// Unlike a [Buffer], a Reader is read-only and supports seeking.
// The zero value for Reader operates like a Reader of an empty slice.
type Reader struct {
	s        []byte
	i        int64 // current reading index
	prevRune int   // index of previous rune; or < 0
}

// Len returns the number of bytes of the unread portion of the
// slice.
func (r *Reader) Len() int {
	if r.i >= int64(len(r.s)) {
		return 0
	}
	return int(int64(len(r.s)) - r.i)
}

// Size returns the original length of the underlying byte slice.
// Size is the number of bytes available for reading via [Reader.ReadAt].
// The result is unaffected by any method calls except [Reader.Reset].
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
	var err error
	if off < 0 {
		return 0, io.ErrOffset
	}
	if off >= int64(len(r.s)) {
		return 0, io.EOF
	}
	n := copy(b, r.s[off:])
	if n < len(b) {
		err = io.EOF
	}
	return n, err
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

// UnreadByte complements [Reader.ReadByte] in implementing the [io.ByteScanner] interface.
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
	ch, size := utf8.DecodeRune(r.s[r.i:])
	r.i += int64(size)
	return io.RuneSizeResult{Rune: ch, Size: size, Err: nil}
}

// UnreadRune complements [Reader.ReadRune] in implementing the [io.RuneScanner] interface.
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
	b := r.s[r.i:]
	m, err := w.Write(b)
	if m > len(b) {
		panic("bytes.Reader.WriteTo: invalid Write count")
	}
	r.i += int64(m)
	n := int64(m)
	if m != len(b) && err == nil {
		err = io.ErrShortWrite
	}
	return n, err
}

// Reset resets the [Reader] to be reading from b.
func (r *Reader) Reset(b []byte) { *r = Reader{b, 0, -1} }

// NewReader returns a new [Reader] reading from b.
func NewReader(b []byte) Reader { return Reader{b, 0, -1} }
