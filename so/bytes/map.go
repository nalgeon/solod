// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

import (
	"solod.dev/so/mem"
	"solod.dev/so/slices"
	"solod.dev/so/unicode"
	"solod.dev/so/unicode/utf8"
)

// RuneFunc maps a rune to another rune. If mapping returns
// a negative value, the rune is dropped from the result.
type RuneFunc func(rune) rune

// ToLower returns a copy of the byte slice s with all Unicode letters mapped to
// their lower case.
//
// If the allocator is nil, uses the system allocator.
// The returned slice is allocated; the caller owns it.
func ToLower(a mem.Allocator, s []byte) []byte {
	isASCII, hasUpper := true, false
	for i := range s {
		c := s[i]
		if c >= utf8.RuneSelf {
			isASCII = false
			break
		}
		hasUpper = hasUpper || ('A' <= c && c <= 'Z')
	}

	if isASCII { // optimize for ASCII-only byte slices.
		if !hasUpper {
			return slices.Clone(a, s)
		}
		b := mem.AllocSlice[byte](a, len(s), len(s))
		for i := range s {
			c := s[i]
			if 'A' <= c && c <= 'Z' {
				c += 'a' - 'A'
			}
			b[i] = c
		}
		return b
	}
	return Map(a, unicode.ToLower, s)
}

// ToUpper returns a copy of the byte slice s with all Unicode letters mapped to
// their upper case.
//
// If the allocator is nil, uses the system allocator.
// The returned slice is allocated; the caller owns it.
func ToUpper(a mem.Allocator, s []byte) []byte {
	isASCII, hasLower := true, false
	for i := range s {
		c := s[i]
		if c >= utf8.RuneSelf {
			isASCII = false
			break
		}
		hasLower = hasLower || ('a' <= c && c <= 'z')
	}

	if isASCII { // optimize for ASCII-only byte slices.
		if !hasLower {
			// Just return a copy.
			return slices.Clone(a, s)
		}
		b := mem.AllocSlice[byte](a, len(s), len(s))
		for i := range s {
			c := s[i]
			if 'a' <= c && c <= 'z' {
				c -= 'a' - 'A'
			}
			b[i] = c
		}
		return b
	}
	return Map(a, unicode.ToUpper, s)
}

// Map returns a copy of the byte slice s with all its characters modified
// according to the mapping function. If mapping returns a negative value, the character is
// dropped from the byte slice with no replacement. The characters in s and the
// output are interpreted as UTF-8-encoded code points.
//
// If the allocator is nil, uses the system allocator.
// The returned slice is allocated; the caller owns it.
func Map(a mem.Allocator, mapping RuneFunc, s []byte) []byte {
	b := mem.AllocSlice[byte](a, 0, len(s))
	for i := 0; i < len(s); {
		r, wid := utf8.DecodeRune(s[i:])
		r = mapping(r)
		if r >= 0 {
			var buf [utf8.UTFMax]byte
			n := utf8.EncodeRune(buf[:], r)
			b = slices.Extend(a, b, buf[:n])
		}
		i += wid
	}
	return b
}
