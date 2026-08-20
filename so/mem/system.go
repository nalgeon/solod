package mem

import "solod.dev/so/c"

// System is an instance of a memory allocator that uses
// the system's malloc, realloc, and free functions.
//
// System is thread-safe, except on a freestanding target
// with no lock-free atomics of pointer width.
var System Allocator = &SystemAllocator{}

// SystemAllocator uses the system's malloc, realloc, and free functions.
// It zeros out new memory on allocation and reallocation.
//
// A hosted build forwards to the malloc of the host, which is thread-safe.
//
// A freestanding build uses a bump allocator over a static heap, which is
// thread-safe on a target with lock-free atomics of pointer width. Otherwise,
// the allocator is not thread-safe.
type SystemAllocator struct{}

func (*SystemAllocator) Alloc(size int, align int) (any, error) {
	if size <= 0 {
		panic("mem: invalid allocation size")
	}
	if align <= 0 || (align&(align-1)) != 0 {
		panic("mem: invalid alignment")
	}
	ptr := calloc(1, uintptr(size))
	if ptr == nil {
		return nil, ErrOutOfMemory
	}
	return ptr, nil
}

func (*SystemAllocator) Realloc(ptr any, oldSize int, newSize int, align int) (any, error) {
	if oldSize <= 0 || newSize <= 0 {
		panic("mem: invalid allocation size")
	}
	if align <= 0 || (align&(align-1)) != 0 {
		panic("mem: invalid alignment")
	}
	newPtr := realloc(ptr, uintptr(newSize))
	if newPtr == nil {
		return nil, ErrOutOfMemory
	}
	if newSize > oldSize {
		// Zero new memory beyond the old size.
		p := c.PtrAdd(newPtr.(*byte), oldSize)
		Clear(p, newSize-oldSize)
	}
	return newPtr, nil
}

func (*SystemAllocator) Free(ptr any, size int, align int) {
	_ = size
	_ = align
	free(ptr)
}
