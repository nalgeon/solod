// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

type eofReader struct{}

func (*eofReader) Read(b []byte) (int, error) {
	_ = b
	return 0, EOF
}

var eofRdr = &eofReader{}

type MultiReader struct {
	readers []Reader
}

func (mr *MultiReader) Read(p []byte) (int, error) {
	var n int
	var err error
	for len(mr.readers) > 0 {
		// Optimization to flatten nested multiReaders (Issue 13558).
		if len(mr.readers) == 1 {
			r0 := mr.readers[0]
			if _, ok := r0.(*MultiReader); ok {
				mr0 := r0.(*MultiReader)
				mr.readers = mr0.readers
				continue
			}
		}
		n, err = mr.readers[0].Read(p)
		if err == EOF {
			// Use eofReader instead of nil to avoid nil panic
			// after performing flatten (Issue 18232).
			mr.readers[0] = eofRdr
			mr.readers = mr.readers[1:]
		}
		if n > 0 || err != EOF {
			if err == EOF && len(mr.readers) > 0 {
				// Don't return EOF yet. More readers remain.
				err = nil
			}
			return n, err
		}
	}
	return 0, EOF
}

func (mr *MultiReader) WriteTo(w Writer) (int64, error) {
	return mr.writeToWithBuffer(w, make([]byte, defaultBufSize))
}

func (mr *MultiReader) writeToWithBuffer(w Writer, buf []byte) (int64, error) {
	var sum int64
	var err error
	for i, r := range mr.readers {
		var n int64
		if _, ok := r.(*MultiReader); ok { // reuse buffer with nested multiReaders
			subMr := r.(*MultiReader)
			n, err = subMr.writeToWithBuffer(w, buf)
		} else {
			n, err = copyBuffer(w, r, buf)
		}
		sum += n
		if err != nil {
			mr.readers = mr.readers[i:] // permit resume / retry after error
			return sum, err
		}
		mr.readers[i] = nil // permit early GC
	}
	mr.readers = nil
	return sum, nil
}

// NewMultiReader returns a Reader that's the logical concatenation of
// the provided input readers. They're read sequentially. Once all
// inputs have returned EOF, Read will return EOF.  If any of the readers
// return a non-nil, non-EOF error, Read will return that error.
//
// NewMultiReader wraps the provided readers without copying.
func NewMultiReader(readers ...Reader) MultiReader {
	return MultiReader{readers}
}

type MultiWriter struct {
	writers []Writer
}

func (t *MultiWriter) Write(p []byte) (int, error) {
	var n int
	var err error
	for _, w := range t.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
		if n != len(p) {
			err = ErrShortWrite
			return n, err
		}
	}
	return len(p), nil
}

func (t *MultiWriter) WriteString(s string) (int, error) {
	var n int
	var err error
	var p []byte // lazily initialized if/when needed
	for _, w := range t.writers {
		if p == nil {
			p = []byte(s)
		}
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
		if n != len(s) {
			return n, ErrShortWrite
		}
	}
	return len(s), nil
}

// NewMultiWriter creates a writer that duplicates its writes to all the
// provided writers, similar to the Unix tee(1) command.
//
// Each write is written to each listed writer, one at a time.
// If a listed writer returns an error, that overall write operation
// stops and returns the error; it does not continue down the list.
//
// NewMultiWriter wraps the provided writers without copying.
// It also doesn't flatten nested MultiWriters.
func NewMultiWriter(writers ...Writer) MultiWriter {
	return MultiWriter{writers}
}
