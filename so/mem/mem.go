// Package mem provides memory allocation facilities.
package mem

import (
	"unsafe"

	"solod.dev/so/errors"
)

//so:embed mem.h
var mem_h string

// ErrOutOfMemory is returned when a memory allocation
// fails due to insufficient memory.
var ErrOutOfMemory = errors.New("out of memory")

// Alloc allocates a single value of type T using allocator a.
// Returns a pointer to the allocated memory or panics on failure.
// Whether new memory is zeroed depends on the allocator.
// If the allocator is nil, uses the system allocator.
//
//so:extern
func Alloc[T any](a Allocator) *T { return new(T) }

// TryAlloc is like [Alloc] but returns an error
// instead of panicking on failure.
//
//so:extern
func TryAlloc[T any](a Allocator) (*T, error) {
	return new(T), nil
}

// Free frees a value previously allocated with [Alloc] or [TryAlloc].
// If the allocator is nil, uses the system allocator.
//
//so:extern
func Free[T any](a Allocator, ptr *T) {}

// AllocSlice allocates a slice of type T with given length
// and capacity using allocator a.
// Returns a slice of the allocated memory or panics on failure.
// Whether new memory is zeroed depends on the allocator.
// If the allocator is nil, uses the system allocator.
//
//so:extern
func AllocSlice[T any](a Allocator, len int, cap int) []T {
	return make([]T, len, cap)
}

// TryAllocSlice is like [AllocSlice] but returns an error
// instead of panicking on allocation failure.
//
//so:extern
func TryAllocSlice[T any](a Allocator, len int, cap int) ([]T, error) {
	return make([]T, len, cap), nil
}

// ReallocSlice reallocates a slice of type T with new length and capacity
// using allocator a. Preserves contents up to the old capacity.
// Returns the reallocated slice or panics on failure.
// Whether new memory is zeroed depends on the allocator.
// If the allocator is nil, uses the system allocator.
//
//so:extern
func ReallocSlice[T any](a Allocator, slice []T, newLen int, newCap int) []T {
	buf := make([]T, newLen, newCap)
	copy(buf, slice)
	return buf
}

// TryReallocSlice is like [ReallocSlice] but returns an error
// instead of panicking on allocation failure.
//
//so:extern
func TryReallocSlice[T any](a Allocator, slice []T, newLen int, newCap int) ([]T, error) {
	buf := make([]T, newLen, newCap)
	copy(buf, slice)
	return buf, nil
}

// FreeSlice frees a slice previously allocated with [AllocSlice] or [TryAllocSlice].
// If the allocator is nil, uses the system allocator.
//
//so:extern
func FreeSlice[T any](a Allocator, slice []T) {}

// FreeString frees a heap-allocated string.
// If the allocator is nil, uses the system allocator.
func FreeString(a Allocator, s string) {
	if len(s) == 0 {
		return
	}
	Free(a, unsafe.StringData(s))
}

// Clear zeroes size bytes starting at ptr + offset.
//
//so:extern
func Clear(ptr any, offset int, size int) {}

// Move copies n bytes from src to dst. Returns dst.
// The memory areas may overlap.
// Panics if either dst or src is nil.
//
//so:extern
func Move(dst any, src any, n int) any { _, _, _ = dst, src, n; return nil }

//so:extern
var maxAllocSize = 1 << 10 // 1 KiB, for testing purposes

// void* calloc(size_t num, size_t size);
//
//so:extern
func calloc(count uintptr, size uintptr) any {
	if count*size > uintptr(maxAllocSize) {
		return nil
	}
	return make([]byte, count*size)
}

// void* malloc(size_t size);
//
//so:extern
func malloc(size uintptr) any {
	if size > uintptr(maxAllocSize) {
		return nil
	}
	return make([]byte, size)
}

// void* realloc(void* ptr, size_t new_size);
//
//so:extern
func realloc(ptr any, newSize uintptr) any {
	_ = ptr
	if newSize > uintptr(maxAllocSize) {
		return nil
	}
	return make([]byte, newSize)
}

//so:extern
func free(ptr any) {}

// void* memmove(void* dest, const void* src, size_t count);
//
//so:extern
func memmove(dst any, src any, count uintptr) any {
	dstSlice := unsafe.Slice(dst.(*byte), int(count))
	srcSlice := unsafe.Slice(src.(*byte), int(count))
	copy(dstSlice, srcSlice)
	return dst
}
