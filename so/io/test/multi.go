// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

// A readCase is one Read call of the MultiReader tests.
type readCase struct {
	size int    // the number of the requested bytes
	want string // the expected bytes
	err  int    // the expected error
}

var multiCases1 = []readCase{
	{2, "fo", errNone},
	{5, "o ", errNone},
	{5, "bar", errNone},
	{5, "", errEOF},
}

var multiCases2 = []readCase{
	{4, "foo ", errNone},
	{1, "b", errNone},
	{3, "ar", errNone},
	{1, "", errEOF},
}

var multiCases3 = []readCase{
	{5, "foo ", errNone},
}

// readMulti runs the read cases on a MultiReader over "foo ", "" and "bar".
// Each read takes the part of the buffer after the bytes of the reads before
// it, the way a caller that collects the whole input does.
func readMulti(t *testing.T, cases []readCase) {
	r1 := strings.NewReader("foo ")
	r2 := strings.NewReader("")
	r3 := strings.NewReader("bar")
	mr := io.NewMultiReader(&r1, &r2, &r3)

	var buf [20]byte
	pos := 0
	for i, tc := range cases {
		n, err := mr.Read(buf[pos : pos+tc.size])
		if errCode(err) != tc.err {
			t.Errorf("read %d: error = %s, want %s",
				i, errName(errCode(err)), errName(tc.err))
			continue
		}
		if n != len(tc.want) {
			t.Errorf("read %d: = %d bytes, want %d", i, n, len(tc.want))
			continue
		}
		if string(buf[pos:pos+n]) != tc.want {
			t.Errorf("read %d: = %s, want %s", i, string(buf[pos:pos+n]), tc.want)
		}
		pos += n
	}
}

func TestMultiReader(t *testing.T) {
	readMulti(t, multiCases1)
}

func TestMultiReaderWholeParts(t *testing.T) {
	readMulti(t, multiCases2)
}

func TestMultiReaderPartOnly(t *testing.T) {
	// One read never crosses the border between two readers.
	readMulti(t, multiCases3)
}

func TestMultiReaderEmpty(t *testing.T) {
	mr := io.NewMultiReader()
	var buf [4]byte
	n, err := mr.Read(buf[:])
	if n != 0 || errCode(err) != errEOF {
		t.Errorf("Read() = %d, %s, want 0, EOF", n, errName(errCode(err)))
	}
}

func TestMultiReaderCopy(t *testing.T) {
	r1 := strings.NewReader("hello ")
	r2 := strings.NewReader("world")
	mr := io.NewMultiReader(&r1, &r2)
	w := bufWriter{buf: make([]byte, 0, 16)}

	n, err := io.Copy(&w, &mr)
	if n != 11 || err != nil {
		t.Errorf("Copy() = %d, %s, want 11, nil", n, errName(errCode(err)))
	}
	if w.String() != "hello world" {
		t.Errorf("Copy() wrote %s, want hello world", w.String())
	}
}

func TestMultiReaderReadErr(t *testing.T) {
	// A reader that fails with an error other than EOF stops the read.
	r1 := strings.NewReader("foo")
	var r2 errReader
	r3 := strings.NewReader("bar")
	mr := io.NewMultiReader(&r1, &r2, &r3)

	var buf [8]byte
	n, err := mr.Read(buf[:])
	if n != 3 || err != nil {
		t.Errorf("Read() = %d, %s, want 3, nil", n, errName(errCode(err)))
	}
	n, err = mr.Read(buf[:])
	if n != 0 || errCode(err) != errRead {
		t.Errorf("Read() = %d, %s, want 0, read failed", n, errName(errCode(err)))
	}
}

func TestMultiReaderNested(t *testing.T) {
	// A chain of the nested readers gives the bytes of the innermost reader.
	// NewMultiReader keeps the given list without copying it, so the list must
	// outlive the reader. A list built in the loop body dies at the end of the
	// iteration. So the chain takes the slices of one array, which lives until
	// the end of the test.
	r0 := strings.NewReader("foo bar")
	var lists [8]io.Reader
	var chain [8]io.MultiReader
	lists[0] = &r0
	chain[0] = io.NewMultiReader(lists[0:1]...)
	for i := 1; i < len(chain); i++ {
		lists[i] = &chain[i-1]
		chain[i] = io.NewMultiReader(lists[i : i+1]...)
	}

	var buf [8]byte
	n, err := io.ReadFull(&chain[len(chain)-1], buf[:7])
	if n != 7 || err != nil {
		t.Errorf("ReadFull() = %d, %s, want 7, nil", n, errName(errCode(err)))
	}
	if string(buf[:n]) != "foo bar" {
		t.Errorf("ReadFull() = %s, want foo bar", string(buf[:n]))
	}
}

