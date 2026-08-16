// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io_test

import (
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

// The text the copy tests read.
const text = "hello, world."

// fill writes the letters of the alphabet into buf, repeated, and returns the
// result as a string.
func fill(buf []byte) string {
	for i := range buf {
		buf[i] = byte('a' + i%26)
	}
	return string(buf)
}

func TestCopy(t *testing.T) {
	r := strings.NewReader(text)
	w := bufWriter{buf: make([]byte, 0, 32)}
	n, err := io.Copy(&w, &r)
	if err != nil {
		t.Errorf("Copy() error = %s", errName(errCode(err)))
		return
	}
	if n != int64(len(text)) {
		t.Errorf("Copy() = %d, want %d", n, len(text))
	}
	if w.String() != text {
		t.Errorf("Copy() wrote %s, want %s", w.String(), text)
	}
}

func TestCopyEmpty(t *testing.T) {
	r := strings.NewReader("")
	w := bufWriter{buf: make([]byte, 0, 8)}
	n, err := io.Copy(&w, &r)
	if n != 0 || err != nil {
		t.Errorf("Copy() = %d, %s, want 0, nil", n, errName(errCode(err)))
	}
}

func TestCopyChunks(t *testing.T) {
	// The reader gives two bytes at a time, so Copy makes several writes.
	r := chunkReader{s: text, n: 2}
	w := bufWriter{buf: make([]byte, 0, 32)}
	n, err := io.Copy(&w, &r)
	if n != int64(len(text)) || err != nil {
		t.Errorf("Copy() = %d, %s, want %d, nil", n, errName(errCode(err)), len(text))
	}
	if w.String() != text {
		t.Errorf("Copy() wrote %s, want %s", w.String(), text)
	}
}

func TestCopyNegativeLimit(t *testing.T) {
	// A LimitedReader with a negative limit gives no bytes.
	r := strings.NewReader(text)
	lr := io.LimitedReader{R: &r, N: -1}
	w := bufWriter{buf: make([]byte, 0, 32)}
	n, err := io.Copy(&w, &lr)
	if n != 0 || err != nil {
		t.Errorf("Copy() = %d, %s, want 0, nil", n, errName(errCode(err)))
	}
	if w.String() != "" {
		t.Errorf("Copy() wrote %s, want nothing", w.String())
	}
}

func TestCopyBuffer(t *testing.T) {
	// A buffer of one byte keeps the copy honest.
	r := strings.NewReader(text)
	w := bufWriter{buf: make([]byte, 0, 32)}
	buf := make([]byte, 1)
	n, err := io.CopyBuffer(&w, &r, buf)
	if n != int64(len(text)) || err != nil {
		t.Errorf("CopyBuffer() = %d, %s, want %d, nil", n, errName(errCode(err)), len(text))
	}
	if w.String() != text {
		t.Errorf("CopyBuffer() wrote %s, want %s", w.String(), text)
	}
}

func TestCopyReadErr(t *testing.T) {
	// Copy reports the error of the reader.
	var r errReader
	w := bufWriter{buf: make([]byte, 0, 8)}
	n, err := io.Copy(&w, &r)
	if n != 0 || errCode(err) != errRead {
		t.Errorf("Copy() = %d, %s, want 0, read failed", n, errName(errCode(err)))
	}
}

func TestCopyWriteErr(t *testing.T) {
	// Copy reports the error of the writer.
	r := strings.NewReader(text)
	var w errWriter
	n, err := io.Copy(&w, &r)
	if n != 0 || errCode(err) != errWrite {
		t.Errorf("Copy() = %d, %s, want 0, write failed", n, errName(errCode(err)))
	}
}

func TestCopyReadErrWriteErr(t *testing.T) {
	// A read gives bytes and an error, and the write of those bytes also
	// fails. Copy reports the error of the write, because the write is the
	// operation that stopped the copy.
	var r zeroErrReader
	var w errWriter
	n, err := io.Copy(&w, &r)
	if n != 0 || errCode(err) != errWrite {
		t.Errorf("Copy() = %d, %s, want 0, write failed", n, errName(errCode(err)))
	}
}

func TestCopyShortWrite(t *testing.T) {
	// The writer accepts fewer bytes than requested and reports no error.
	r := strings.NewReader(text)
	var w shortWriter
	n, err := io.Copy(&w, &r)
	if n != int64(len(text)-1) || errCode(err) != errShortWrite {
		t.Errorf("Copy() = %d, %s, want %d, ErrShortWrite",
			n, errName(errCode(err)), len(text)-1)
	}
}

func TestCopyLargeWriter(t *testing.T) {
	// The writer reports an impossible count and no error.
	r := strings.NewReader(text)
	w := largeWriter{}
	n, err := io.Copy(&w, &r)
	if n != 0 || errCode(err) != errInvalidWrite {
		t.Errorf("Copy() = %d, %s, want 0, ErrInvalidWrite", n, errName(errCode(err)))
	}

	// The writer reports an impossible count and an error. The error of the
	// writer wins over ErrInvalidWrite.
	r = strings.NewReader(text)
	w = largeWriter{fail: true}
	n, err = io.Copy(&w, &r)
	if n != 0 || errCode(err) != errWrite {
		t.Errorf("Copy() = %d, %s, want 0, write failed", n, errName(errCode(err)))
	}
}

func TestCopyN(t *testing.T) {
	r := strings.NewReader(text)
	w := bufWriter{buf: make([]byte, 0, 32)}
	n, err := io.CopyN(&w, &r, 5)
	if n != 5 || err != nil {
		t.Errorf("CopyN() = %d, %s, want 5, nil", n, errName(errCode(err)))
	}
	if w.String() != "hello" {
		t.Errorf("CopyN() wrote %s, want hello", w.String())
	}
}

func TestCopyNZero(t *testing.T) {
	r := strings.NewReader(text)
	w := bufWriter{buf: make([]byte, 0, 8)}
	n, err := io.CopyN(&w, &r, 0)
	if n != 0 || err != nil {
		t.Errorf("CopyN() = %d, %s, want 0, nil", n, errName(errCode(err)))
	}
	if w.String() != "" {
		t.Errorf("CopyN() wrote %s, want nothing", w.String())
	}
}

func TestCopyNNegative(t *testing.T) {
	r := strings.NewReader(text)
	w := bufWriter{buf: make([]byte, 0, 8)}
	n, err := io.CopyN(&w, &r, -1)
	if n != 0 || err != nil {
		t.Errorf("CopyN() = %d, %s, want 0, nil", n, errName(errCode(err)))
	}
	if w.String() != "" {
		t.Errorf("CopyN() wrote %s, want nothing", w.String())
	}
}

func TestCopyNAll(t *testing.T) {
	// CopyN of exactly the number of the available bytes reports no error.
	r := strings.NewReader("foo")
	w := bufWriter{buf: make([]byte, 0, 8)}
	n, err := io.CopyN(&w, &r, 3)
	if n != 3 || err != nil {
		t.Errorf("CopyN() = %d, %s, want 3, nil", n, errName(errCode(err)))
	}
	if w.String() != "foo" {
		t.Errorf("CopyN() wrote %s, want foo", w.String())
	}
}

func TestCopyNEOF(t *testing.T) {
	// CopyN of more than the number of the available bytes reports EOF.
	r := strings.NewReader("foo")
	w := bufWriter{buf: make([]byte, 0, 8)}
	n, err := io.CopyN(&w, &r, 4)
	if n != 3 || errCode(err) != errEOF {
		t.Errorf("CopyN() = %d, %s, want 3, EOF", n, errName(errCode(err)))
	}
	if w.String() != "foo" {
		t.Errorf("CopyN() wrote %s, want foo", w.String())
	}
}

func TestCopyNFullErrReader(t *testing.T) {
	// The reader gives all the requested bytes and an error. CopyN copied the
	// requested number of the bytes, so it drops the error.
	var r fullErrReader
	w := bufWriter{buf: make([]byte, 0, 8)}
	n, err := io.CopyN(&w, &r, 5)
	if n != 5 || err != nil {
		t.Errorf("CopyN() = %d, %s, want 5, nil", n, errName(errCode(err)))
	}
	if len(w.String()) != 5 {
		t.Errorf("CopyN() wrote %d bytes, want 5", len(w.String()))
	}
}

func TestReadAll(t *testing.T) {
	alloc := t.Allocator()
	r := strings.NewReader(text)
	b, err := io.ReadAll(alloc, &r)
	if err != nil {
		t.Errorf("ReadAll() error = %s", errName(errCode(err)))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != text {
		t.Errorf("ReadAll() = %s, want %s", string(b), text)
	}
}

func TestReadAllEmpty(t *testing.T) {
	alloc := t.Allocator()
	r := strings.NewReader("")
	b, err := io.ReadAll(alloc, &r)
	if err != nil {
		t.Errorf("ReadAll() error = %s", errName(errCode(err)))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if len(b) != 0 {
		t.Errorf("ReadAll() = %d bytes, want 0", len(b))
	}
}

func TestReadAllLarge(t *testing.T) {
	// The data is larger than the first buffer of ReadAll, so ReadAll moves
	// to the intermediate slices and joins them at the end.
	var data [1500]byte
	want := fill(data[:])

	alloc := t.Allocator()
	r := chunkReader{s: want, n: 100}
	b, err := io.ReadAll(alloc, &r)
	if err != nil {
		t.Errorf("ReadAll() error = %s", errName(errCode(err)))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if len(b) != len(want) {
		t.Errorf("ReadAll() = %d bytes, want %d", len(b), len(want))
		return
	}
	if string(b) != want {
		t.Error("ReadAll() gave the wrong bytes")
	}
}

func TestReadAllError(t *testing.T) {
	// ReadAll gives the bytes it read before the error.
	alloc := t.Allocator()
	r := dataErrReader{s: "abc"}
	b, err := io.ReadAll(alloc, &r)
	defer mem.FreeSlice(alloc, b)

	if errCode(err) != errRead {
		t.Errorf("ReadAll() error = %s, want read failed", errName(errCode(err)))
	}
	if string(b) != "abc" {
		t.Errorf("ReadAll() = %s, want abc", string(b))
	}
}

func TestReadFull(t *testing.T) {
	// The reader gives two bytes at a time, so ReadFull makes several reads.
	r := chunkReader{s: text, n: 2}
	var buf [13]byte
	n, err := io.ReadFull(&r, buf[:])
	if n != len(text) || err != nil {
		t.Errorf("ReadFull() = %d, %s, want %d, nil", n, errName(errCode(err)), len(text))
	}
	if string(buf[:n]) != text {
		t.Errorf("ReadFull() = %s, want %s", string(buf[:n]), text)
	}
}

func TestReadFullShort(t *testing.T) {
	// The reader has fewer bytes than the buffer.
	r := strings.NewReader("hello")
	var buf [20]byte
	n, err := io.ReadFull(&r, buf[:])
	if n != 5 || errCode(err) != errUnexpectedEOF {
		t.Errorf("ReadFull() = %d, %s, want 5, ErrUnexpectedEOF", n, errName(errCode(err)))
	}
	if string(buf[:n]) != "hello" {
		t.Errorf("ReadFull() = %s, want hello", string(buf[:n]))
	}
}

func TestReadFullEOF(t *testing.T) {
	// The reader has no bytes at all, so the error is EOF, not ErrUnexpectedEOF.
	r := strings.NewReader("")
	var buf [5]byte
	n, err := io.ReadFull(&r, buf[:])
	if n != 0 || errCode(err) != errEOF {
		t.Errorf("ReadFull() = %d, %s, want 0, EOF", n, errName(errCode(err)))
	}
}

func TestReadFullEmptyBuffer(t *testing.T) {
	// An empty buffer needs no read, so the reader is never called.
	var r errReader
	var buf [1]byte
	n, err := io.ReadFull(&r, buf[:0])
	if n != 0 || err != nil {
		t.Errorf("ReadFull() = %d, %s, want 0, nil", n, errName(errCode(err)))
	}
}

func TestReadFullReadErr(t *testing.T) {
	// ReadFull reports the error of the reader as is.
	r := dataErrReader{s: "ab"}
	var buf [5]byte
	n, err := io.ReadFull(&r, buf[:])
	if n != 2 || errCode(err) != errRead {
		t.Errorf("ReadFull() = %d, %s, want 2, read failed", n, errName(errCode(err)))
	}
}

func TestWriteString(t *testing.T) {
	w := bufWriter{buf: make([]byte, 0, 32)}
	n, err := io.WriteString(&w, text)
	if n != len(text) || err != nil {
		t.Errorf("WriteString() = %d, %s, want %d, nil", n, errName(errCode(err)), len(text))
	}
	if w.String() != text {
		t.Errorf("WriteString() wrote %s, want %s", w.String(), text)
	}
}

func TestWriteStringErr(t *testing.T) {
	var w errWriter
	n, err := io.WriteString(&w, text)
	if n != 0 || errCode(err) != errWrite {
		t.Errorf("WriteString() = %d, %s, want 0, write failed", n, errName(errCode(err)))
	}
}
