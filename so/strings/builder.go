// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"solod.dev/so/mem"
	"solod.dev/so/unicode/utf8"
)

// A Builder is used to efficiently build a string using [Builder.Write] methods.
// It minimizes memory copying. The zero value is ready to use (with default allocator).
// Do not copy a non-zero Builder.
//
// The caller is responsible for freeing the builder's resources
// with [Builder.Free] when done using it.
type Builder struct {
	a   mem.Allocator
	buf []byte
}

// String returns the accumulated string.
func (b *Builder) String() string {
	return string(b.buf)
}

// Len returns the number of accumulated bytes; b.Len() == len(b.String()).
func (b *Builder) Len() int { return len(b.buf) }

// Cap returns the capacity of the builder's underlying byte slice. It is the
// total space allocated for the string being built and includes any bytes
// already written.
func (b *Builder) Cap() int { return cap(b.buf) }

// Reset resets the builder to be empty without freeing the underlying buffer.
func (b *Builder) Reset() {
	b.buf = b.buf[:0]
}

// Free frees the internal buffer and resets the builder.
// After Free, the builder can be reused with new writes.
func (b *Builder) Free() {
	mem.FreeSlice(b.a, b.buf)
	b.buf = nil
}

// grow copies the buffer to a new, larger buffer so that there are at least n
// bytes of capacity beyond len(b.buf).
func (b *Builder) grow(n int) {
	newCap := 2*cap(b.buf) + n
	b.buf = mem.ReallocSlice(b.a, b.buf, len(b.buf), newCap)
}

// Grow grows b's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to b
// without another allocation. If n is negative, Grow panics.
func (b *Builder) Grow(n int) {
	if n < 0 {
		panic("strings: negative grow")
	}
	if cap(b.buf)-len(b.buf) < n {
		b.grow(n)
	}
}

// Write appends the contents of p to b's buffer.
// Write always returns len(p), nil.
func (b *Builder) Write(p []byte) (int, error) {
	b.Grow(len(p))
	l := len(b.buf)
	b.buf = b.buf[:l+len(p)]
	copy(b.buf[l:], p)
	return len(p), nil
}

// WriteByte appends the byte c to b's buffer.
// The returned error is always nil.
func (b *Builder) WriteByte(c byte) error {
	b.Grow(1)
	b.buf = append(b.buf, c)
	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to b's buffer.
// It returns the length of r and a nil error.
func (b *Builder) WriteRune(r rune) (int, error) {
	b.Grow(utf8.UTFMax)
	n := len(b.buf)
	b.buf = utf8.AppendRune(b.buf, r)
	return len(b.buf) - n, nil
}

// WriteString appends the contents of s to b's buffer.
// It returns the length of s and a nil error.
func (b *Builder) WriteString(s string) (int, error) {
	b.Grow(len(s))
	l := len(b.buf)
	b.buf = b.buf[:l+len(s)]
	copy(b.buf[l:], s)
	return len(s), nil
}
