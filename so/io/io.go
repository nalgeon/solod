// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package io provides basic interfaces to I/O primitives.
// Its primary job is to wrap existing implementations of such primitives,
// such as those in package os, into shared public interfaces that
// abstract the functionality, plus some other related primitives.
//
// Because these interfaces and primitives wrap lower-level operations with
// various implementations, unless otherwise informed clients should not
// assume they are safe for parallel execution.
//
// Based on the [io] package, with the following modifications:
//   - Copy and CopyBuffer don't use WriterTo and ReaderFrom.
//   - CopyN doesn't use ReaderFrom.
//   - Discard doesn't use sync.Pool.
//   - NopCloser doesn't use WriterTo.
//   - RuneReader.ReadRune returns a struct instead of multi-return.
//   - WriteString doesn't use WriterTo.
//
// CAUTION: This package is under development, do not use yet.
//
// [io]: https://github.com/golang/go/blob/go1.26.1/src/io/io.go
package io

import (
	_ "embed"

	"github.com/nalgeon/solod/so/errors"
	"github.com/nalgeon/solod/so/mem"
)

// Seek whence values.
const (
	SeekStart   = 0 // seek relative to the origin of the file
	SeekCurrent = 1 // seek relative to the current offset
	SeekEnd     = 2 // seek relative to the end
)

//so:extern
var maxInt64 = int64(1<<63 - 1)

//so:embed io.h
var io_h string

// ErrShortWrite means that a write accepted fewer bytes than requested
// but failed to return an explicit error.
var ErrShortWrite = errors.New("short write")

// errInvalidWrite means that a write returned an impossible count.
var errInvalidWrite = errors.New("invalid write result")

// ErrShortBuffer means that a read required a longer buffer than was provided.
var ErrShortBuffer = errors.New("short buffer")

// EOF is the error returned by Read when no more input is available.
// (Read must return EOF itself, not an error wrapping EOF,
// because callers will test for EOF using ==.)
// Functions should return EOF only to signal a graceful end of input.
// If the EOF occurs unexpectedly in a structured data stream,
// the appropriate error is either [ErrUnexpectedEOF] or some other error
// giving more detail.
var EOF = errors.New("EOF")

// ErrUnexpectedEOF means that EOF was encountered in the
// middle of reading a fixed-size block or data structure.
var ErrUnexpectedEOF = errors.New("unexpected EOF")

// ErrNoProgress is returned by some clients of a [Reader] when
// many calls to Read have failed to return any data or error,
// usually the sign of a broken [Reader] implementation.
var ErrNoProgress = errors.New("multiple Read calls return no data or error")

// Reader is the interface that wraps the basic Read method.
//
// Read reads up to len(p) bytes into p. It returns the number of bytes
// read (0 <= n <= len(p)) and any error encountered. Even if Read
// returns n < len(p), it may use all of p as scratch space during the call.
// If some data is available but not len(p) bytes, Read conventionally
// returns what is available instead of waiting for more.
//
// When Read encounters an error or end-of-file condition after
// successfully reading n > 0 bytes, it returns the number of
// bytes read. It may return the (non-nil) error from the same call
// or return the error (and n == 0) from a subsequent call.
// An instance of this general case is that a Reader returning
// a non-zero number of bytes at the end of the input stream may
// return either err == EOF or err == nil. The next Read should
// return 0, EOF.
//
// Callers should always process the n > 0 bytes returned before
// considering the error err. Doing so correctly handles I/O errors
// that happen after reading some bytes and also both of the
// allowed EOF behaviors.
//
// If len(p) == 0, Read should always return n == 0. It may return a
// non-nil error if some error condition is known, such as EOF.
//
// Implementations of Read are discouraged from returning a
// zero byte count with a nil error, except when len(p) == 0.
// Callers should treat a return of 0 and nil as indicating that
// nothing happened; in particular it does not indicate EOF.
//
// Implementations must not retain p.
type Reader interface {
	Read(p []byte) (n int, err error)
}

// Writer is the interface that wraps the basic Write method.
//
// Write writes len(p) bytes from p to the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
// Write must not modify the slice data, even temporarily.
//
// Implementations must not retain p.
type Writer interface {
	Write(p []byte) (n int, err error)
}

// Closer is the interface that wraps the basic Close method.
//
// The behavior of Close after the first call is undefined.
// Specific implementations may document their own behavior.
type Closer interface {
	Close() error
}

