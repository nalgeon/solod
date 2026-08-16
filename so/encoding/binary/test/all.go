// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binary_test

import "solod.dev/so/encoding/binary"

// widths are the value widths in bytes the package converts.
var widths = [3]int{2, 4, 8}

// orders are the byte orders under test. True is big endian.
var orders = [2]bool{false, true}

// testValues are the values the table tests convert.
var testValues = [11]uint64{
	0x0000000000000000,
	0x0000000000000001,
	0x00000000000000ff,
	0x000000000000ff00,
	0x0123456789abcdef,
	0xfedcba9876543210,
	0xffffffffffffffff,
	0xaaaaaaaaaaaaaaaa,
	0x5555555555555555,
	0x400921fb54442d18, // math.Float64bits(math.Pi)
	0x4005bf0a8b145769, // math.Float64bits(math.E)
}

// orderName returns the name of the byte order.
func orderName(be bool) string {
	if be {
		return "BigEndian"
	}
	return "LittleEndian"
}

// lowBytes returns the low width bytes of v.
func lowBytes(v uint64, width int) uint64 {
	if width == 8 {
		return v
	}
	return v & (uint64(1)<<uint(8*width) - 1)
}

// putUint writes the low width bytes of v into b.
func putUint(be bool, b []byte, v uint64, width int) {
	if be {
		putUintBE(b, v, width)
		return
	}
	putUintLE(b, v, width)
}

func putUintLE(b []byte, v uint64, width int) {
	switch width {
	case 2:
		binary.LittleEndian.PutUint16(b, uint16(v))
	case 4:
		binary.LittleEndian.PutUint32(b, uint32(v))
	case 8:
		binary.LittleEndian.PutUint64(b, v)
	}
}

func putUintBE(b []byte, v uint64, width int) {
	switch width {
	case 2:
		binary.BigEndian.PutUint16(b, uint16(v))
	case 4:
		binary.BigEndian.PutUint32(b, uint32(v))
	case 8:
		binary.BigEndian.PutUint64(b, v)
	}
}

// uintAt returns the value of the first width bytes of b.
func uintAt(be bool, b []byte, width int) uint64 {
	if be {
		return uintAtBE(b, width)
	}
	return uintAtLE(b, width)
}

func uintAtLE(b []byte, width int) uint64 {
	switch width {
	case 2:
		return uint64(binary.LittleEndian.Uint16(b))
	case 4:
		return uint64(binary.LittleEndian.Uint32(b))
	}
	return binary.LittleEndian.Uint64(b)
}

func uintAtBE(b []byte, width int) uint64 {
	switch width {
	case 2:
		return uint64(binary.BigEndian.Uint16(b))
	case 4:
		return uint64(binary.BigEndian.Uint32(b))
	}
	return binary.BigEndian.Uint64(b)
}

// appendUint appends the low width bytes of v to b.
func appendUint(be bool, b []byte, v uint64, width int) []byte {
	if be {
		return appendUintBE(b, v, width)
	}
	return appendUintLE(b, v, width)
}

func appendUintLE(b []byte, v uint64, width int) []byte {
	switch width {
	case 2:
		return binary.LittleEndian.AppendUint16(b, uint16(v))
	case 4:
		return binary.LittleEndian.AppendUint32(b, uint32(v))
	}
	return binary.LittleEndian.AppendUint64(b, v)
}

func appendUintBE(b []byte, v uint64, width int) []byte {
	switch width {
	case 2:
		return binary.BigEndian.AppendUint16(b, uint16(v))
	case 4:
		return binary.BigEndian.AppendUint32(b, uint32(v))
	}
	return binary.BigEndian.AppendUint64(b, v)
}

// putRef writes the low width bytes of v into b, one byte at a time.
// It is the reference implementation of PutUint16, PutUint32 and PutUint64.
func putRef(be bool, b []byte, v uint64, width int) {
	for i := range width {
		shift := uint(8 * i)
		if be {
			shift = uint(8 * (width - 1 - i))
		}
		b[i] = byte(v >> shift)
	}
}

// uintRef returns the value of the first width bytes of b, one byte at a time.
// It is the reference implementation of Uint16, Uint32 and Uint64.
func uintRef(be bool, b []byte, width int) uint64 {
	var v uint64
	for i := range width {
		if be {
			v = v<<8 | uint64(b[i])
			continue
		}
		v |= uint64(b[i]) << uint(8*i)
	}
	return v
}
