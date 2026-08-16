// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bufio_test

import (
	"solod.dev/so/bufio"
	"solod.dev/so/errors"
	"solod.dev/so/io"
)

// minReadBufferSize is the smallest buffer of a Reader.
const minReadBufferSize = 16

// The errors of the bufio tests, by number.
const (
	errNone = iota
	errEOF
	errUnexpectedEOF
	errShortWrite
	errNoProgress
	errClosedPipe
	errBufferFull
	errInvalidUnreadByte
	errInvalidUnreadRune
	errNegativeCount
	errTooLong
	errNegativeAdvance
	errAdvanceTooFar
	errBadReadCount
	errFinalToken
	errTimeout
	errFake
	errFive
	errWrite
	errSplit
	errOther
)

// errTimeoutFake is a fake timeout error.
var errTimeoutFake = errors.New("bufio_test: fake timeout")

// errFakeRead is the error of the readers that fail once.
var errFakeRead = errors.New("bufio_test: fake error")

// errFiveThenError is the error after five bytes.
var errFiveThenError = errors.New("bufio_test: 5-then-error")

// errWriteFailed is the error of the writers that fail.
var errWriteFailed = errors.New("bufio_test: write error")

// errSplitFailed is the error of the split functions that fail.
var errSplitFailed = errors.New("bufio_test: split error")

// errCode returns the number of the error.
func errCode(err error) int {
	if err == nil {
		return errNone
	}
	switch err {
	case io.EOF:
		return errEOF
	case io.ErrUnexpectedEOF:
		return errUnexpectedEOF
	case io.ErrShortWrite:
		return errShortWrite
	case io.ErrNoProgress:
		return errNoProgress
	case io.ErrClosedPipe:
		return errClosedPipe
	case bufio.ErrBufferFull:
		return errBufferFull
	case bufio.ErrInvalidUnreadByte:
		return errInvalidUnreadByte
	case bufio.ErrInvalidUnreadRune:
		return errInvalidUnreadRune
	case bufio.ErrNegativeCount:
		return errNegativeCount
	case bufio.ErrTooLong:
		return errTooLong
	case bufio.ErrNegativeAdvance:
		return errNegativeAdvance
	case bufio.ErrAdvanceTooFar:
		return errAdvanceTooFar
	case bufio.ErrBadReadCount:
		return errBadReadCount
	case bufio.ErrFinalToken:
		return errFinalToken
	case errTimeoutFake:
		return errTimeout
	case errFakeRead:
		return errFake
	case errFiveThenError:
		return errFive
	case errWriteFailed:
		return errWrite
	case errSplitFailed:
		return errSplit
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
	case errUnexpectedEOF:
		return "ErrUnexpectedEOF"
	case errShortWrite:
		return "ErrShortWrite"
	case errNoProgress:
		return "ErrNoProgress"
	case errClosedPipe:
		return "ErrClosedPipe"
	case errBufferFull:
		return "ErrBufferFull"
	case errInvalidUnreadByte:
		return "ErrInvalidUnreadByte"
	case errInvalidUnreadRune:
		return "ErrInvalidUnreadRune"
	case errNegativeCount:
		return "ErrNegativeCount"
	case errTooLong:
		return "ErrTooLong"
	case errNegativeAdvance:
		return "ErrNegativeAdvance"
	case errAdvanceTooFar:
		return "ErrAdvanceTooFar"
	case errBadReadCount:
		return "ErrBadReadCount"
	case errFinalToken:
		return "ErrFinalToken"
	case errTimeout:
		return "timeout"
	case errFake:
		return "fake error"
	case errFive:
		return "5-then-error"
	case errWrite:
		return "write error"
	case errSplit:
		return "split error"
	}
	return "other"
}

// codeErr returns the error with the number.
func codeErr(code int) error {
	switch code {
	case errEOF:
		return io.EOF
	case errUnexpectedEOF:
		return io.ErrUnexpectedEOF
	case errShortWrite:
		return io.ErrShortWrite
	case errClosedPipe:
		return io.ErrClosedPipe
	case errWrite:
		return errWriteFailed
	}
	return nil
}

// The wrapping modes of a wrapReader.
const (
	wrapFull    = iota // gives all the read bytes
	wrapByte           // gives one byte at each read
	wrapHalf           // gives half of the requested bytes at each read
	wrapDataErr        // gives the last bytes and the final error at one read
	wrapTimeout        // fails the second read, then gives the bytes
)

// wrapReader reads from another reader in one of the wrapping modes.
// It replaces the so/testing/iotest readers, which need closures.
type wrapReader struct {
	r      io.Reader
	mode   int
	count  int
	unread []byte
	data   [512]byte
}

func (r *wrapReader) Read(p []byte) (int, error) {
	switch r.mode {
	case wrapByte:
		if len(p) == 0 {
			return 0, nil
		}
		return r.r.Read(p[0:1])
	case wrapHalf:
		return r.r.Read(p[0 : (len(p)+1)/2])
	case wrapDataErr:
		return r.readDataErr(p)
	case wrapTimeout:
		r.count++
		if r.count == 2 {
			return 0, errTimeoutFake
		}
	}
	return r.r.Read(p)
}

