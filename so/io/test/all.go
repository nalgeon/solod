// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io_test

import (
	"solod.dev/so/errors"
	"solod.dev/so/io"
)

// The errors of the io tests, by number.
const (
	errNone = iota
	errEOF
	errOffset
	errWhence
	errInvalidWrite
	errShortWrite
	errUnexpectedEOF
	errRead
	errWrite
	errOther
)

// errReadFailed is the error of the readers that fail.
var errReadFailed = errors.New("io_test: read failed")

// errWriteFailed is the error of the writers that fail.
var errWriteFailed = errors.New("io_test: write failed")

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
	if err == io.ErrInvalidWrite {
		return errInvalidWrite
	}
	if err == io.ErrShortWrite {
		return errShortWrite
	}
	if err == io.ErrUnexpectedEOF {
		return errUnexpectedEOF
	}
	if err == errReadFailed {
		return errRead
	}
	if err == errWriteFailed {
		return errWrite
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
	case errInvalidWrite:
		return "ErrInvalidWrite"
	case errShortWrite:
		return "ErrShortWrite"
	case errUnexpectedEOF:
		return "ErrUnexpectedEOF"
	case errRead:
		return "read failed"
	case errWrite:
		return "write failed"
	}
	return "other"
}

// bufWriter collects the written bytes in a buffer of a fixed capacity.
type bufWriter struct {
	buf []byte
}

func (w *bufWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	return len(p), nil
}

// String returns the collected bytes.
func (w *bufWriter) String() string {
	return string(w.buf)
}

// chunkReader gives at most n bytes at one read.
type chunkReader struct {
	s string
	n int
}

func (r *chunkReader) Read(p []byte) (int, error) {
	if len(r.s) == 0 {
		return 0, io.EOF
	}
	if len(p) > r.n {
		p = p[:r.n]
	}
	n := copy(p, []byte(r.s))
	r.s = r.s[n:]
	return n, nil
}

// dataErrReader gives its bytes, then fails.
type dataErrReader struct {
	s string
}

func (r *dataErrReader) Read(p []byte) (int, error) {
	if len(r.s) == 0 {
		return 0, errReadFailed
	}
	n := copy(p, []byte(r.s))
	r.s = r.s[n:]
	return n, nil
}

// errReader fails every read.
type errReader struct{}

func (*errReader) Read(p []byte) (int, error) {
	_ = p
	return 0, errReadFailed
}

// zeroErrReader gives one zero byte and an error at every read.
type zeroErrReader struct{}

func (*zeroErrReader) Read(p []byte) (int, error) {
	var zero [1]byte
	return copy(p, zero[:]), errReadFailed
}

// fullErrReader fills the buffer and reports an error at every read.
type fullErrReader struct{}

func (*fullErrReader) Read(p []byte) (int, error) {
	return len(p), errReadFailed
}

// byteAndEOFReader gives one byte and EOF at once.
type byteAndEOFReader struct {
	b byte
}

func (r *byteAndEOFReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		// A read of no bytes is useless. The tests make no such call.
		panic("io_test: read with an empty buffer")
	}
	p[0] = r.b
	return 1, io.EOF
}

// flakyReader fails at the first read and gives its bytes at the next reads.
type flakyReader struct {
	s      string
	failed bool
}

func (r *flakyReader) Read(p []byte) (int, error) {
	if !r.failed {
		r.failed = true
		return 0, errReadFailed
	}
	if len(r.s) == 0 {
		return 0, io.EOF
	}
	n := copy(p, []byte(r.s))
	r.s = r.s[n:]
	return n, nil
}

// errWriter fails every write.
type errWriter struct{}

func (*errWriter) Write(p []byte) (int, error) {
	_ = p
	return 0, errWriteFailed
}

// largeWriter reports a count larger than the number of the given bytes.
type largeWriter struct {
	fail bool
}

func (w *largeWriter) Write(p []byte) (int, error) {
	if w.fail {
		return len(p) + 1, errWriteFailed
	}
	return len(p) + 1, nil
}

// shortWriter accepts all but the last byte of a write and reports no error.
type shortWriter struct{}

func (*shortWriter) Write(p []byte) (int, error) {
	return len(p) - 1, nil
}

// halfWriter accepts half of a write and reports ErrShortWrite.
type halfWriter struct{}

func (*halfWriter) Write(p []byte) (int, error) {
	return len(p) / 2, io.ErrShortWrite
}

// calledWriter accepts every write and records the call.
type calledWriter struct {
	called bool
}

func (w *calledWriter) Write(p []byte) (int, error) {
	w.called = true
	return len(p), nil
}
