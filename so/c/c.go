// Package c provides convenience helpers for C interop.
// It bridges C's null-terminated strings and raw pointers
// with So's string and slice types.
package c

import _ "embed"

//so:embed c.h
var c_h string

// Bytes wraps a raw byte pointer and length into a []byte without copying.
// If ptr is nil, returns nil.
//
//so:extern
func Bytes(ptr *byte, n int) []byte { _, _ = ptr, n; return nil }

// String converts a null-terminated C string to a So string.
// If ptr is nil, returns "".
//
//so:extern
func String(ptr *byte) string { _ = ptr; return "" }

// CharPtr casts a *byte (uint8_t*) to char* for C functions
// that expect char* instead of uint8_t*.
//
//so:extern
func CharPtr(ptr *byte) *byte { _ = ptr; return nil }
