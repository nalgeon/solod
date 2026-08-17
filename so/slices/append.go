package slices

import (
	"solod.dev/so/c"
	"solod.dev/so/mem"
)

// Append appends elements to a heap-allocated slice, growing it if needed.
// If the allocator is nil, uses the system allocator.
// Returns an updated allocated slice; the caller owns it.
//
// Only use Append to append individual elements.
// For appending another slice, use [Extend] instead.
//
//so:extern
func Append[T any](a mem.Allocator, s []T, elems ...T) []T {
	return append(s, elems...)
}

// Extend appends all elements from another heap-allocated slice, growing if needed.
// If the allocator is nil, uses the system allocator.
// Returns an updated allocated slice; the caller owns it.
//
// The slices s and other must not overlap. Extend can move the backing
// array of s. The move invalidates a pointer into the old array.
//
//so:extern
func Extend[T any](a mem.Allocator, s []T, other []T) []T {
	return append(s, other...)
}

// nextcap computes the capacity for a grown slice using Go's growth
// formula: 2x for small slices (< 256 elements), transitioning to ~1.25x
// for larger ones.
//
//so:inline
func slices_nextcap(newLen, oldCap int) int {
	newCap := oldCap
	doubleCap := newCap + newCap
	if newLen > doubleCap {
		return newLen
	}
	const threshold = 256
	if oldCap < threshold {
		return doubleCap
	}
	for {
		newCap += (newCap + 3*threshold) >> 2
		if newCap >= newLen {
			break
		}
	}
	return newCap
}

// grow grows a slice's backing allocation to hold at least newLen elements.
// Returns a result with the updated slice or an error if reallocation fails.
// If the allocator is nil, uses the system allocator.
//
//so:inline
func slices_grow(a mem.Allocator, s Slice, newLen, elemSize, elemAlign int) sliceResult {
	if a == nil {
		a = mem.System
	}
	if newLen <= s.cap {
		return sliceResult{val: s, err: nil}
	}
	newCap := slices_nextcap(newLen, s.cap)
	oldSize := s.cap * elemSize
	newSize := newCap * elemSize
	var newPtr any
	var err error
	if s.cap == 0 {
		newPtr, err = a.Alloc(newSize, elemAlign)
	} else {
		newPtr, err = a.Realloc(s.ptr, oldSize, newSize, elemAlign)
	}
	if err != nil {
		return sliceResult{val: s, err: err}
	}
	s.ptr = newPtr.(*byte)
	s.cap = newCap
	return sliceResult{val: s, err: nil}
}

// tryExtend appends all elements from another slice, growing if needed.
// Returns a result with the updated slice or an error if reallocation fails.
// If the allocator is nil, uses the system allocator.
//
//so:inline
func slices_tryExtend(a mem.Allocator, s Slice, other Slice, elemSize, elemAlign int) sliceResult {
	res := slices_grow(a, s, s.len+other.len, elemSize, elemAlign)
	if res.err != nil {
		return res
	}
	s = res.val
	// An empty slice has a nil data pointer, and mem.Copy rejects one.
	if other.len > 0 {
		mem.Copy(c.PtrAdd(s.ptr, s.len*elemSize), other.ptr, other.len*elemSize)
	}
	s.len += other.len
	return sliceResult{val: s, err: nil}
}

// extend appends all elements from another slice, growing if needed.
// Returns the updated slice or panics on allocation failure.
// If the allocator is nil, uses the system allocator.
//
//so:inline
func slices_extend(a mem.Allocator, s Slice, other Slice, elemSize, elemAlign int) Slice {
	res := slices_tryExtend(a, s, other, elemSize, elemAlign)
	if res.err != nil {
		panic(res.err)
	}
	return res.val
}
