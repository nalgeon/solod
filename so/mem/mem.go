package mem

import (
	_ "embed"

	"github.com/nalgeon/solod/so/errors"
)

var ErrOutOfMemory = errors.New("out of memory")
var ErrInvalidSize = errors.New("invalid size")
var System Allocator = SystemAllocator{}

// Allocator defines the interface for memory allocators.
type Allocator interface {
	// Alloc allocates a block of memory of the given size and alignment.
	Alloc(size uintptr, align uintptr) (any, error)
	// Realloc resizes a previously allocated block of memory.
	Realloc(ptr any, oldSize uintptr, newSize uintptr, align uintptr) (any, error)
	// Dealloc frees a previously allocated block of memory.
	Dealloc(ptr any, size uintptr, align uintptr)
}

// SystemAllocator uses the system's malloc, realloc, and free functions.
type SystemAllocator struct{}

func (SystemAllocator) Alloc(size uintptr, align uintptr) (any, error) {
	_ = align
	if size == 0 {
		return nil, ErrInvalidSize
	}
	ptr := calloc(1, size)
	if ptr == nil {
		return nil, ErrOutOfMemory
	}
	return ptr, nil
}

func (a SystemAllocator) Realloc(ptr any, oldSize uintptr, newSize uintptr, align uintptr) (any, error) {
	if newSize == 0 {
		a.Dealloc(ptr, oldSize, align)
		return nil, ErrInvalidSize
	}
	newPtr := realloc(ptr, newSize)
	if newPtr == nil {
		return nil, ErrOutOfMemory
	}
	return newPtr, nil
}

func (SystemAllocator) Dealloc(ptr any, size uintptr, align uintptr) {
	_ = size
	_ = align
	free(ptr)
}

// Alloc allocates memory for a single value of type T using allocator a.
// Returns a pointer to the allocated memory or an error if allocation fails.
//
//so:extern
func Alloc[T any](a Allocator) (*T, error) { return nil, nil }

// Dealloc frees a value previously allocated with Alloc.
//
//so:extern
func Dealloc[T any](a Allocator, ptr *T) {}

// AllocSlice allocates a slice of type T with given length and capacity using allocator a.
// Returns a slice of the allocated memory or an error if allocation fails.
//
//so:extern
func AllocSlice[T any](a Allocator, len int, cap int) ([]T, error) { return nil, nil }

// DeallocSlice frees a slice previously allocated with AllocSlice.
//
//so:extern
func DeallocSlice[T any](a Allocator, slice []T) {}

// New allocates a single value of type T using the system allocator.
// Returns a pointer to the allocated memory or panics on failure.
//
//so:extern
func New[T any]() *T { return nil }

// Free frees a value previously allocated with New.
//
//so:extern
func Free[T any](ptr *T) {}

// NewSlice allocates a slice of type T with given length
// and capacity using the system allocator.
//
//so:extern
func NewSlice[T any](len int, cap int) []T { return nil }

// FreeSlice frees a slice previously allocated with NewSlice.
//
//so:extern
func FreeSlice[T any](slice []T) {}

//so:embed mem.h
var Header string

//so:extern
func malloc(size uintptr) any

//so:extern
func calloc(count uintptr, size uintptr) any

//so:extern
func realloc(ptr any, newSize uintptr) any

//so:extern
func free(ptr any)
