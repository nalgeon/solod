// Package cstring wraps the C <string.h> header.
// It offers memory block operations.
package cstring

import _ "embed"

//so:embed cstring.h
var cstring_h string

// Memcpy copies n bytes from src to dst. Returns dst.
// The memory areas must not overlap; use [Memmove] for overlapping regions.
//
// If either dst or src is nil, the behavior is undefined.
//
//so:extern
func Memcpy(dst any, src any, n uintptr) any { _, _, _ = dst, src, n; return nil }

// Memmove copies n bytes from src to dst. Returns dst.
// Unlike [Memcpy], the memory areas may overlap.
//
// If either dst or src is nil, the behavior is undefined.
//
//so:extern
func Memmove(dst any, src any, n uintptr) any { _, _, _ = dst, src, n; return nil }

// Memset fills the first n bytes of the memory area pointed to by ptr
// with the byte value (interpreted as unsigned char). Returns ptr.
//
// If ptr is nil, the behavior is undefined.
//
//so:extern
func Memset(ptr any, value int, n uintptr) any { _, _, _ = ptr, value, n; return nil }

// Memcmp compares the first n bytes of memory areas a and b.
// Returns a negative value if a < b, zero if a == b,
// or a positive value if a > b.
//
// If either a or b is nil, the behavior is undefined.
//
//so:extern
func Memcmp(a any, b any, n uintptr) int { _, _, _ = a, b, n; return 0 }
