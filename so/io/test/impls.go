// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/io"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

func TestDiscard(t *testing.T) {
	n, err := io.Discard.Write([]byte("hello"))
	if n != 5 || err != nil {
		t.Errorf("Write() = %d, %s, want 5, nil", n, errName(errCode(err)))
	}

	var w io.DiscardWriter
	n, err = w.WriteString("hello")
	if n != 5 || err != nil {
		t.Errorf("WriteString() = %d, %s, want 5, nil", n, errName(errCode(err)))
	}

	n, err = w.Write([]byte{})
	if n != 0 || err != nil {
		t.Errorf("Write() of no bytes = %d, %s, want 0, nil", n, errName(errCode(err)))
	}
}

func TestLimitReader(t *testing.T) {
	r := strings.NewReader("hello world")
	lr := io.LimitReader(&r, 5)
	var buf [11]byte

	n, err := lr.Read(buf[:])
	if n != 5 || err != nil {
		t.Errorf("Read() = %d, %s, want 5, nil", n, errName(errCode(err)))
	}
	if string(buf[:n]) != "hello" {
		t.Errorf("Read() = %s, want hello", string(buf[:n]))
	}
	if lr.N != 0 {
		t.Errorf("N = %d after the read, want 0", lr.N)
	}

	// The limit is spent, so the next read gives EOF.
	n, err = lr.Read(buf[:])
	if n != 0 || errCode(err) != errEOF {
		t.Errorf("Read() at the limit = %d, %s, want 0, EOF", n, errName(errCode(err)))
	}
}

func TestLimitReaderSteps(t *testing.T) {
	// Each read moves the limit by the number of the read bytes.
	r := strings.NewReader("hello world")
	lr := io.LimitReader(&r, 5)
	var buf [2]byte

	n, err := lr.Read(buf[:])
	if n != 2 || err != nil || lr.N != 3 {
		t.Errorf("Read() = %d, %s, N = %d, want 2, nil, 3", n, errName(errCode(err)), lr.N)
	}
	n, err = lr.Read(buf[:])
	if n != 2 || err != nil || lr.N != 1 {
		t.Errorf("Read() = %d, %s, N = %d, want 2, nil, 1", n, errName(errCode(err)), lr.N)
	}
	n, err = lr.Read(buf[:])
	if n != 1 || err != nil || lr.N != 0 {
		t.Errorf("Read() = %d, %s, N = %d, want 1, nil, 0", n, errName(errCode(err)), lr.N)
	}
}

func TestLimitReaderZero(t *testing.T) {
	r := strings.NewReader("hello")
	lr := io.LimitReader(&r, 0)
	var buf [5]byte
	n, err := lr.Read(buf[:])
	if n != 0 || errCode(err) != errEOF {
		t.Errorf("Read() = %d, %s, want 0, EOF", n, errName(errCode(err)))
	}
}

func TestLimitReaderPastEnd(t *testing.T) {
	// The limit is larger than the data, so the reader gives EOF first.
	r := strings.NewReader("hello")
	lr := io.LimitReader(&r, 20)
	var buf [10]byte

	n, err := lr.Read(buf[:])
	if n != 5 || err != nil {
		t.Errorf("Read() = %d, %s, want 5, nil", n, errName(errCode(err)))
	}
	if lr.N != 15 {
		t.Errorf("N = %d after the read, want 15", lr.N)
	}
	n, err = lr.Read(buf[:])
	if n != 0 || errCode(err) != errEOF {
		t.Errorf("Read() at the end = %d, %s, want 0, EOF", n, errName(errCode(err)))
	}
}

func TestLimitReaderCopy(t *testing.T) {
	// Copy sizes its buffer by the limit of a LimitedReader.
	r := strings.NewReader("hello world")
	lr := io.LimitReader(&r, 5)
	w := bufWriter{buf: make([]byte, 0, 16)}
	n, err := io.Copy(&w, &lr)
	if n != 5 || err != nil {
		t.Errorf("Copy() = %d, %s, want 5, nil", n, errName(errCode(err)))
	}
	if w.String() != "hello" {
		t.Errorf("Copy() wrote %s, want hello", w.String())
	}
}

