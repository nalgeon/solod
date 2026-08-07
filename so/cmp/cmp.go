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
	"reflect"

	"solod.dev/so/c"
	"solod.dev/so/mem"
)

//so:embed cmp.h
var cmp_h string

// Func is a comparison function that returns a negative value if a < b,
// zero if a == b, and a positive value if a > b.
type Func func(a, b any) int

// FuncFor returns the comparison function for type T.
// These types are supported:
//
//	int, int8, int16, int32 (rune), int64
//	uint, uint8 (byte), uint16, uint32, uint64
//	float32, float64
//	string
//
// A named type is supported if its underlying type is in the list.
// For any other type, FuncFor returns nil. Notably, uintptr, bool,
// pointers, structs and arrays are not supported.
//
// Assign the result to a variable before a comparison with nil:
// GCC rejects a direct comparison of a function address to nil.
//
//so:extern
func FuncFor[T any]() Func {
	switch reflect.TypeFor[T]().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(a, b any) int {
			i1 := reflect.ValueOf(*c.PtrAs[T](a)).Int()
			i2 := reflect.ValueOf(*c.PtrAs[T](b)).Int()
			return cmp.Compare(i1, i2)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return func(a, b any) int {
			u1 := reflect.ValueOf(*c.PtrAs[T](a)).Uint()
			u2 := reflect.ValueOf(*c.PtrAs[T](b)).Uint()
			return cmp.Compare(u1, u2)
		}
	case reflect.Float32, reflect.Float64:
		return func(a, b any) int {
			f1 := reflect.ValueOf(*c.PtrAs[T](a)).Float()
			f2 := reflect.ValueOf(*c.PtrAs[T](b)).Float()
			return cmp.Compare(f1, f2)
		}
	case reflect.String:
		return func(a, b any) int {
			s1 := reflect.ValueOf(*c.PtrAs[T](a)).String()
			s2 := reflect.ValueOf(*c.PtrAs[T](b)).String()
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
// Panics for a type that [FuncFor] does not support.
// The constraint accepts uintptr, but the implementation does not.
//
//so:inline
func Compare[T cmp.Ordered](x, y T) int {
	_x, _y := x, y
	_fn := FuncFor[T]()
	c.Assert(_fn != nil, "cmp: unsupported ordered type")
	return _fn(&_x, &_y)
}

// Equal reports whether x and y are equal.
// For floating-point types, a NaN is considered equal to a NaN, and -0.0 is equal to 0.0.
// For a type that [FuncFor] does not support, compares by raw byte value (memcmp).
//
// The constraint accepts an array, but the implementation does not:
// Equal copies both arguments into local variables, and C cannot copy an array.
//
//so:inline
func Equal[T comparable](x, y T) bool {
	_x, _y := x, y
	_fn := FuncFor[T]()
	var _eq bool
	if _fn != nil {
		_eq = _fn(&_x, &_y) == 0
	} else {
		_eq = mem.Compare(&_x, &_y, c.Sizeof[T]()) == 0
	}
	return _eq
}

// Less reports whether x is less than y.
// For floating-point types, a NaN is considered less than any non-NaN,
// and -0.0 is not less than (is equal to) 0.0.
//
// Panics for a type that [FuncFor] does not support.
// The constraint accepts uintptr, but FuncFor does not support uintptr.
//
//so:inline
func Less[T cmp.Ordered](x, y T) bool {
	_x, _y := x, y
	_fn := FuncFor[T]()
	c.Assert(_fn != nil, "cmp: unsupported ordered type")
	return _fn(&_x, &_y) < 0
}