// Seeker is the interface that wraps the basic Seek method.
//
// Seek sets the offset for the next Read or Write to offset,
// interpreted according to whence:
// [SeekStart] means relative to the start of the file,
// [SeekCurrent] means relative to the current offset, and
// [SeekEnd] means relative to the end
// (for example, offset = -2 specifies the penultimate byte of the file).
// Seek returns the new offset relative to the start of the
// file or an error, if any.
//
// Seeking to an offset before the start of the file is an error.
// Seeking to any positive offset may be allowed, but if the new offset exceeds
// the size of the underlying object the behavior of subsequent I/O operations
// is implementation-dependent.
type Seeker interface {
	Seek(offset int64, whence int) (int64, error)
}

// ReadWriter is the interface that groups the basic Read and Write methods.
type ReadWriter interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
}

// ReadCloser is the interface that groups the basic Read and Close methods.
type ReadCloser interface {
	Read(p []byte) (n int, err error)
	Close() error
}

// WriteCloser is the interface that groups the basic Write and Close methods.
type WriteCloser interface {
	Write(p []byte) (n int, err error)
	Close() error
}

// ReadWriteCloser is the interface that groups the basic Read, Write and Close methods.
type ReadWriteCloser interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
}

// ReadSeeker is the interface that groups the basic Read and Seek methods.
type ReadSeeker interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
}

// ReadSeekCloser is the interface that groups the basic Read, Seek and Close
// methods.
type ReadSeekCloser interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
	Close() error
}

// WriteSeeker is the interface that groups the basic Write and Seek methods.
type WriteSeeker interface {
	Write(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
}

// ReadWriteSeeker is the interface that groups the basic Read, Write and Seek methods.
type ReadWriteSeeker interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
}

// ReaderFrom is the interface that wraps the ReadFrom method.
//
// ReadFrom reads data from r until EOF or error.
// The return value n is the number of bytes read.
// Any error except EOF encountered during the read is also returned.
//
// The [Copy] function uses [ReaderFrom] if available.
type ReaderFrom interface {
	ReadFrom(r Reader) (n int64, err error)
}

// WriterTo is the interface that wraps the WriteTo method.
//
// WriteTo writes data to w until there's no more data to write or
// when an error occurs. The return value n is the number of bytes
// written. Any error encountered during the write is also returned.
//
// The Copy function uses WriterTo if available.
type WriterTo interface {
	WriteTo(w Writer) (n int64, err error)
}

// ReaderAt is the interface that wraps the basic ReadAt method.
//
// ReadAt reads len(p) bytes into p starting at offset off in the
// underlying input source. It returns the number of bytes
// read (0 <= n <= len(p)) and any error encountered.
//
// When ReadAt returns n < len(p), it returns a non-nil error
// explaining why more bytes were not returned. In this respect,
// ReadAt is stricter than Read.
//
// Even if ReadAt returns n < len(p), it may use all of p as scratch
// space during the call. If some data is available but not len(p) bytes,
// ReadAt blocks until either all the data is available or an error occurs.
// In this respect ReadAt is different from Read.
//
// If the n = len(p) bytes returned by ReadAt are at the end of the
// input source, ReadAt may return either err == EOF or err == nil.
//
// If ReadAt is reading from an input source with a seek offset,
// ReadAt should not affect nor be affected by the underlying
// seek offset.
//
// Clients of ReadAt can execute parallel ReadAt calls on the
// same input source.
//
// Implementations must not retain p.
type ReaderAt interface {
	ReadAt(p []byte, off int64) (n int, err error)
}

// WriterAt is the interface that wraps the basic WriteAt method.
//
// WriteAt writes len(p) bytes from p to the underlying data stream
// at offset off. It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// WriteAt must return a non-nil error if it returns n < len(p).
//
// If WriteAt is writing to a destination with a seek offset,
// WriteAt should not affect nor be affected by the underlying
// seek offset.
//
// Clients of WriteAt can execute parallel WriteAt calls on the same
// destination if the ranges do not overlap.
//
// Implementations must not retain p.
type WriterAt interface {
	WriteAt(p []byte, off int64) (n int, err error)
}

