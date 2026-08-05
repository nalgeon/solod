// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cmp provides types and functions related to comparing ordered values.
// Based on the [cmp] package.
//
// [cmp]: https://github.com/golang/go/blob/go1.26.2/src/cmp/cmp.go
package cmp

import (
	"cmp"
	"unsafe"

	"solod.dev/so/c"
	"solod.dev/so/mem"
)

//so:embed cmp.h
var cmp_h string

// Func is a comparison function that returns a negative value if a < b,
// zero if a == b, and a positive value if a > b.
type Func func(a, b any) int

// FuncFor returns the appropriate comparison function for type T.
// If T is not supported, returns nil.
//
//so:extern
func FuncFor[T any]() Func {
	var zero T
	if _, ok := any(zero).(int); ok {
		return func(a, b any) int {
			i1 := *c.PtrAs[int](a)
			i2 := *c.PtrAs[int](b)
			return cmp.Compare(i1, i2)
		}
	}
	if _, ok := any(zero).(float64); ok {
		return func(a, b any) int {
			f1 := *c.PtrAs[float64](a)
			f2 := *c.PtrAs[float64](b)
			return cmp.Compare(f1, f2)
		}
	}
	if _, ok := any(zero).(string); ok {
		type header struct {
			ptr *byte
			len int
		}
		return func(a, b any) int {
			h1 := c.PtrAs[header](a)
			h2 := c.PtrAs[header](b)
			s1 := unsafe.String(h1.ptr, h1.len)
			s2 := unsafe.String(h2.ptr, h2.len)
			return cmp.Compare(s1, s2)
		}
	}
	return nil
}

// Compare returns
//
//	-1 if x is less than y,
//	 0 if x equals y,
//	+1 if x is greater than y.
//
// For floating-point types, a NaN is considered less than any non-NaN,
// a NaN is considered equal to a NaN, and -0.0 is equal to 0.0.
//
//so:inline
func Compare[T cmp.Ordered](x, y T) int {
	_fn := FuncFor[T]()
	c.Assert(_fn != nil, "cmp: unsupported ordered type")
	return _fn(&x, &y)
}

// Equal reports whether x and y are equal.
// For floating-point types, a NaN is considered equal to a NaN, and -0.0 is equal to 0.0.
// For non-ordered types, compares by raw byte value (memcmp).
//
//so:inline
func Equal[T comparable](x, y T) bool {
	_fn := FuncFor[T]()
	var _eq bool
	if _fn != nil {
		_eq = _fn(&x, &y) == 0
	} else {
		_eq = mem.Compare(&x, &y, c.Sizeof[T]()) == 0
	}
	return _eq
}

// Less reports whether x is less than y.
// For floating-point types, a NaN is considered less than any non-NaN,
// and -0.0 is not less than (is equal to) 0.0.
//
//so:inline
func Less[T cmp.Ordered](x, y T) bool {
	_fn := FuncFor[T]()
	c.Assert(_fn != nil, "cmp: unsupported ordered type")
	return _fn(&x, &y) < 0
}
