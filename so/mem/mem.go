// Package mem provides memory allocation facilities.
package mem

import (
	_ "embed"

	"github.com/nalgeon/solod/so/errors"
)

//so:embed mem.h
var mem_h string

var ErrOutOfMemory = errors.New("out of memory")

// Alloc allocates a single value of type T using allocator a.
// Returns a pointer to the allocated memory or panics on failure.
// If the allocator is nil, uses the system allocator.
//
//so:extern
func Alloc[T any](a Allocator) *T { return new(T) }

// TryAlloc allocates memory for a single value of type T using allocator a.
// Returns a pointer to the allocated memory or an error if allocation fails.
// If the allocator is nil, uses the system allocator.
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
// If the allocator is nil, uses the system allocator.
//
//so:extern
func AllocSlice[T any](a Allocator, len int, cap int) []T {
	return make([]T, len, cap)
}

// TryAllocSlice allocates a slice of type T with given length and capacity using allocator a.
// Returns a slice of the allocated memory or an error if allocation fails.
// If the allocator is nil, uses the system allocator.
//
//so:extern
func TryAllocSlice[T any](a Allocator, len int, cap int) ([]T, error) {
	return make([]T, len, cap), nil
}

// FreeSlice frees a slice previously allocated with [AllocSlice] or [TryAllocSlice].
// If the allocator is nil, uses the system allocator.
//
//so:extern
func FreeSlice[T any](a Allocator, slice []T) {}

// Append appends elements to a heap-allocated slice, growing it if needed.
// Returns the updated slice or panics on allocation failure.
// If the allocator is nil, uses the system allocator.
//
//so:extern
func Append[T any](a Allocator, s []T, elems ...T) []T {
	return append(s, elems...)
}

// TryAppend appends elements to a heap-allocated slice, growing it if needed.
// Returns the updated slice or an error if allocation fails.
// If the allocator is nil, uses the system allocator.
//
//so:extern
func TryAppend[T any](a Allocator, s []T, elems ...T) ([]T, error) {
	return append(s, elems...), nil
}

// Extend appends all elements from another slice, growing if needed.
// Returns the updated slice or panics on allocation failure.
// If the allocator is nil, uses the system allocator.
//
//so:extern
func Extend[T any](a Allocator, s []T, other []T) []T {
	return append(s, other...)
}

// TryExtend appends all elements from another slice, growing if needed.
// Returns the updated slice or an error if allocation fails.
// If the allocator is nil, uses the system allocator.
//
//so:extern
func TryExtend[T any](a Allocator, s []T, other []T) ([]T, error) {
	return append(s, other...), nil
}

// MaxAllocaSize is the maximum size that can be allocated with Alloca.
// Defined as the so_MaxAllocaSize constant in the C code.
//
//so:extern
var MaxAllocaSize = 64 << 10 // 64 KiB

// Alloca allocates a block of memory of the given size on the stack.
// The memory is automatically freed when the function that called Alloca returns.
// Panics if the requested size exceeds [MaxAllocaSize].
//
//so:extern
func Alloca(size int) []byte {
	if size > MaxAllocaSize {
		panic("mem: alloca size exceeds allowed")
	}
	return make([]byte, size)
}

//so:extern
var maxAllocSize = 1 << 10 // 1 KiB, for testing purposes

//so:extern
func calloc(count uintptr, size uintptr) any {
	if count*size > uintptr(maxAllocSize) {
		return nil
	}
	return make([]byte, count*size)
}

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
