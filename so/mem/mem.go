package mem

import (
	_ "embed"

	"github.com/nalgeon/solod/so/errors"
)

var ErrOutOfMemory = errors.New("out of memory")

// Alloc allocates memory for a single value of type T using allocator a.
// Returns a pointer to the allocated memory or an error if allocation fails.
//
//so:extern
func Alloc[T any](a Allocator) (*T, error) {
	return new(T), nil
}

// Dealloc frees a value previously allocated with Alloc.
//
//so:extern
func Dealloc[T any](a Allocator, ptr *T) {}

// AllocSlice allocates a slice of type T with given length and capacity using allocator a.
// Returns a slice of the allocated memory or an error if allocation fails.
//
//so:extern
func AllocSlice[T any](a Allocator, len int, cap int) ([]T, error) {
	return make([]T, len, cap), nil
}

// DeallocSlice frees a slice previously allocated with AllocSlice.
//
//so:extern
func DeallocSlice[T any](a Allocator, slice []T) {}

// New allocates a single value of type T using the system allocator.
// Returns a pointer to the allocated memory or panics on failure.
//
//so:extern
func New[T any]() *T { return new(T) }

// Free frees a value previously allocated with New.
//
//so:extern
func Free[T any](ptr *T) {}

// NewSlice allocates a slice of type T with given length
// and capacity using the system allocator.
//
//so:extern
func NewSlice[T any](len int, cap int) []T {
	return make([]T, len, cap)
}

// FreeSlice frees a slice previously allocated with NewSlice.
//
//so:extern
func FreeSlice[T any](slice []T) {}

//so:embed mem.h
var Header string

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
