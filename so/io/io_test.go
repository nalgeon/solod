// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io_test

import (
	stdio "io"
	stdstrings "strings"
	"testing"

	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
)

// The classes of the errors the fuzzers compare. An error of so/io and the
// matching error of io are different values, so the fuzzers compare the class
// of an error instead of the error.
const (
	kindNil = iota
	kindEOF
	kindUnexpectedEOF
	kindOther
)

// errKind returns the class of the error.
func errKind(err error) int {
	switch err {
	case nil:
		return kindNil
	case io.EOF, stdio.EOF:
		return kindEOF
	case io.ErrUnexpectedEOF, stdio.ErrUnexpectedEOF:
		return kindUnexpectedEOF
	}
	return kindOther
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
	n := copy(p, r.s)
	r.s = r.s[n:]
	return n, nil
}

// stdChunkReader is chunkReader for the io package of Go. The two packages
// have different EOF values, so a reader of the reference side must give the
// EOF of Go.
type stdChunkReader struct {
	s string
	n int
}

func (r *stdChunkReader) Read(p []byte) (int, error) {
	if len(r.s) == 0 {
		return 0, stdio.EOF
	}
	if len(p) > r.n {
		p = p[:r.n]
	}
	n := copy(p, r.s)
	r.s = r.s[n:]
	return n, nil
}

func TestCopyBufferPanic(t *testing.T) {
	// So has no recover, so only a Go test checks a panic.
	t.Run("nil buffer", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("want panic")
			}
		}()
		var w io.DiscardWriter
		r := strings.NewReader("hello")
		_, _ = io.CopyBuffer(&w, &r, nil)
	})
	t.Run("empty buffer", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("want panic")
			}
		}()
		var w io.DiscardWriter
		r := strings.NewReader("hello")
		_, _ = io.CopyBuffer(&w, &r, []byte{})
	})
}

func FuzzReadAll(f *testing.F) {
	// ReadAll grows through a chain of the intermediate slices, so the size of
	// the input and the size of a read decide the shape of the chain.
	f.Add([]byte(""), uint8(1))
	f.Add([]byte("hello, world."), uint8(1))
	f.Add([]byte("hello, world."), uint8(200))
	f.Add(make([]byte, 512), uint8(64))
	f.Add(make([]byte, 4096), uint8(255))

	f.Fuzz(func(t *testing.T, data []byte, chunk uint8) {
		r := chunkReader{s: string(data), n: int(chunk)%255 + 1}
		got, err := io.ReadAll(mem.System, &r)
		defer mem.FreeSlice(mem.System, got)

		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if string(got) != string(data) {
			t.Fatalf("ReadAll() = %d bytes, want %d bytes", len(got), len(data))
		}
	})
}

func FuzzCopy(f *testing.F) {
	// Copy must move the same bytes as the Copy of Go, whatever the number of
	// the bytes one read gives.
	f.Add([]byte(""), uint8(1))
	f.Add([]byte("hello, world."), uint8(1))
	f.Add([]byte("hello, world."), uint8(5))
	f.Add(make([]byte, 9000), uint8(255))

	f.Fuzz(func(t *testing.T, data []byte, chunk uint8) {
		size := int(chunk)%255 + 1

		src := chunkReader{s: string(data), n: size}
		var dst stdstrings.Builder
		n, err := io.Copy(&dst, &src)
		if err != nil {
			t.Fatalf("Copy() error = %v", err)
		}

		refSrc := stdChunkReader{s: string(data), n: size}
		var refDst stdstrings.Builder
		refN, refErr := stdio.Copy(&refDst, &refSrc)
		if refErr != nil {
			t.Fatalf("std Copy() error = %v", refErr)
		}

		if n != refN {
			t.Fatalf("Copy() = %d, want %d", n, refN)
		}
		if dst.String() != refDst.String() {
			t.Fatal("Copy() wrote the wrong bytes")
		}
	})
}

func FuzzReadFull(f *testing.F) {
	f.Add([]byte("hello, world."), uint8(1), uint8(13))
	f.Add([]byte("hello, world."), uint8(3), uint8(20))
	f.Add([]byte(""), uint8(1), uint8(4))

	f.Fuzz(func(t *testing.T, data []byte, chunk, bufLen uint8) {
		size := int(chunk)%255 + 1

		r := chunkReader{s: string(data), n: size}
		buf := make([]byte, bufLen)
		n, err := io.ReadFull(&r, buf)

		refR := stdChunkReader{s: string(data), n: size}
		refBuf := make([]byte, bufLen)
		refN, refErr := stdio.ReadFull(&refR, refBuf)

		if n != refN || errKind(err) != errKind(refErr) {
			t.Fatalf("ReadFull() = %d, %v, want %d, %v", n, err, refN, refErr)
		}
		if string(buf[:n]) != string(refBuf[:refN]) {
			t.Fatal("ReadFull() read the wrong bytes")
		}
	})
}