// ByteReader is the interface that wraps the ReadByte method.
//
// ReadByte reads and returns the next byte from the input or
// any error encountered. If ReadByte returns an error, no input
// byte was consumed, and the returned byte value is undefined.
//
// ReadByte provides an efficient interface for byte-at-time
// processing. A [Reader] that does not implement  ByteReader
// can be wrapped using bufio.NewReader to add this method.
type ByteReader interface {
	ReadByte() (byte, error)
}

// ByteScanner is the interface that adds the UnreadByte method to the
// basic ReadByte method.
//
// UnreadByte causes the next call to ReadByte to return the last byte read.
// If the last operation was not a successful call to ReadByte, UnreadByte may
// return an error, unread the last byte read (or the byte prior to the
// last-unread byte), or (in implementations that support the [Seeker] interface)
// seek to one byte before the current offset.
type ByteScanner interface {
	ReadByte() (byte, error)
	UnreadByte() error
}

// ByteWriter is the interface that wraps the WriteByte method.
type ByteWriter interface {
	WriteByte(c byte) error
}

type RuneSizeResult struct {
	Rune rune
	Size int
	Err  error
}

// RuneReader is the interface that wraps the ReadRune method.
//
// ReadRune reads a single encoded Unicode character
// and returns the rune and its size in bytes. If no character is
// available, err will be set.
type RuneReader interface {
	ReadRune() RuneSizeResult
}

// RuneScanner is the interface that adds the UnreadRune method to the
// basic ReadRune method.
//
// UnreadRune causes the next call to ReadRune to return the last rune read.
// If the last operation was not a successful call to ReadRune, UnreadRune may
// return an error, unread the last rune read (or the rune prior to the
// last-unread rune), or (in implementations that support the [Seeker] interface)
// seek to the start of the rune before the current offset.
type RuneScanner interface {
	ReadRune() RuneSizeResult
	UnreadRune() error
}

// StringWriter is the interface that wraps the WriteString method.
type StringWriter interface {
	WriteString(s string) (n int, err error)
}

// WriteString writes the contents of the string s to w, which accepts a slice of bytes.
// [Writer.Write] is called exactly once.
func WriteString(w Writer, s string) (int, error) {
	return w.Write([]byte(s))
}

// ReadAtLeast reads from r into buf until it has read at least min bytes.
// It returns the number of bytes copied and an error if fewer bytes were read.
// The error is EOF only if no bytes were read.
// If an EOF happens after reading fewer than min bytes,
// ReadAtLeast returns [ErrUnexpectedEOF].
// If min is greater than the length of buf, ReadAtLeast returns [ErrShortBuffer].
// On return, n >= min if and only if err == nil.
// If r returns an error having read at least min bytes, the error is dropped.
func ReadAtLeast(r Reader, buf []byte, min int) (int, error) {
	if len(buf) < min {
		return 0, ErrShortBuffer
	}
	var n int
	var err error
	for n < min && err == nil {
		var nn int
		nn, err = r.Read(buf[n:])
		n += nn
	}
	if n >= min {
		err = nil
	} else if n > 0 && err == EOF {
		err = ErrUnexpectedEOF
	}
	return n, err
}

// ReadFull reads exactly len(buf) bytes from r into buf.
// It returns the number of bytes copied and an error if fewer bytes were read.
// The error is EOF only if no bytes were read.
// If an EOF happens after reading some but not all the bytes,
// ReadFull returns [ErrUnexpectedEOF].
// On return, n == len(buf) if and only if err == nil.
// If r returns an error having read at least len(buf) bytes, the error is dropped.
func ReadFull(r Reader, buf []byte) (int, error) {
	return ReadAtLeast(r, buf, len(buf))
}

// CopyN copies n bytes (or until an error) from src to dst.
// It returns the number of bytes copied and the earliest
// error encountered while copying.
// On return, written == n if and only if err == nil.
func CopyN(dst Writer, src Reader, n int64) (int64, error) {
	r := &LimitedReader{src, n}
	written, err := Copy(dst, r)
	if written == n {
		return n, nil
	}
	if written < n && err == nil {
		// src stopped early; must have been EOF.
		err = EOF
	}
	return written, err
}

