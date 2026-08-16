// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binary_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/encoding/binary"
	"solod.dev/so/testing"
)

func TestUint16(t *testing.T) {
	b := []byte{0x01, 0x23}

	var wantLE uint16 = 0x2301
	if got := binary.LittleEndian.Uint16(b); got != wantLE {
		t.Errorf("LittleEndian.Uint16 = %x, want %x", got, wantLE)
	}

	var wantBE uint16 = 0x0123
	if got := binary.BigEndian.Uint16(b); got != wantBE {
		t.Errorf("BigEndian.Uint16 = %x, want %x", got, wantBE)
	}
}

func TestUint32(t *testing.T) {
	b := []byte{0x01, 0x23, 0x45, 0x67}

	var wantLE uint32 = 0x67452301
	if got := binary.LittleEndian.Uint32(b); got != wantLE {
		t.Errorf("LittleEndian.Uint32 = %x, want %x", got, wantLE)
	}

	var wantBE uint32 = 0x01234567
	if got := binary.BigEndian.Uint32(b); got != wantBE {
		t.Errorf("BigEndian.Uint32 = %x, want %x", got, wantBE)
	}
}

func TestUint64(t *testing.T) {
	b := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}

	var wantLE uint64 = 0xefcdab8967452301
	if got := binary.LittleEndian.Uint64(b); got != wantLE {
		t.Errorf("LittleEndian.Uint64 = %x, want %x", got, wantLE)
	}

	var wantBE uint64 = 0x0123456789abcdef
	if got := binary.BigEndian.Uint64(b); got != wantBE {
		t.Errorf("BigEndian.Uint64 = %x, want %x", got, wantBE)
	}
}

func TestPutUint16(t *testing.T) {
	var v uint16 = 0x0123
	b := make([]byte, 2)

	binary.LittleEndian.PutUint16(b, v)
	if !bytes.Equal(b, []byte{0x23, 0x01}) {
		t.Errorf("LittleEndian.PutUint16 = %x, want 2301", b)
	}

	binary.BigEndian.PutUint16(b, v)
	if !bytes.Equal(b, []byte{0x01, 0x23}) {
		t.Errorf("BigEndian.PutUint16 = %x, want 0123", b)
	}
}

func TestPutUint32(t *testing.T) {
	var v uint32 = 0x01234567
	b := make([]byte, 4)

	binary.LittleEndian.PutUint32(b, v)
	if !bytes.Equal(b, []byte{0x67, 0x45, 0x23, 0x01}) {
		t.Errorf("LittleEndian.PutUint32 = %x, want 67452301", b)
	}

	binary.BigEndian.PutUint32(b, v)
	if !bytes.Equal(b, []byte{0x01, 0x23, 0x45, 0x67}) {
		t.Errorf("BigEndian.PutUint32 = %x, want 01234567", b)
	}
}

func TestPutUint64(t *testing.T) {
	var v uint64 = 0x0123456789abcdef
	b := make([]byte, 8)

	binary.LittleEndian.PutUint64(b, v)
	if !bytes.Equal(b, []byte{0xef, 0xcd, 0xab, 0x89, 0x67, 0x45, 0x23, 0x01}) {
		t.Errorf("LittleEndian.PutUint64 = %x, want efcdab8967452301", b)
	}

	binary.BigEndian.PutUint64(b, v)
	if !bytes.Equal(b, []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}) {
		t.Errorf("BigEndian.PutUint64 = %x, want 0123456789abcdef", b)
	}
}

func TestAppendUint16(t *testing.T) {
	var v uint16 = 0x0123
	b := make([]byte, 0, 4)

	le := binary.LittleEndian.AppendUint16(b, v)
	if !bytes.Equal(le, []byte{0x23, 0x01}) {
		t.Errorf("LittleEndian.AppendUint16 = %x, want 2301", le)
	}

	be := binary.BigEndian.AppendUint16(b, v)
	if !bytes.Equal(be, []byte{0x01, 0x23}) {
		t.Errorf("BigEndian.AppendUint16 = %x, want 0123", be)
	}
}

func TestAppendUint32(t *testing.T) {
	var v uint32 = 0x01234567
	b := make([]byte, 0, 8)

	le := binary.LittleEndian.AppendUint32(b, v)
	if !bytes.Equal(le, []byte{0x67, 0x45, 0x23, 0x01}) {
		t.Errorf("LittleEndian.AppendUint32 = %x, want 67452301", le)
	}

	be := binary.BigEndian.AppendUint32(b, v)
	if !bytes.Equal(be, []byte{0x01, 0x23, 0x45, 0x67}) {
		t.Errorf("BigEndian.AppendUint32 = %x, want 01234567", be)
	}
}

func TestAppendUint64(t *testing.T) {
	var v uint64 = 0x0123456789abcdef
	b := make([]byte, 0, 16)

	le := binary.LittleEndian.AppendUint64(b, v)
	if !bytes.Equal(le, []byte{0xef, 0xcd, 0xab, 0x89, 0x67, 0x45, 0x23, 0x01}) {
		t.Errorf("LittleEndian.AppendUint64 = %x, want efcdab8967452301", le)
	}

	be := binary.BigEndian.AppendUint64(b, v)
	if !bytes.Equal(be, []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}) {
		t.Errorf("BigEndian.AppendUint64 = %x, want 0123456789abcdef", be)
	}
}

