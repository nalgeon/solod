package mem

import "solod.dev/so/errors"

var ErrNoAlloc = errors.New("mem: allocation not allowed")

// NoAlloc is an instance of an Allocator that returns an error
// on any allocation. Free is a no-op.
// It's meant for cases when allocations are strictly forbidden.
var NoAlloc Allocator = &NoAllocator{}

// NoAllocator is an Allocator that returns an error
// on any allocation. Free is a no-op.
// It's meant for cases when allocations are strictly forbidden.
type NoAllocator struct{}

func (*NoAllocator) Alloc(size int, align int) (any, error) {
	_, _ = size, align
	return nil, ErrNoAlloc
}

func (*NoAllocator) Realloc(ptr any, oldSize int, newSize int, align int) (any, error) {
	_, _, _, _ = ptr, oldSize, newSize, align
	return nil, ErrNoAlloc
}

func (*NoAllocator) Free(ptr any, size int, align int) {
	_, _, _ = ptr, size, align
}
