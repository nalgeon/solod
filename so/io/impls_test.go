// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io_test

import (
	"testing"

	"solod.dev/so/bytes"
	. "solod.dev/so/io"
	"solod.dev/so/strings"
)

func TestSectionReader_ReadAt(t *testing.T) {
	dat := "a long sample data, 1234567890"
	tests := []struct {
		data   string
		off    int
		n      int
		bufLen int
		at     int
		exp    string
		err    error
	}{
		{data: "", off: 0, n: 10, bufLen: 2, at: 0, exp: "", err: EOF},
		{data: dat, off: 0, n: len(dat), bufLen: 0, at: 0, exp: "", err: nil},
		{data: dat, off: len(dat), n: 1, bufLen: 1, at: 0, exp: "", err: EOF},
		{data: dat, off: 0, n: len(dat) + 2, bufLen: len(dat), at: 0, exp: dat, err: nil},
		{data: dat, off: 0, n: len(dat), bufLen: len(dat) / 2, at: 0, exp: dat[:len(dat)/2], err: nil},
		{data: dat, off: 0, n: len(dat), bufLen: len(dat), at: 0, exp: dat, err: nil},
		{data: dat, off: 0, n: len(dat), bufLen: len(dat) / 2, at: 2, exp: dat[2 : 2+len(dat)/2], err: nil},
		{data: dat, off: 3, n: len(dat), bufLen: len(dat) / 2, at: 2, exp: dat[5 : 5+len(dat)/2], err: nil},
		{data: dat, off: 3, n: len(dat) / 2, bufLen: len(dat)/2 - 2, at: 2, exp: dat[5 : 5+len(dat)/2-2], err: nil},
		{data: dat, off: 3, n: len(dat) / 2, bufLen: len(dat)/2 + 2, at: 2, exp: dat[5 : 5+len(dat)/2-2], err: EOF},
		{data: dat, off: 0, n: 0, bufLen: 0, at: -1, exp: "", err: EOF},
		{data: dat, off: 0, n: 0, bufLen: 0, at: 1, exp: "", err: EOF},
	}
	for i, tt := range tests {
		r := strings.NewReader(tt.data)
		s := NewSectionReader(&r, int64(tt.off), int64(tt.n))
		buf := make([]byte, tt.bufLen)
		if n, err := s.ReadAt(buf, int64(tt.at)); n != len(tt.exp) || string(buf[:n]) != tt.exp || !errEqual(err, tt.err) {
			t.Fatalf("%d: ReadAt(%d) = %q, %v; expected %q, %v", i, tt.at, buf[:n], err, tt.exp, tt.err)
		}
		if off := s.Outer(); off.R != &r || off.Off != int64(tt.off) || off.N != int64(tt.n) {
			t.Fatalf("%d: Outer() = %v, %d, %d; expected %v, %d, %d", i, off.R, off.Off, off.N, &r, tt.off, tt.n)
		}
	}
}

func TestSectionReader_Seek(t *testing.T) {
	// Verifies that NewSectionReader's Seeker behaves like bytes.NewReader (which is like strings.NewReader)
	br := bytes.NewReader([]byte("foo"))
	sr := NewSectionReader(&br, 0, int64(len("foo")))

	for _, whence := range []int{SeekStart, SeekCurrent, SeekEnd} {
		for offset := int64(-3); offset <= 4; offset++ {
			brOff, brErr := br.Seek(offset, whence)
			srOff, srErr := sr.Seek(offset, whence)
			if (brErr != nil) != (srErr != nil) || brOff != srOff {
				t.Errorf("For whence %d, offset %d: bytes.Reader.Seek = (%v, %v) != SectionReader.Seek = (%v, %v)",
					whence, offset, brOff, brErr, srErr, srOff)
			}
		}
	}

	// And verify we can just seek past the end and get an EOF
	got, err := sr.Seek(100, SeekStart)
	if err != nil || got != 100 {
		t.Errorf("Seek = %v, %v; want 100, nil", got, err)
	}

	n, err := sr.Read(make([]byte, 10))
	if n != 0 || err != EOF {
		t.Errorf("Read = %v, %v; want 0, EOF", n, err)
	}
}

func TestSectionReader_Size(t *testing.T) {
	tests := []struct {
		data string
		want int64
	}{
		{"a long sample data, 1234567890", 30},
		{"", 0},
	}

	for _, tt := range tests {
		r := strings.NewReader(tt.data)
		sr := NewSectionReader(&r, 0, int64(len(tt.data)))
		if got := sr.Size(); got != tt.want {
			t.Errorf("Size = %v; want %v", got, tt.want)
		}
	}
}

func TestSectionReader_Max(t *testing.T) {
	r := strings.NewReader("abcdef")
	const maxint64 = 1<<63 - 1
	sr := NewSectionReader(&r, 3, maxint64)
	n, err := sr.Read(make([]byte, 3))
	if n != 3 || err != nil {
		t.Errorf("Read = %v %v, want 3, nil", n, err)
	}
	n, err = sr.Read(make([]byte, 3))
	if n != 0 || !errEqual(err, EOF) {
		t.Errorf("Read = %v, %v, want 0, EOF", n, err)
	}
	if off := sr.Outer(); off.R != &r || off.Off != 3 || off.N != maxint64 {
		t.Fatalf("Outer = %v, %d, %d; expected %v, %d, %d", off.R, off.Off, off.N, &r, 3, int64(maxint64))
	}
}
