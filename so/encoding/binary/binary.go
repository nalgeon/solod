// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package binary translates between numbers and byte sequences.
//
// [LittleEndian] and [BigEndian] convert a 16-, 32- or 64-bit unsigned integer
// to a byte sequence and back.
//
// Based on the [encoding/binary] package.
//
// [encoding/binary]: https://github.com/golang/go/blob/go1.26.2/src/encoding/binary
package binary

// A ByteOrder specifies how to convert byte slices into
// 16-, 32-, or 64-bit unsigned integers.
//
// It is implemented by [LittleEndian] and [BigEndian].
type ByteOrder interface {
	Uint16([]byte) uint16
	Uint32([]byte) uint32
	Uint64([]byte) uint64
	PutUint16([]byte, uint16)
	PutUint32([]byte, uint32)
	PutUint64([]byte, uint64)
	String() string
}

// AppendByteOrder specifies how to append 16-, 32-, or 64-bit unsigned integers
// into a byte slice.
//
// It is implemented by [LittleEndian] and [BigEndian].
type AppendByteOrder interface {
	AppendUint16([]byte, uint16) []byte
	AppendUint32([]byte, uint32) []byte
	AppendUint64([]byte, uint64) []byte
	String() string
}

// LittleEndian is the little-endian implementation of [ByteOrder] and [AppendByteOrder].
var LittleEndian LE

// BigEndian is the big-endian implementation of [ByteOrder] and [AppendByteOrder].
var BigEndian BE

type LE struct {
	empty byte
}

// Uint16 returns the uint16 representation of b[0:2].
func (*LE) Uint16(b []byte) uint16 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	return uint16(b[0]) | uint16(b[1])<<8
}

// PutUint16 stores v into b[0:2].
func (*LE) PutUint16(b []byte, v uint16) {
	_ = b[1] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
}

// AppendUint16 appends the bytes of v to b and returns the appended slice.
// Requires at least 2 bytes of spare capacity in b.
func (*LE) AppendUint16(b []byte, v uint16) []byte {
	return append(b,
		byte(v),
		byte(v>>8),
	)
}

// Uint32 returns the uint32 representation of b[0:4].
func (*LE) Uint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

// PutUint32 stores v into b[0:4].
func (*LE) PutUint32(b []byte, v uint32) {
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

// AppendUint32 appends the bytes of v to b and returns the appended slice.
// Requires at least 4 bytes of spare capacity in b.
func (*LE) AppendUint32(b []byte, v uint32) []byte {
	return append(b,
		byte(v),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24),
	)
}

// Uint64 returns the uint64 representation of b[0:8].
func (*LE) Uint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}

// PutUint64 stores v into b[0:8].
func (*LE) PutUint64(b []byte, v uint64) {
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
}

// AppendUint64 appends the bytes of v to b and returns the appended slice.
// Requires at least 8 bytes of spare capacity in b.
func (*LE) AppendUint64(b []byte, v uint64) []byte {
	return append(b,
		byte(v),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24),
		byte(v>>32),
		byte(v>>40),
		byte(v>>48),
		byte(v>>56),
	)
}

func (*LE) String() string { return "LittleEndian" }

type BE struct {
	empty byte
}

// Uint16 returns the uint16 representation of b[0:2].
func (*BE) Uint16(b []byte) uint16 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	return uint16(b[1]) | uint16(b[0])<<8
}

// PutUint16 stores v into b[0:2].
func (*BE) PutUint16(b []byte, v uint16) {
	_ = b[1] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 8)
	b[1] = byte(v)
}

// AppendUint16 appends the bytes of v to b and returns the appended slice.
// Requires at least 2 bytes of spare capacity in b.
func (*BE) AppendUint16(b []byte, v uint16) []byte {
	return append(b,
		byte(v>>8),
		byte(v),
	)
}

// Uint32 returns the uint32 representation of b[0:4].
func (*BE) Uint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
}

// PutUint32 stores v into b[0:4].
func (*BE) PutUint32(b []byte, v uint32) {
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}

// AppendUint32 appends the bytes of v to b and returns the appended slice.
// Requires at least 4 bytes of spare capacity in b.
func (*BE) AppendUint32(b []byte, v uint32) []byte {
	return append(b,
		byte(v>>24),
		byte(v>>16),
		byte(v>>8),
		byte(v),
	)
}

// Uint64 returns the uint64 representation of b[0:8].
func (*BE) Uint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
}

// PutUint64 stores v into b[0:8].
func (*BE) PutUint64(b []byte, v uint64) {
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
}

// AppendUint64 appends the bytes of v to b and returns the appended slice.
// Requires at least 8 bytes of spare capacity in b.
func (*BE) AppendUint64(b []byte, v uint64) []byte {
	return append(b,
		byte(v>>56),
		byte(v>>48),
		byte(v>>40),
		byte(v>>32),
		byte(v>>24),
		byte(v>>16),
		byte(v>>8),
		byte(v),
	)
}

func (*BE) String() string { return "BigEndian" }
