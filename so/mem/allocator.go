package mem

var System Allocator = SystemAllocator{}

// Allocator defines the interface for memory allocators.
type Allocator interface {
	// Alloc allocates a block of memory of the given size and alignment.
	Alloc(size int, align int) (any, error)
	// Realloc resizes a previously allocated block of memory.
	Realloc(ptr any, oldSize int, newSize int, align int) (any, error)
	// Dealloc frees a previously allocated block of memory.
	Dealloc(ptr any, size int, align int)
}

// SystemAllocator uses the system's malloc, realloc, and free functions.
type SystemAllocator struct{}

func (SystemAllocator) Alloc(size int, align int) (any, error) {
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

func (SystemAllocator) Realloc(ptr any, oldSize int, newSize int, align int) (any, error) {
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
	return newPtr, nil
}

func (SystemAllocator) Dealloc(ptr any, size int, align int) {
	_ = size
	_ = align
	free(ptr)
}