func TestAppendPrefix(t *testing.T) {
	const offset = 3
	var v uint64 = 0x0123456789abcdef
	buf := make([]byte, offset, offset+8)
	for i := range offset {
		buf[i] = byte(0xa0 + i)
	}

	var want [8]byte
	for _, be := range orders {
		for _, width := range widths {
			got := appendUint(be, buf, v, width)
			if len(got) != offset+width {
				t.Errorf("%s.AppendUint%d: len = %d, want %d",
					orderName(be), 8*width, len(got), offset+width)
				continue
			}
			if !bytes.Equal(got[:offset], buf) {
				t.Errorf("%s.AppendUint%d: prefix = %x, want %x",
					orderName(be), 8*width, got[:offset], buf)
			}
			putRef(be, want[:width], v, width)
			if !bytes.Equal(got[offset:], want[:width]) {
				t.Errorf("%s.AppendUint%d = %x, want %x",
					orderName(be), 8*width, got[offset:], want[:width])
			}
		}
	}
}

func TestRoundTrip(t *testing.T) {
	// Write a value at an offset and read it back,
	// for every width and both orders.
	const offset = 3
	buf := make([]byte, offset+8)
	for _, be := range orders {
		for _, value := range testValues {
			for _, width := range widths {
				putUint(be, buf[offset:], value, width)
				got := uintAt(be, buf[offset:], width)
				want := lowBytes(value, width)
				if got != want {
					t.Errorf("%s: PutUint%d(%x) reads back as %x",
						orderName(be), 8*width, want, got)
				}
			}
		}
	}
}

func TestLayout(t *testing.T) {
	// Compare Put and Uint against the reference implementations,
	// for every width and both orders.
	var got [8]byte
	var want [8]byte
	for _, be := range orders {
		for _, value := range testValues {
			for _, width := range widths {
				putUint(be, got[:width], value, width)
				putRef(be, want[:width], value, width)
				if !bytes.Equal(got[:width], want[:width]) {
					t.Errorf("%s.PutUint%d(%x) = %x, want %x",
						orderName(be), 8*width, value, got[:width], want[:width])
				}
				if v := uintAt(be, want[:width], width); v != uintRef(be, want[:width], width) {
					t.Errorf("%s.Uint%d(%x) = %x, want %x", orderName(be), 8*width,
						want[:width], v, uintRef(be, want[:width], width))
				}
			}
		}
	}
}

func TestReverse(t *testing.T) {
	// Check that the big endian bytes of a value are the little
	// endian bytes in the reverse order.
	var le [8]byte
	var be [8]byte
	for _, value := range testValues {
		for _, width := range widths {
			putUint(false, le[:width], value, width)
			putUint(true, be[:width], value, width)
			for i := range width {
				if le[i] != be[width-1-i] {
					t.Errorf("PutUint%d(%x): little endian %x, big endian %x",
						8*width, value, le[:width], be[:width])
					break
				}
			}
		}
	}
}

func TestPutBounds(t *testing.T) {
	const fill = 0x5a
	var v uint64 = 0x0123456789abcdef
	buf := make([]byte, 24)
	for _, be := range orders {
		for _, width := range widths {
			for offset := range 8 {
				for i := range len(buf) {
					buf[i] = fill
				}
				putUint(be, buf[offset:offset+width], v, width)

				for i := range len(buf) {
					if i >= offset && i < offset+width {
						continue
					}
					if buf[i] != fill {
						t.Errorf("%s.PutUint%d at %d: byte %d = %x, want 5a",
							orderName(be), 8*width, offset, i, buf[i])
					}
				}
				if got := uintAt(be, buf[offset:], width); got != lowBytes(v, width) {
					t.Errorf("%s.PutUint%d at %d reads back as %x",
						orderName(be), 8*width, offset, got)
				}
			}
		}
	}
}

func TestExtraBytes(t *testing.T) {
	// Check that Put and Uint use the first bytes of a longer slice only.
	var v uint64 = 0x0123456789abcdef
	buf := make([]byte, 16)
	var want [8]byte
	for _, be := range orders {
		for _, width := range widths {
			for i := range len(buf) {
				buf[i] = 0
			}
			putUint(be, buf, v, width)
			putRef(be, want[:width], v, width)
			if !bytes.Equal(buf[:width], want[:width]) {
				t.Errorf("%s.PutUint%d = %x, want %x",
					orderName(be), 8*width, buf[:width], want[:width])
			}

			buf[width] = 0xff
			if got := uintAt(be, buf, width); got != lowBytes(v, width) {
				t.Errorf("%s.Uint%d = %x, want %x",
					orderName(be), 8*width, got, lowBytes(v, width))
			}
		}
	}
}

func TestString(t *testing.T) {
	if got := binary.LittleEndian.String(); got != "LittleEndian" {
		t.Errorf("LittleEndian.String() = %s, want LittleEndian", got)
	}
	if got := binary.BigEndian.String(); got != "BigEndian" {
		t.Errorf("BigEndian.String() = %s, want BigEndian", got)
	}
}
