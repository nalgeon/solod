// Copyright 2010 The Go Authors. All rights reserved.
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

// cut3 splits data into three parts at the two given points.
func cut3(data []byte, a, b uint8) (string, string, string) {
	i := int(a) % (len(data) + 1)
	j := int(b) % (len(data) + 1)
	if i > j {
		i, j = j, i
	}
	return string(data[:i]), string(data[i:j]), string(data[j:])
}

func FuzzMultiReader(f *testing.F) {
	f.Add([]byte("foo bar"), uint8(3), uint8(3), uint8(4))
	f.Add([]byte("foo bar"), uint8(0), uint8(7), uint8(1))
	f.Add([]byte(""), uint8(0), uint8(0), uint8(2))

	f.Fuzz(func(t *testing.T, data []byte, cutA, cutB, bufLen uint8) {
		p1, p2, p3 := cut3(data, cutA, cutB)

		r1, r2, r3 := strings.NewReader(p1), strings.NewReader(p2), strings.NewReader(p3)
		mr := io.NewMultiReader(&r1, &r2, &r3)
		ref := stdio.MultiReader(
			stdstrings.NewReader(p1),
			stdstrings.NewReader(p2),
			stdstrings.NewReader(p3),
		)

		buf := make([]byte, int(bufLen)%16+1)
		refBuf := make([]byte, len(buf))
		for range 1000 {
			n, err := mr.Read(buf)
			refN, refErr := ref.Read(refBuf)
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

func FuzzMultiReaderWriteTo(f *testing.F) {
	f.Add([]byte("foo bar"), uint8(3), uint8(3))
	f.Add([]byte(""), uint8(0), uint8(0))

	f.Fuzz(func(t *testing.T, data []byte, cutA, cutB uint8) {
		p1, p2, p3 := cut3(data, cutA, cutB)

		r1, r2, r3 := strings.NewReader(p1), strings.NewReader(p2), strings.NewReader(p3)
		mr := io.NewMultiReader(&r1, &r2, &r3)
		var dst stdstrings.Builder
		n, err := mr.WriteTo(&dst)
		if err != nil {
			t.Fatalf("WriteTo() error = %v", err)
		}

		if n != int64(len(data)) {
			t.Fatalf("WriteTo() = %d, want %d", n, len(data))
		}
		if dst.String() != string(data) {
			t.Fatal("WriteTo() wrote the wrong bytes")
		}
	})
}

func FuzzMultiWriter(f *testing.F) {
	f.Add([]byte("foo bar"), uint8(3), uint8(3))
	f.Add([]byte(""), uint8(0), uint8(0))

	f.Fuzz(func(t *testing.T, data []byte, cutA, cutB uint8) {
		p1, p2, p3 := cut3(data, cutA, cutB)

		var w1, w2 stdstrings.Builder
		mw := io.NewMultiWriter(&w1, &w2)
		for _, part := range []string{p1, p2, p3} {
			n, err := io.WriteString(&mw, part)
			if err != nil {
				t.Fatalf("WriteString() error = %v", err)
			}
			if n != len(part) {
				t.Fatalf("WriteString() = %d, want %d", n, len(part))
			}
		}

		if w1.String() != string(data) {
			t.Fatal("the first writer holds the wrong bytes")
		}
		if w2.String() != string(data) {
			t.Fatal("the second writer holds the wrong bytes")
		}
	})
}