// Copy copies from src to dst until either EOF is reached
// on src or an error occurs. It returns the number of bytes
// copied and the first error encountered while copying, if any.
//
// A successful Copy returns err == nil, not err == EOF.
// Because Copy is defined to read from src until EOF, it does
// not treat an EOF from Read as an error to be reported.
//
// Copy allocates a buffer on the stack to hold data during the copy.
func Copy(dst Writer, src Reader) (int64, error) {
	return copyBuffer(dst, src, []byte{})
}

// CopyBuffer is identical to Copy except that it stages through the
// provided buffer rather than allocating a temporary one.
// If buf has zero length, CopyBuffer panics.
func CopyBuffer(dst Writer, src Reader, buf []byte) (int64, error) {
	if len(buf) == 0 {
		panic("empty buffer in CopyBuffer")
	}
	return copyBuffer(dst, src, buf)
}

// copyBuffer is the actual implementation of Copy and CopyBuffer.
// if buf has zero length, one is allocated on the stack.
func copyBuffer(dst Writer, src Reader, buf []byte) (int64, error) {
	var written int64
	var err error

	// Allocate a buffer if the caller did not provide one.
	if len(buf) == 0 {
		size := 8 * 1024 // 8 KiB
		_, ok := src.(*LimitedReader)
		if ok {
			l := src.(*LimitedReader)
			if int64(size) > l.N {
				if l.N < 1 {
					size = 1
				} else {
					size = int(l.N)
				}
			}
		}
		buf = mem.Alloca(size)
	}

	// Copy data from src to dst until EOF or an error.
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

// LimitReader returns a LimitedReader that reads from r
// but stops with EOF after n bytes.
func LimitReader(r Reader, n int64) LimitedReader { return LimitedReader{r, n} }

// A LimitedReader reads from R but limits the amount of
// data returned to just N bytes. Each call to Read
// updates N to reflect the new amount remaining.
// Read returns EOF when N <= 0 or when the underlying R returns EOF.
type LimitedReader struct {
	R Reader // underlying reader
	N int64  // max bytes remaining
}

func (l *LimitedReader) Read(p []byte) (int, error) {
	if l.N <= 0 {
		return 0, EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err := l.R.Read(p)
	l.N -= int64(n)
	return n, err
}

// NewSectionReader returns a [SectionReader] that reads from r
// starting at offset off and stops with EOF after n bytes.
func NewSectionReader(r ReaderAt, off int64, n int64) SectionReader {
	var remaining int64
	if off <= maxInt64-n {
		remaining = n + off
	} else {
		// Overflow, with no way to return error.
		// Assume we can read up to an offset of 1<<63 - 1.
		remaining = maxInt64
	}
	return SectionReader{r, off, off, remaining, n}
}

// SectionReader implements Read, Seek, and ReadAt on a section
// of an underlying [ReaderAt].
type SectionReader struct {
	r     ReaderAt // constant after creation
	base  int64    // constant after creation
	off   int64
	limit int64 // constant after creation
	n     int64 // constant after creation
}

func (s *SectionReader) Read(p []byte) (int, error) {
	if s.off >= s.limit {
		return 0, EOF
	}
	if max := s.limit - s.off; int64(len(p)) > max {
		p = p[0:max]
	}
	n, err := s.r.ReadAt(p, s.off)
	s.off += int64(n)
	return n, err
}

var errWhence = errors.New("Seek: invalid whence")
var errOffset = errors.New("Seek: invalid offset")

func (s *SectionReader) Seek(offset int64, whence int) (int64, error) {
	if whence == SeekStart {
		offset += s.base
	} else if whence == SeekCurrent {
		offset += s.off
	} else if whence == SeekEnd {
		offset += s.limit
	} else {
		return 0, errWhence
	}

	if offset < s.base {
		return 0, errOffset
	}
	s.off = offset
	return offset - s.base, nil
}

func (s *SectionReader) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 || off >= s.Size() {
		return 0, EOF
	}
	off += s.base
	if max := s.limit - off; int64(len(p)) > max {
		p = p[0:max]
		n, err := s.r.ReadAt(p, off)
		if err == nil {
			err = EOF
		}
		return n, err
	}
	return s.r.ReadAt(p, off)
}

// Size returns the size of the section in bytes.
func (s *SectionReader) Size() int64 { return s.limit - s.base }

// ReaderAtOffset represents the underlying [ReaderAt] and offsets for a section.
type ReaderAtOffset struct {
	R   ReaderAt
	Off int64
	N   int64
}

// Outer returns the underlying [ReaderAt] and offsets for the section.
//
// The returned values are the same that were passed to [NewSectionReader]
// when the [SectionReader] was created.
func (s *SectionReader) Outer() ReaderAtOffset {
	return ReaderAtOffset{s.r, s.base, s.n}
}

// An OffsetWriter maps writes at offset base to offset base+off in the underlying writer.
type OffsetWriter struct {
	w    WriterAt
	base int64 // the original offset
	off  int64 // the current offset
}

// NewOffsetWriter returns an [OffsetWriter] that writes to w
// starting at offset off.
func NewOffsetWriter(w WriterAt, off int64) OffsetWriter {
	return OffsetWriter{w, off, off}
}

func (o *OffsetWriter) Write(p []byte) (int, error) {
	n, err := o.w.WriteAt(p, o.off)
	o.off += int64(n)
	return n, err
}

func (o *OffsetWriter) WriteAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, errOffset
	}

	off += o.base
	return o.w.WriteAt(p, off)
}