// readDataErr gives the last bytes and the final error at one read.
// The first call needs two reads: one for the bytes and one for the error.
func (r *wrapReader) readDataErr(p []byte) (int, error) {
	n := 0
	var err error
	for {
		if len(r.unread) == 0 {
			n1, err1 := r.r.Read(r.data[:])
			r.unread = r.data[0:n1]
			err = err1
		}
		if n > 0 || err != nil {
			break
		}
		n = copy(p, r.unread)
		r.unread = r.unread[n:]
	}
	return n, err
}

// rot13Reader reads from another reader and rot13s the result.
type rot13Reader struct {
	r io.Reader
}

func (r13 *rot13Reader) Read(p []byte) (int, error) {
	n, err := r13.r.Read(p)
	for i := range n {
		c := p[i] | 0x20 // lowercase byte
		if 'a' <= c && c <= 'm' {
			p[i] += 13
		} else if 'n' <= c && c <= 'z' {
			p[i] -= 13
		}
	}
	return n, err
}

// stringReader gives its data one string segment at a time.
type stringReader struct {
	data []string
	step int
}

func (r *stringReader) Read(p []byte) (int, error) {
	if r.step >= len(r.data) {
		return 0, io.EOF
	}
	s := r.data[r.step]
	n := copy(p, []byte(s))
	r.step++
	return n, nil
}

// zeroReader gives no bytes and no error at every read.
type zeroReader struct{}

func (*zeroReader) Read(p []byte) (int, error) {
	_ = p
	return 0, nil
}

// dataAndEOFReader gives its bytes and io.EOF at one read.
type dataAndEOFReader struct {
	s string
}

func (r *dataAndEOFReader) Read(p []byte) (int, error) {
	return copy(p, []byte(r.s)), io.EOF
}

// strideReader gives reads of a fixed length.
type strideReader struct {
	data   []byte
	stride int
}

func (r *strideReader) Read(buf []byte) (int, error) {
	n := min(r.stride, len(r.data))
	if n > len(buf) {
		n = len(buf)
	}
	copy(buf, r.data)
	r.data = r.data[n:]
	var err error
	if len(r.data) == 0 {
		err = io.EOF
	}
	return n, err
}

// errorThenGoodReader fails the first read and fills the buffer at the next reads.
type errorThenGoodReader struct {
	didErr bool
	nread  int
}

func (r *errorThenGoodReader) Read(p []byte) (int, error) {
	r.nread++
	if !r.didErr {
		r.didErr = true
		return 0, errFakeRead
	}
	return len(p), nil
}

// emptyThenNonEmptyReader gives no bytes at the first n reads,
// then reads from another reader.
type emptyThenNonEmptyReader struct {
	r io.Reader
	n int
}

func (r *emptyThenNonEmptyReader) Read(p []byte) (int, error) {
	if r.n <= 0 {
		return r.r.Read(p)
	}
	r.n--
	return 0, nil
}

// fiveThenErrReader gives five bytes and an error at every read.
// It fails the test if the buffer is too small.
type fiveThenErrReader struct {
	small bool
}

func (r *fiveThenErrReader) Read(p []byte) (int, error) {
	if len(p) < 5 {
		r.small = true
		return 0, errFiveThenError
	}
	copy(p, []byte("12345"))
	return 5, errFiveThenError
}

// unusedReader records the reads. The tests that use it expect no read.
type unusedReader struct {
	used bool
}

func (r *unusedReader) Read(p []byte) (int, error) {
	_ = p
	r.used = true
	return 0, io.EOF
}

// eofReader gives io.EOF with the last bytes of its buffer.
type eofReader struct {
	buf []byte
}

func (r *eofReader) Read(p []byte) (int, error) {
	read := copy(p, r.buf)
	r.buf = r.buf[read:]
	if read == 0 || read == len(r.buf) {
		// As the io.Reader documentation allows, this returns io.EOF
		// with the last bytes of the buffer.
		return read, io.EOF
	}
	return read, nil
}

// errorWriter accepts len(p)*n/m bytes of a write and reports an error.
type errorWriter struct {
	n, m int
	err  int
}

func (w *errorWriter) Write(p []byte) (int, error) {
	return len(p) * w.n / w.m, codeErr(w.err)
}

// countingWriter drops the written bytes and counts the writes.
type countingWriter struct {
	count int
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.count++
	return len(p), nil
}

// errorReadWriter gives len(p)*rn bytes with the rerr error at a read.
// It accepts len(p)*wn bytes with the werr error at a write.
type errorReadWriter struct {
	rn, wn     int
	rerr, werr int
}

func (rw *errorReadWriter) Read(p []byte) (int, error) {
	return len(p) * rw.rn, codeErr(rw.rerr)
}

func (rw *errorReadWriter) Write(p []byte) (int, error) {
	return len(p) * rw.wn, codeErr(rw.werr)
}

// errorOnlyWriter fails every write.
type errorOnlyWriter struct{}

func (*errorOnlyWriter) Write(p []byte) (int, error) {
	_ = p
	return 0, errWriteFailed
}
