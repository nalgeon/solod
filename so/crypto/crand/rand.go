// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package crand implements a cryptographically secure
// random number generator.
package crand

import (
	"crypto/rand"
	"unsafe"

	"solod.dev/so/io"
	_ "solod.dev/so/runtime"
)

//so:embed rand.h
var rand_h string

// Reader is a global, shared instance of a cryptographically
// secure random number generator. It is safe for concurrent use.
//
//   - On Linux, FreeBSD, and Dragonfly, uses getrandom(2).
//   - On macOS, NetBSD, and OpenBSD, uses arc4random_buf(3).
//   - On Windows, uses BCryptGenRandom.
//   - In a freestanding environment, uses so_crand_read from the target.
var Reader io.Reader = &R{}

type R struct{}

// Read fills b with cryptographically secure random bytes.
func (*R) Read(b []byte) (int, error) {
	return Read(b)
}

// Read fills b with cryptographically secure random bytes.
// It never returns an error, and always fills b entirely.
func Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	read(unsafe.SliceData(b), len(b))
	return len(b), nil
}

const base32alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

// Text returns a cryptographically random string using the standard RFC 4648 base32 alphabet
// for use when a secret string, token, password, or other text is needed.
// The result contains at least 128 bits of randomness, enough to prevent brute force
// guessing attacks and to make the likelihood of collisions vanishingly small.
// A future version may return longer texts as needed to maintain those properties.
//
// Requires a buffer of 26 bytes (⌈log₃₂ 2¹²⁸⌉ = 26 chars).
func Text(b []byte) string {
	if len(b) < 26 {
		panic("crypto/crand.Text: buffer too small")
	}
	Read(b)
	for i := range b {
		b[i] = base32alphabet[b[i]%32]
	}
	return string(b)
}

//so:extern crand_read
func read(b *byte, size int) {
	buf := unsafe.Slice(b, size)
	rand.Read(buf)
}