func (o *OffsetWriter) Seek(offset int64, whence int) (int64, error) {
	if whence == SeekStart {
		offset += o.base
	} else if whence == SeekCurrent {
		offset += o.off
	} else {
		return 0, errWhence
	}

	if offset < o.base {
		return 0, errOffset
	}
	o.off = offset
	return offset - o.base, nil
}

// A TeeReader writes to w what it reads from r.
// All reads from r performed through it are matched with
// corresponding writes to w. There is no internal buffering -
// the write must complete before the read completes.
// Any error encountered while writing is reported as a read error.
type TeeReader struct {
	r Reader
	w Writer
}

// NewTeeReader returns a [TeeReader] that reads from r and writes to w.
func NewTeeReader(r Reader, w Writer) TeeReader {
	return TeeReader{r, w}
}

func (t *TeeReader) Read(p []byte) (int, error) {
	n, err := t.r.Read(p)
	if n > 0 {
		if n, err := t.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return n, err
}

// A DiscardWriter provides Write methods
// that succeed without doing anything.
type DiscardWriter struct{}

// Discard is a [Writer] on which all Write calls
// succeed without doing anything.
var Discard Writer = DiscardWriter{}

func (DiscardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (DiscardWriter) WriteString(s string) (int, error) {
	return len(s), nil
}

// A NopCloser is a [ReadCloser] with a no-op Close method wrapping
// the provided [Reader] r.
type NopCloser struct {
	r Reader
}

// NewNopCloser returns a [NopCloser] wrapping r.
func NewNopCloser(r Reader) NopCloser {
	return NopCloser{r}
}

func (n NopCloser) Read(p []byte) (int, error) {
	return n.r.Read(p)
}

func (NopCloser) Close() error { return nil }

// ReadAll reads from r until an error or EOF and returns the data it read.
// A successful call returns err == nil, not err == EOF. Because ReadAll is
// defined to read from src until EOF, it does not treat an EOF from Read
// as an error to be reported.
//
// The returned slice is heap-allocated and owned by the caller.
func ReadAll(r Reader) ([]byte, error) {
	// Build slices of exponentially growing size,
	// then copy into a perfectly-sized slice at the end.
	b := mem.NewSlice[byte](0, 512)
	// Starting with next equal to 256 (instead of say 512 or 1024)
	// allows less memory usage for small inputs that finish in the
	// early growth stages, but we grow the read sizes quickly such that
	// it does not materially impact medium or large inputs.
	next := 256
	chunks := make([][]byte, 0, 4)
	// Invariant: finalSize = sum(len(c) for c in chunks)
	var finalSize int
	for {
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == EOF {
				err = nil
			}
			if len(chunks) == 0 {
				return b, err
			}

			// Build our final right-sized slice.
			finalSize += len(b)
			final := mem.NewSlice[byte](0, finalSize)
			for _, chunk := range chunks {
				final = append(final, chunk...)
			}
			final = append(final, b...)

			// Free the intermediate slices.
			for _, chunk := range chunks {
				mem.FreeSlice(chunk)
			}
			mem.FreeSlice(b)

			return final, err
		}

		if cap(b)-len(b) < cap(b)/16 {
			// Move to the next intermediate slice.
			chunks = append(chunks, b)
			finalSize += len(b)
			b = mem.NewSlice[byte](0, next)
			next += next / 2
		}
	}
}