func TestNopCloser(t *testing.T) {
	r := strings.NewReader("hello")
	nc := io.NewNopCloser(&r)
	var buf [5]byte

	n, err := nc.Read(buf[:])
	if n != 5 || err != nil {
		t.Errorf("Read() = %d, %s, want 5, nil", n, errName(errCode(err)))
	}
	if string(buf[:n]) != "hello" {
		t.Errorf("Read() = %s, want hello", string(buf[:n]))
	}
	if err := nc.Close(); err != nil {
		t.Errorf("Close() error = %s", errName(errCode(err)))
	}
	// Close does nothing, so the reader still works after it.
	n, err = nc.Read(buf[:])
	if n != 0 || errCode(err) != errEOF {
		t.Errorf("Read() after Close = %d, %s, want 0, EOF", n, errName(errCode(err)))
	}
}

func TestNopCloserInterface(t *testing.T) {
	r := strings.NewReader("hello")
	nc := io.NewNopCloser(&r)
	var rc io.ReadCloser = &nc
	var buf [5]byte
	n, err := rc.Read(buf[:])
	if n != 5 || err != nil {
		t.Errorf("Read() = %d, %s, want 5, nil", n, errName(errCode(err)))
	}
	if err := rc.Close(); err != nil {
		t.Errorf("Close() error = %s", errName(errCode(err)))
	}
}

// The data of the section reader tests. The data is 30 bytes long.
const sectionData = "a long sample data, 1234567890"

// A sectionCase is a test case of SectionReader.ReadAt.
type sectionCase struct {
	data   string
	off    int64 // the offset of the section
	n      int64 // the length of the section
	bufLen int   // the length of the read buffer
	at     int64 // the offset of the read
	want   string
	err    int
}

var sectionCases = []sectionCase{
	{"", 0, 10, 2, 0, "", errEOF},
	{sectionData, 0, 30, 0, 0, "", errNone},
	{sectionData, 30, 1, 1, 0, "", errEOF},
	{sectionData, 0, 32, 30, 0, sectionData, errNone},
	{sectionData, 0, 30, 15, 0, "a long sample d", errNone},
	{sectionData, 0, 30, 30, 0, sectionData, errNone},
	{sectionData, 0, 30, 15, 2, "long sample dat", errNone},
	{sectionData, 3, 30, 15, 2, "g sample data, ", errNone},
	{sectionData, 3, 15, 13, 2, "g sample data", errNone},
	{sectionData, 3, 15, 17, 2, "g sample data", errEOF},
	{sectionData, 0, 0, 0, -1, "", errEOF},
	{sectionData, 0, 0, 0, 1, "", errEOF},
}

func TestSectionReaderReadAt(t *testing.T) {
	var buf [32]byte
	for i, tc := range sectionCases {
		r := strings.NewReader(tc.data)
		s := io.NewSectionReader(&r, tc.off, tc.n)
		n, err := s.ReadAt(buf[:tc.bufLen], tc.at)
		if errCode(err) != tc.err {
			t.Errorf("case %d: ReadAt() error = %s, want %s",
				i, errName(errCode(err)), errName(tc.err))
			continue
		}
		if string(buf[:n]) != tc.want {
			t.Errorf("case %d: ReadAt() = %s, want %s", i, string(buf[:n]), tc.want)
		}
		out := s.Outer()
		if out.R != io.ReaderAt(&r) || out.Off != tc.off || out.N != tc.n {
			t.Errorf("case %d: Outer() = %d, %d, want %d, %d",
				i, out.Off, out.N, tc.off, tc.n)
		}
	}
}

func TestSectionReaderRead(t *testing.T) {
	// The section covers the middle of the data.
	r := strings.NewReader(sectionData)
	s := io.NewSectionReader(&r, 7, 6)
	var buf [4]byte

	n, err := s.Read(buf[:])
	if n != 4 || err != nil {
		t.Errorf("Read() = %d, %s, want 4, nil", n, errName(errCode(err)))
	}
	if string(buf[:n]) != "samp" {
		t.Errorf("Read() = %s, want samp", string(buf[:n]))
	}
	// The section has two bytes left, so the read stops at the end of it.
	n, err = s.Read(buf[:])
	if n != 2 || err != nil {
		t.Errorf("Read() = %d, %s, want 2, nil", n, errName(errCode(err)))
	}
	if string(buf[:n]) != "le" {
		t.Errorf("Read() = %s, want le", string(buf[:n]))
	}
	n, err = s.Read(buf[:])
	if n != 0 || errCode(err) != errEOF {
		t.Errorf("Read() at the end = %d, %s, want 0, EOF", n, errName(errCode(err)))
	}
}

