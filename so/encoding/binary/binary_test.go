// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binary

import (
	stdbinary "encoding/binary"
	"testing"
)

// soOrder is the byte order of this package.
type soOrder interface {
	ByteOrder
	AppendByteOrder
}

// stdOrder is the byte order of Go's encoding/binary.
type stdOrder interface {
	stdbinary.ByteOrder
	stdbinary.AppendByteOrder
}

func FuzzByteOrder(f *testing.F) {
	f.Add([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}, uint64(0x0123456789abcdef))
	f.Add([]byte{0, 0, 0, 0, 0, 0, 0, 0}, uint64(0))
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, ^uint64(0))
	f.Add([]byte{0xaa}, uint64(0x8000000000000000))

	f.Fuzz(func(t *testing.T, b []byte, v uint64) {
		var src [8]byte
		copy(src[:], b)

		checkRead(t, "LittleEndian", &LittleEndian, stdbinary.LittleEndian, src)
		checkRead(t, "BigEndian", &BigEndian, stdbinary.BigEndian, src)
		checkWrite(t, "LittleEndian", &LittleEndian, stdbinary.LittleEndian, v)
		checkWrite(t, "BigEndian", &BigEndian, stdbinary.BigEndian, v)
		checkAppend(t, "LittleEndian", &LittleEndian, stdbinary.LittleEndian, v)
		checkAppend(t, "BigEndian", &BigEndian, stdbinary.BigEndian, v)
	})
}

// checkRead compares Uint16, Uint32 and Uint64 against Go's encoding/binary.
func checkRead(t *testing.T, name string, so soOrder, std stdOrder, src [8]byte) {
	if got, want := so.Uint16(src[:]), std.Uint16(src[:]); got != want {
		t.Errorf("%s.Uint16(%x) = %#x, want %#x", name, src, got, want)
	}
	if got, want := so.Uint32(src[:]), std.Uint32(src[:]); got != want {
		t.Errorf("%s.Uint32(%x) = %#x, want %#x", name, src, got, want)
	}
	if got, want := so.Uint64(src[:]), std.Uint64(src[:]); got != want {
		t.Errorf("%s.Uint64(%x) = %#x, want %#x", name, src, got, want)
	}
}

// checkWrite compares PutUint16, PutUint32 and PutUint64 against Go's
// encoding/binary.
func checkWrite(t *testing.T, name string, so soOrder, std stdOrder, v uint64) {
	var got, want [8]byte

	so.PutUint16(got[:2], uint16(v))
	std.PutUint16(want[:2], uint16(v))
	if got != want {
		t.Errorf("%s.PutUint16(%#x) = %x, want %x", name, uint16(v), got[:2], want[:2])
	}

	got, want = [8]byte{}, [8]byte{}
	so.PutUint32(got[:4], uint32(v))
	std.PutUint32(want[:4], uint32(v))
	if got != want {
		t.Errorf("%s.PutUint32(%#x) = %x, want %x", name, uint32(v), got[:4], want[:4])
	}

	got, want = [8]byte{}, [8]byte{}
	so.PutUint64(got[:], v)
	std.PutUint64(want[:], v)
	if got != want {
		t.Errorf("%s.PutUint64(%#x) = %x, want %x", name, v, got[:], want[:])
	}
}

// checkAppend compares AppendUint16, AppendUint32 and AppendUint64 against
// Go's encoding/binary. It appends to a non-empty slice, so it also checks
// that Append keeps the bytes of the input slice.
func checkAppend(t *testing.T, name string, so soOrder, std stdOrder, v uint64) {
	prefix := []byte{0xa5}
	buf := make([]byte, 1, 9)
	buf[0] = 0xa5

	got := so.AppendUint16(buf[:1], uint16(v))
	want := std.AppendUint16(prefix, uint16(v))
	if string(got) != string(want) {
		t.Errorf("%s.AppendUint16(%#x) = %x, want %x", name, uint16(v), got, want)
	}

	got = so.AppendUint32(buf[:1], uint32(v))
	want = std.AppendUint32(prefix, uint32(v))
	if string(got) != string(want) {
		t.Errorf("%s.AppendUint32(%#x) = %x, want %x", name, uint32(v), got, want)
	}

	got = so.AppendUint64(buf[:1], v)
	want = std.AppendUint64(prefix, v)
	if string(got) != string(want) {
		t.Errorf("%s.AppendUint64(%#x) = %x, want %x", name, v, got, want)
	}
}