func TestMultiReaderSingleByteWithEOF(t *testing.T) {
	// A reader that gives one byte and EOF at once must not repeat forever.
	r1 := byteAndEOFReader{b: 'a'}
	r2 := byteAndEOFReader{b: 'b'}
	mr := io.NewMultiReader(&r1, &r2)
	lr := io.LimitReader(&mr, 10)

	alloc := t.Allocator()
	b, err := io.ReadAll(alloc, &lr)
	if err != nil {
		t.Errorf("ReadAll() error = %s", errName(errCode(err)))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != "ab" {
		t.Errorf("ReadAll() = %s, want ab", string(b))
	}
}

func TestMultiReaderFinalEOF(t *testing.T) {
	// The last reader of the chain gives its byte and EOF at once. The
	// MultiReader keeps the EOF, instead of giving no bytes and EOF later.
	var empty []byte
	br := bytes.NewReader(empty)
	r := byteAndEOFReader{b: 'a'}
	mr := io.NewMultiReader(&br, &r)

	var buf [2]byte
	n, err := mr.Read(buf[:])
	if n != 1 || errCode(err) != errEOF {
		t.Errorf("Read() = %d, %s, want 1, EOF", n, errName(errCode(err)))
	}
	if string(buf[:n]) != "a" {
		t.Errorf("Read() = %s, want a", string(buf[:n]))
	}
}

func TestInterleavedMultiReader(t *testing.T) {
	// The outer reader takes the readers of the inner reader, so the two share
	// one list. A read of the inner reader after the outer reader consumed the
	// first reader must still work.
	r1 := strings.NewReader("123")
	r2 := strings.NewReader("45678")
	mr1 := io.NewMultiReader(&r1, &r2)
	mr2 := io.NewMultiReader(&mr1)

	var buf [4]byte
	n, err := io.ReadFull(&mr2, buf[:])
	if string(buf[:n]) != "1234" || err != nil {
		t.Errorf("ReadFull(mr2) = %s, %s, want 1234, nil",
			string(buf[:n]), errName(errCode(err)))
	}

	n, err = io.ReadFull(&mr1, buf[:])
	if string(buf[:n]) != "5678" || err != nil {
		t.Errorf("ReadFull(mr1) = %s, %s, want 5678, nil",
			string(buf[:n]), errName(errCode(err)))
	}
}

func TestMultiReaderWriteTo(t *testing.T) {
	sr1 := strings.NewReader("foo ")
	sr2 := strings.NewReader("")
	sr3 := strings.NewReader("bar")
	mr2 := io.NewMultiReader(&sr2, &sr3)
	mr := io.NewMultiReader(&sr1, &mr2)
	w := bufWriter{buf: make([]byte, 0, 16)}

	n, err := mr.WriteTo(&w)
	if n != 7 || err != nil {
		t.Errorf("WriteTo() = %d, %s, want 7, nil", n, errName(errCode(err)))
	}
	if w.String() != "foo bar" {
		t.Errorf("WriteTo() wrote %s, want foo bar", w.String())
	}

	// The readers are spent, so the next call writes nothing.
	n, err = mr.WriteTo(&w)
	if n != 0 || err != nil {
		t.Errorf("WriteTo() at the end = %d, %s, want 0, nil", n, errName(errCode(err)))
	}
}

func TestMultiReaderWriteToResume(t *testing.T) {
	// The second reader fails at the first read. WriteTo keeps the failed
	// reader at the front of the list, so the next call takes it again.
	r1 := strings.NewReader("foo")
	r2 := flakyReader{s: "bar"}
	mr := io.NewMultiReader(&r1, &r2)
	w := bufWriter{buf: make([]byte, 0, 16)}

	n, err := mr.WriteTo(&w)
	if n != 3 || errCode(err) != errRead {
		t.Errorf("WriteTo() = %d, %s, want 3, read failed", n, errName(errCode(err)))
	}
	if w.String() != "foo" {
		t.Errorf("WriteTo() wrote %s, want foo", w.String())
	}

	n, err = mr.WriteTo(&w)
	if n != 3 || err != nil {
		t.Errorf("WriteTo() again = %d, %s, want 3, nil", n, errName(errCode(err)))
	}
	if w.String() != "foobar" {
		t.Errorf("WriteTo() wrote %s, want foobar", w.String())
	}
}

func TestMultiWriter(t *testing.T) {
	w1 := bufWriter{buf: make([]byte, 0, 16)}
	w2 := bufWriter{buf: make([]byte, 0, 16)}
	mw := io.NewMultiWriter(&w1, &w2)

	n, err := mw.Write([]byte("hello"))
	if n != 5 || err != nil {
		t.Errorf("Write() = %d, %s, want 5, nil", n, errName(errCode(err)))
	}
	if w1.String() != "hello" || w2.String() != "hello" {
		t.Errorf("Write() wrote %s and %s, want hello and hello", w1.String(), w2.String())
	}
}

func TestMultiWriterCopy(t *testing.T) {
	r := strings.NewReader(text)
	w1 := bufWriter{buf: make([]byte, 0, 16)}
	w2 := bufWriter{buf: make([]byte, 0, 16)}
	mw := io.NewMultiWriter(&w1, &w2)

	n, err := io.Copy(&mw, &r)
	if n != int64(len(text)) || err != nil {
		t.Errorf("Copy() = %d, %s, want %d, nil", n, errName(errCode(err)), len(text))
	}
	if w1.String() != text || w2.String() != text {
		t.Errorf("Copy() wrote %s and %s, want %s", w1.String(), w2.String(), text)
	}
}

func TestMultiWriterEmpty(t *testing.T) {
	mw := io.NewMultiWriter()
	n, err := mw.Write([]byte("hello"))
	if n != 5 || err != nil {
		t.Errorf("Write() = %d, %s, want 5, nil", n, errName(errCode(err)))
	}
}

func TestMultiWriterString(t *testing.T) {
	w1 := bufWriter{buf: make([]byte, 0, 16)}
	w2 := bufWriter{buf: make([]byte, 0, 16)}
	mw := io.NewMultiWriter(&w1, &w2)

	n, err := mw.WriteString("hello")
	if n != 5 || err != nil {
		t.Errorf("WriteString() = %d, %s, want 5, nil", n, errName(errCode(err)))
	}
	if w1.String() != "hello" || w2.String() != "hello" {
		t.Errorf("WriteString() wrote %s and %s, want hello and hello",
			w1.String(), w2.String())
	}
}

func TestMultiWriterShortWrite(t *testing.T) {
	// The first writer accepts half of the bytes, so the write stops and the
	// second writer never runs.
	var w1 halfWriter
	w2 := calledWriter{}
	mw := io.NewMultiWriter(&w1, &w2)

	var buf [100]byte
	n, err := mw.Write(buf[:])
	if n != 50 || errCode(err) != errShortWrite {
		t.Errorf("Write() = %d, %s, want 50, ErrShortWrite", n, errName(errCode(err)))
	}
	if w2.called {
		t.Error("Write() called the second writer")
	}
}

func TestMultiWriterError(t *testing.T) {
	// The first writer fails, so the write stops and the second writer never
	// runs.
	var w1 errWriter
	w2 := calledWriter{}
	mw := io.NewMultiWriter(&w1, &w2)

	n, err := mw.Write([]byte("hello"))
	if n != 0 || errCode(err) != errWrite {
		t.Errorf("Write() = %d, %s, want 0, write failed", n, errName(errCode(err)))
	}
	if w2.called {
		t.Error("Write() called the second writer")
	}
}

func TestMultiWriterStringError(t *testing.T) {
	var w1 errWriter
	w2 := calledWriter{}
	mw := io.NewMultiWriter(&w1, &w2)

	n, err := mw.WriteString("hello")
	if n != 0 || errCode(err) != errWrite {
		t.Errorf("WriteString() = %d, %s, want 0, write failed", n, errName(errCode(err)))
	}
	if w2.called {
		t.Error("WriteString() called the second writer")
	}
}

func TestMultiWriterStringShortWrite(t *testing.T) {
	var w1 shortWriter
	w2 := calledWriter{}
	mw := io.NewMultiWriter(&w1, &w2)

	n, err := mw.WriteString("hello")
	if n != 4 || errCode(err) != errShortWrite {
		t.Errorf("WriteString() = %d, %s, want 4, ErrShortWrite", n, errName(errCode(err)))
	}
	if w2.called {
		t.Error("WriteString() called the second writer")
	}
}
