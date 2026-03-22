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
// Based on the [io] package, with fewer features.
//
// [io]: https://github.com/golang/go/blob/go1.26.1/src/io/io.go
package io

import (
	"solod.dev/so/errors"
	"solod.dev/so/mem"
)

// Seek whence values.
const (
	SeekStart   = 0 // seek relative to the origin of the file
	SeekCurrent = 1 // seek relative to the current offset
	SeekEnd     = 2 // seek relative to the end
)

// EOF is the error returned by Read when no more input is available.
// (Read must return EOF itself, not an error wrapping EOF,
// because callers will test for EOF using ==.)
// Functions should return EOF only to signal a graceful end of input.
// If the EOF occurs unexpectedly in a structured data stream,
// the appropriate error is either [ErrUnexpectedEOF] or some other error
// giving more detail.
var EOF = errors.New("EOF")

// ErrInvalidWrite means that a write returned an impossible count.
var ErrInvalidWrite = errors.New("io: invalid write result")

// ErrNegativeRead means that a read returned a negative count.
var ErrNegativeRead = errors.New("io: Read returned negative count")

// ErrNoProgress is returned by some clients of a [Reader] when
// many calls to Read have failed to return any data or error,
// usually the sign of a broken [Reader] implementation.
var ErrNoProgress = errors.New("io: multiple Read calls return no data or error")

// ErrOffset is returned by seek functions when the offset argument is invalid.
var ErrOffset = errors.New("io: invalid offset")

// ErrShortBuffer means that a read required a longer buffer than was provided.
var ErrShortBuffer = errors.New("io: short buffer")

// ErrShortWrite means that a write accepted fewer bytes than requested
// but failed to return an explicit error.
var ErrShortWrite = errors.New("io: short write")

// ErrUnexpectedEOF means that EOF was encountered in the
// middle of reading a fixed-size block or data structure.
var ErrUnexpectedEOF = errors.New("io: unexpected EOF")

// ErrUnread is returned by unread operations when they can't perform for some reason.
var ErrUnread = errors.New("io: cannot unread previous read operation")

// ErrWhence is returned by seek functions when the whence argument is invalid.
var ErrWhence = errors.New("io: invalid whence")

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
	buf := mem.Alloca(size)

	var written int64
	var err error
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = ErrInvalidWrite
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

// CopyN copies n bytes (or until an error) from src to dst.
// It returns the number of bytes copied and the earliest
// error encountered while copying.
// On return, written == n if and only if err == nil.
//
// Allocates a buffer on the stack to hold data during the copy.
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

// ReadAll reads from r until an error or EOF and returns the data it read.
// A successful call returns err == nil, not err == EOF. Because ReadAll is
// defined to read from src until EOF, it does not treat an EOF from Read
// as an error to be reported.
//
// The returned slice is allocated; the caller owns it.
func ReadAll(a mem.Allocator, r Reader) ([]byte, error) {
	// Build slices of exponentially growing size,
	// then copy into a perfectly-sized slice at the end.
	b := mem.AllocSlice[byte](a, 0, 512)
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
			final := mem.AllocSlice[byte](a, 0, finalSize)
			for _, chunk := range chunks {
				final = append(final, chunk...)
			}
			final = append(final, b...)

			// Free the intermediate slices.
			for _, chunk := range chunks {
				mem.FreeSlice(a, chunk)
			}
			mem.FreeSlice(a, b)

			return final, err
		}

		if cap(b)-len(b) < cap(b)/16 {
			// Move to the next intermediate slice.
			chunks = append(chunks, b)
			finalSize += len(b)
			b = mem.AllocSlice[byte](a, 0, next)
			next += next / 2
		}
	}
}

// ReadFull reads exactly len(buf) bytes from r into buf.
// It returns the number of bytes copied and an error if fewer bytes were read.
// The error is EOF only if no bytes were read.
// If an EOF happens after reading some but not all the bytes,
// ReadFull returns [ErrUnexpectedEOF].
// On return, n == len(buf) if and only if err == nil.
// If r returns an error having read at least len(buf) bytes, the error is dropped.
func ReadFull(r Reader, buf []byte) (int, error) {
	var n int
	var err error
	for n < len(buf) && err == nil {
		var nn int
		nn, err = r.Read(buf[n:])
		n += nn
	}
	if n >= len(buf) {
		err = nil
	} else if n > 0 && err == EOF {
		err = ErrUnexpectedEOF
	}
	return n, err
}

// WriteString writes the contents of the string s to w, which accepts a slice of bytes.
// [Writer.Write] is called exactly once.
func WriteString(w Writer, s string) (int, error) {
	return w.Write([]byte(s))
}
