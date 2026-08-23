// Package slices provides various functions useful with slices of any type.
// Based on the [slices] package.
//
// [slices]: https://github.com/golang/go/blob/go1.26.2/src/slices/slices.go
package slices

import (
	"unsafe"

	"solod.dev/so/c"
	"solod.dev/so/cmp"
	"solod.dev/so/mem"
)

//so:embed slices.h
var slices_h string

// A Slice is a header for a slice of any type.
//
//so:extern so_Slice
type Slice struct {
	ptr *byte
	len int
	cap int
}

// Header returns the Slice header for a given slice.
//
//so:extern
func Header[T any](s []T) Slice {
	return Slice{
		ptr: c.SliceData[byte](s),
		len: len(s),
		cap: cap(s),
	}
}

//so:extern so_R_slice_err
type sliceResult struct {
	val Slice
	err error
}

// Make allocates a slice of type T with given length using allocator a.
// If the allocator is nil, uses the system allocator.
// The returned slice is allocated; the caller owns it.
//
//so:inline
func Make[T any](a mem.Allocator, len int) []T {
	return mem.AllocSlice[T](a, len, len)
}

// MakeCap allocates a slice of type T with given length and capacity using allocator a.
// If the allocator is nil, uses the system allocator.
// The returned slice is allocated; the caller owns it.
//
//so:inline
func MakeCap[T any](a mem.Allocator, len int, cap int) []T {
	return mem.AllocSlice[T](a, len, cap)
}

// Free frees a previously allocated slice.
// If the allocator is nil, uses the system allocator.
//
//so:inline
func Free[T any](a mem.Allocator, s []T) {
	mem.FreeSlice(a, s)
}

// Clone returns a shallow copy of the slice.
// If the allocator is nil, uses the system allocator.
// The returned slice is allocated; the caller owns it.
//
// Clone skips the copy for an empty or nil slice.
//
//so:inline
func Clone[T any](a mem.Allocator, s []T) []T {
	_s, _slen := s, len(s)
	_elemSize := c.Sizeof[T]()
	_newSlice := mem.AllocSlice[T](a, _slen, _slen)
	if _slen > 0 {
		mem.Copy(unsafe.SliceData(_newSlice), unsafe.SliceData(_s), _slen*_elemSize)
	}
	return _newSlice
}

// Equal reports whether two slices are equal: the same length and all
// elements equal. Empty and nil slices are considered equal.
//
//so:inline
func Equal[T comparable](s1, s2 []T) bool {
	_s1, _s2 := s1, s2
	_eq := len(_s1) == len(_s2)
	for i := 0; i < len(_s1) && _eq; i++ {
		_v1, _v2 := _s1[i], _s2[i]
		_eq = cmp.Equal(_v1, _v2)
	}
	return _eq
}

// Contains reports whether v is present in s.
//
//so:inline
func Contains[T comparable](s []T, v T) bool {
	return Index(s, v) >= 0
}

// Index returns the index of the first occurrence of v in s,
// or -1 if not present.
//
//so:inline
func Index[T comparable](s []T, v T) int {
	_s, _v := s, v
	_idx := -1
	for _j := range _s {
		_sj := _s[_j]
		if cmp.Equal(_sj, _v) {
			_idx = _j
			break
		}
	}
	return _idx
}