func TestSectionReaderSeek(t *testing.T) {
	// The Seeker of a SectionReader behaves like the Seeker of a bytes.Reader.
	br := bytes.NewReader([]byte("foo"))
	sr := io.NewSectionReader(&br, 0, 3)
	whences := []int{io.SeekStart, io.SeekCurrent, io.SeekEnd}

	for _, whence := range whences {
		for offset := int64(-3); offset <= 4; offset++ {
			brOff, brErr := br.Seek(offset, whence)
			srOff, srErr := sr.Seek(offset, whence)
			if errCode(brErr) != errCode(srErr) || brOff != srOff {
				t.Errorf("whence %d, offset %d: bytes.Reader gives %d, %s; SectionReader gives %d, %s",
					whence, offset, brOff, errName(errCode(brErr)),
					srOff, errName(errCode(srErr)))
			}
		}
	}
}

func TestSectionReaderSeekPastEnd(t *testing.T) {
	// A seek past the end of the section succeeds, and the next read gives EOF.
	r := strings.NewReader("foo")
	sr := io.NewSectionReader(&r, 0, 3)
	pos, err := sr.Seek(100, io.SeekStart)
	if pos != 100 || err != nil {
		t.Errorf("Seek() = %d, %s, want 100, nil", pos, errName(errCode(err)))
	}
	var buf [10]byte
	n, err := sr.Read(buf[:])
	if n != 0 || errCode(err) != errEOF {
		t.Errorf("Read() = %d, %s, want 0, EOF", n, errName(errCode(err)))
	}
}

func TestSectionReaderSeekWhence(t *testing.T) {
	r := strings.NewReader("foo")
	sr := io.NewSectionReader(&r, 0, 3)
	pos, err := sr.Seek(0, 42)
	if pos != 0 || errCode(err) != errWhence {
		t.Errorf("Seek() = %d, %s, want 0, ErrWhence", pos, errName(errCode(err)))
	}
}

func TestSectionReaderSeekBase(t *testing.T) {
	// The offsets of a seek are relative to the start of the section, not to
	// the start of the data.
	r := strings.NewReader(sectionData)
	sr := io.NewSectionReader(&r, 7, 6)
	pos, err := sr.Seek(2, io.SeekStart)
	if pos != 2 || err != nil {
		t.Errorf("Seek() = %d, %s, want 2, nil", pos, errName(errCode(err)))
		return
	}
	var buf [4]byte
	n, err := sr.Read(buf[:])
	if n != 4 || err != nil {
		t.Errorf("Read() = %d, %s, want 4, nil", n, errName(errCode(err)))
	}
	if string(buf[:n]) != "mple" {
		t.Errorf("Read() = %s, want mple", string(buf[:n]))
	}
}

func TestSectionReaderSize(t *testing.T) {
	r := strings.NewReader(sectionData)
	sr := io.NewSectionReader(&r, 0, 30)
	if sr.Size() != 30 {
		t.Errorf("Size() = %d, want 30", sr.Size())
	}

	empty := strings.NewReader("")
	sr = io.NewSectionReader(&empty, 0, 0)
	if sr.Size() != 0 {
		t.Errorf("Size() of an empty section = %d, want 0", sr.Size())
	}

	part := strings.NewReader(sectionData)
	sr = io.NewSectionReader(&part, 7, 6)
	if sr.Size() != 6 {
		t.Errorf("Size() of a part = %d, want 6", sr.Size())
	}
}

func TestSectionReaderMax(t *testing.T) {
	// The length of the section overflows the offset, so the section ends at
	// the largest offset.
	const maxint64 = int64(^uint64(0) >> 1)
	r := strings.NewReader("abcdef")
	sr := io.NewSectionReader(&r, 3, maxint64)
	var buf [3]byte

	n, err := sr.Read(buf[:])
	if n != 3 || err != nil {
		t.Errorf("Read() = %d, %s, want 3, nil", n, errName(errCode(err)))
	}
	if string(buf[:n]) != "def" {
		t.Errorf("Read() = %s, want def", string(buf[:n]))
	}
	n, err = sr.Read(buf[:])
	if n != 0 || errCode(err) != errEOF {
		t.Errorf("Read() at the end = %d, %s, want 0, EOF", n, errName(errCode(err)))
	}
	out := sr.Outer()
	if out.R != io.ReaderAt(&r) || out.Off != 3 || out.N != maxint64 {
		t.Errorf("Outer() = %d, %d, want 3, %d", out.Off, out.N, maxint64)
	}
}
