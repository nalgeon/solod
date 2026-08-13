// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fmt

import (
	"solod.dev/so/io"
	"solod.dev/so/unicode/utf8"
)

// bufSize is the size of the scratch space of a buffer.
const bufSize = 256

// buffer is the output of the formatting engine. It collects bytes in a fixed
// scratch space and writes the scratch space to w when the space fills. The
// output has no size limit, and buffer allocates nothing.
//
// A buffer keeps the first write error and drops every byte after it. The
// caller reads total and err after the last flush.
type buffer struct {
	w       io.Writer
	scratch [bufSize]byte
	n       int   // used bytes of scratch
	total   int   // bytes written to w
	err     error // first write error
}

// init prepares the buffer to write to w.
func (b *buffer) init(w io.Writer) {
	b.w = w
	b.n = 0
	b.total = 0
	b.err = nil
}

// flush writes the collected bytes to w.
func (b *buffer) flush() {
	if b.n == 0 || b.err != nil {
		b.n = 0
		return
	}
	n, err := b.w.Write(b.scratch[:b.n])
	b.total += n
	b.n = 0
	if err != nil {
		b.err = err
	}
}

// writeByte collects one byte.
func (b *buffer) writeByte(c byte) {
	if b.n == bufSize {
		b.flush()
	}
	if b.err != nil {
		return
	}
	b.scratch[b.n] = c
	b.n++
}

// write collects the bytes of p.
func (b *buffer) write(p []byte) {
	for len(p) > 0 {
		if b.n == bufSize {
			b.flush()
		}
		if b.err != nil {
			return
		}
		n := copy(b.scratch[b.n:], p)
		b.n += n
		p = p[n:]
	}
}

// writeString collects the bytes of s.
func (b *buffer) writeString(s string) {
	b.write([]byte(s))
}

// writeRune collects the UTF-8 bytes of r.
func (b *buffer) writeRune(r rune) {
	var scratch [utf8.UTFMax]byte
	b.write(utf8.AppendRune(scratch[:0], r))
}

// writeRepeat collects n copies of c.
func (b *buffer) writeRepeat(c byte, n int) {
	for range n {
		b.writeByte(c)
	}
}

// bufWriter writes into a byte slice and drops the bytes that do not fit, so
// Sprintf truncates silently.
type bufWriter struct {
	dst []byte // the destination
	n   int    // written bytes
}

// init prepares the writer to write into buf.
func (w *bufWriter) init(buf []byte) {
	w.dst = buf
	w.n = 0
}

// Write copies the bytes of p that fit and drops the rest. A short write
// reports no error, because Sprintf truncates silently.
func (w *bufWriter) Write(p []byte) (int, error) {
	free := len(w.dst) - w.n
	if free <= 0 {
		return 0, nil
	}
	if len(p) > free {
		p = p[:free]
	}
	n := copy(w.dst[w.n:], p)
	w.n += n
	return n, nil
}

// output returns the written bytes as a string. The string points into the
// destination.
func (w *bufWriter) output() string {
	return string(w.dst[:w.n])
}
