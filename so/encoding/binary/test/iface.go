// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binary_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/encoding/binary"
	"solod.dev/so/testing"
)

func TestByteOrder(t *testing.T) {
	var le binary.ByteOrder = &binary.LittleEndian
	var be binary.ByteOrder = &binary.BigEndian

	if le.String() != "LittleEndian" {
		t.Errorf("ByteOrder.String() = %s, want LittleEndian", le.String())
	}
	if be.String() != "BigEndian" {
		t.Errorf("ByteOrder.String() = %s, want BigEndian", be.String())
	}

	var got [8]byte
	var want [8]byte
	for _, value := range testValues {
		le.PutUint16(got[:2], uint16(value))
		putRef(false, want[:2], value, 2)
		if !bytes.Equal(got[:2], want[:2]) {
			t.Errorf("ByteOrder(LittleEndian).PutUint16(%x) = %x, want %x",
				uint16(value), got[:2], want[:2])
		}
		if v := le.Uint16(want[:2]); v != uint16(value) {
			t.Errorf("ByteOrder(LittleEndian).Uint16(%x) = %x, want %x",
				want[:2], v, uint16(value))
		}

		be.PutUint32(got[:4], uint32(value))
		putRef(true, want[:4], value, 4)
		if !bytes.Equal(got[:4], want[:4]) {
			t.Errorf("ByteOrder(BigEndian).PutUint32(%x) = %x, want %x",
				uint32(value), got[:4], want[:4])
		}
		if v := be.Uint32(want[:4]); v != uint32(value) {
			t.Errorf("ByteOrder(BigEndian).Uint32(%x) = %x, want %x",
				want[:4], v, uint32(value))
		}

		be.PutUint64(got[:], value)
		putRef(true, want[:], value, 8)
		if !bytes.Equal(got[:], want[:]) {
			t.Errorf("ByteOrder(BigEndian).PutUint64(%x) = %x, want %x",
				value, got[:], want[:])
		}
		if v := be.Uint64(want[:]); v != value {
			t.Errorf("ByteOrder(BigEndian).Uint64(%x) = %x, want %x",
				want[:], v, value)
		}
	}
}

func TestAppendByteOrder(t *testing.T) {
	var le binary.AppendByteOrder = &binary.LittleEndian
	var be binary.AppendByteOrder = &binary.BigEndian

	if le.String() != "LittleEndian" {
		t.Errorf("AppendByteOrder.String() = %s, want LittleEndian", le.String())
	}
	if be.String() != "BigEndian" {
		t.Errorf("AppendByteOrder.String() = %s, want BigEndian", be.String())
	}

	buf := make([]byte, 0, 8)
	var want [8]byte
	for _, value := range testValues {
		got := le.AppendUint16(buf, uint16(value))
		putRef(false, want[:2], value, 2)
		if !bytes.Equal(got, want[:2]) {
			t.Errorf("AppendByteOrder(LittleEndian).AppendUint16(%x) = %x, want %x",
				uint16(value), got, want[:2])
		}

		got = be.AppendUint32(buf, uint32(value))
		putRef(true, want[:4], value, 4)
		if !bytes.Equal(got, want[:4]) {
			t.Errorf("AppendByteOrder(BigEndian).AppendUint32(%x) = %x, want %x",
				uint32(value), got, want[:4])
		}

		got = be.AppendUint64(buf, value)
		putRef(true, want[:], value, 8)
		if !bytes.Equal(got, want[:]) {
			t.Errorf("AppendByteOrder(BigEndian).AppendUint64(%x) = %x, want %x",
				value, got, want[:])
		}
	}
}
