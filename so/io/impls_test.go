// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io_test

import (
	stdio "io"
	stdstrings "strings"
	"testing"

	"solod.dev/so/io"
	"solod.dev/so/strings"
)

func FuzzLimitReader(f *testing.F) {
	f.Add([]byte("hello, world."), int8(5), uint8(3))
	f.Add([]byte("hello, world."), int8(0), uint8(4))
	f.Add([]byte("hello, world."), int8(-1), uint8(4))
	f.Add([]byte("hello, world."), int8(100), uint8(255))

	f.Fuzz(func(t *testing.T, data []byte, limit int8, bufLen uint8) {
		src := chunkReader{s: string(data), n: 3}
		lr := io.LimitReader(&src, int64(limit))

		refSrc := stdChunkReader{s: string(data), n: 3}
		refLr := stdio.LimitReader(&refSrc, int64(limit))

		buf := make([]byte, int(bufLen)%32+1)
		refBuf := make([]byte, len(buf))
		for range 100 {
			n, err := lr.Read(buf)
			refN, refErr := refLr.Read(refBuf)
			if n != refN || errKind(err) != errKind(refErr) {
				t.Fatalf("Read() = %d, %v, want %d, %v", n, err, refN, refErr)
			}
			if string(buf[:n]) != string(refBuf[:refN]) {
				t.Fatal("Read() read the wrong bytes")
			}
			if err != nil {
				break
			}
		}
	})
}

func FuzzSectionReader(f *testing.F) {
	// The section length n is never negative: NewSectionReader of Go overflows
	// on a negative n, and So copies that arithmetic.
	f.Add([]byte("a long sample data, 1234567890"), int8(0), uint8(30), int8(0), uint8(30), uint8(0))
	f.Add([]byte("a long sample data, 1234567890"), int8(3), uint8(10), int8(2), uint8(6), uint8(1))
	f.Add([]byte("a long sample data, 1234567890"), int8(-4), uint8(0), int8(0), uint8(4), uint8(2))
	f.Add([]byte(""), int8(0), uint8(0), int8(0), uint8(1), uint8(0))

	f.Fuzz(func(t *testing.T, data []byte, off int8, n uint8, at int8, bufLen, whence uint8) {
		src := strings.NewReader(string(data))
		sec := io.NewSectionReader(&src, int64(off), int64(n))
		ref := stdio.NewSectionReader(stdstrings.NewReader(string(data)), int64(off), int64(n))

		if sec.Size() != ref.Size() {
			t.Fatalf("Size() = %d, want %d", sec.Size(), ref.Size())
		}

		buf := make([]byte, int(bufLen)%32)
		refBuf := make([]byte, len(buf))

		gotN, gotErr := sec.ReadAt(buf, int64(at))
		wantN, wantErr := ref.ReadAt(refBuf, int64(at))
		if gotN != wantN || errKind(gotErr) != errKind(wantErr) {
			t.Fatalf("ReadAt() = %d, %v, want %d, %v", gotN, gotErr, wantN, wantErr)
		}
		if string(buf[:gotN]) != string(refBuf[:wantN]) {
			t.Fatal("ReadAt() read the wrong bytes")
		}

		gotPos, gotErr := sec.Seek(int64(at), int(whence)%4)
		wantPos, wantErr := ref.Seek(int64(at), int(whence)%4)
		if gotPos != wantPos || (gotErr == nil) != (wantErr == nil) {
			t.Fatalf("Seek() = %d, %v, want %d, %v", gotPos, gotErr, wantPos, wantErr)
		}

		for range 100 {
			gotN, gotErr = sec.Read(buf)
			wantN, wantErr = ref.Read(refBuf)
			if gotN != wantN || errKind(gotErr) != errKind(wantErr) {
				t.Fatalf("Read() = %d, %v, want %d, %v", gotN, gotErr, wantN, wantErr)
			}
			if string(buf[:gotN]) != string(refBuf[:wantN]) {
				t.Fatal("Read() read the wrong bytes")
			}
			if gotErr != nil {
				break
			}
		}
	})
}
