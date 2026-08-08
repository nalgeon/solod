// Package mem provides memory allocation facilities.
package mem

import (
	"unsafe"

	"solod.dev/so/c"
	"solod.dev/so/errors"
)

//so:embed mem.h
var mem_h string

// ErrOutOfMemory is returned when a memory allocation
// fails due to insufficient memory.
var ErrOutOfMemory = errors.New("out of memory")

//so:extern so_max_int
const maxInt = int(uint64(^uint(0)) >> 1)

// Alloc allocates a single value of type T using allocator a.
// Returns a pointer to the allocated memory or panics on failure.
// Whether new memory is zeroed depends on the allocator.
// If the allocator is nil, uses the system allocator.
//
//so:inline
func Alloc[T any](a Allocator) *T {
	_ptr, _err := TryAlloc[T](a)
	if _err != nil {
		panic(_err)
	}
	return _ptr
}

// TryAlloc is like [Alloc] but returns an error
// instead of panicking on failure.
//
//so:inline
func TryAlloc[T any](a Allocator) (*T, error) {
	_a := a
	if _a == nil {
		_a = System
	}
	_ptr, _err := _a.Alloc(c.Sizeof[T](), c.Alignof[T]())
	return c.PtrAs[T](_ptr), _err
}

// Free frees a value previously allocated with [Alloc] or [TryAlloc].
// If the allocator is nil, uses the system allocator.
// Calling Free on a nil pointer is a no-op.
//
//so:inline
func Free[T any](a Allocator, ptr *T) {
	_ptr := ptr
	if _ptr != nil {
		_a := a
		if _a == nil {
			_a = System
		}
		_a.Free(_ptr, c.Sizeof[T](), c.Alignof[T]())
	}
}

// AllocSlice allocates a slice of type T with given length
// and capacity using allocator a.
// Returns a slice of the allocated memory or panics on failure.
// Whether new memory is zeroed depends on the allocator.
// If the allocator is nil, uses the system allocator.
//
//so:inline
func AllocSlice[T any](a Allocator, len int, cap int) []T {
	_s, _err := TryAllocSlice[T](a, len, cap)
	if _err != nil {
		panic(_err)
	}
	return _s
}

// TryAllocSlice is like [AllocSlice] but returns an error
// instead of panicking on allocation failure.
//
//so:inline
func TryAllocSlice[T any](a Allocator, len int, cap int) ([]T, error) {
	_a := a
	if _a == nil {
		_a = System
	}

	_len, _cap := len, cap
	_esize, _align := c.Sizeof[T](), c.Alignof[T]()

	c.Assert(_len >= 0, "mem: negative length")
	c.Assert(_cap >= 0, "mem: negative capacity")
	c.Assert(_len <= _cap, "mem: length exceeds capacity")
	c.Assert(_cap < maxInt/_esize, "mem: capacity overflow")

	var _ptr any
	var _err error
	if _cap > 0 {
		_ptr, _err = _a.Alloc(_esize*_cap, _align)
	}

	var _ts []T
	if _err == nil {
		_ts = c.Slice(c.PtrAs[T](_ptr), _len, _cap)
	}
	return _ts, _err
}

// ReallocSlice reallocates a slice of type T with new length and capacity
// using allocator a. Preserves contents up to the old capacity.
// Returns the reallocated slice or panics on failure.
// Whether new memory is zeroed depends on the allocator.
// If the allocator is nil, uses the system allocator.
//
//so:inline
func ReallocSlice[T any](a Allocator, slice []T, newLen int, newCap int) []T {
	_s, _err := TryReallocSlice(a, slice, newLen, newCap)
	if _err != nil {
		panic(_err)
	}
	return _s
}

// TryReallocSlice is like [ReallocSlice] but returns an error
// instead of panicking on allocation failure.
//
//so:inline
func TryReallocSlice[T any](a Allocator, slice []T, newLen int, newCap int) ([]T, error) {
	_a := a
	if _a == nil {
		_a = System
	}

	_oldCap := cap(slice)
	_newLen, _newCap := newLen, newCap
	_esize, _align := c.Sizeof[T](), c.Alignof[T]()

	c.Assert(_newLen >= 0, "mem: negative length")
	c.Assert(_newCap >= 0, "mem: negative capacity")
	c.Assert(_newLen <= _newCap, "mem: length exceeds capacity")
	c.Assert(_newCap < maxInt/_esize, "mem: capacity overflow")

	var _newPtr any
	var _err error
	if _newCap == 0 {
		if _oldCap > 0 {
			_a.Free(unsafe.SliceData(slice), _esize*_oldCap, _align)
		}
	} else if _oldCap == 0 {
		_newPtr, _err = _a.Alloc(_esize*_newCap, _align)
	} else {
		ptr := unsafe.SliceData(slice)
		_newPtr, _err = _a.Realloc(ptr, _esize*_oldCap, _esize*_newCap, _align)
	}

	var _s []T
	if _err == nil {
		_s = c.Slice(c.PtrAs[T](_newPtr), _newLen, _newCap)
	}
	return _s, _err
}

// FreeSlice frees a slice previously allocated with [AllocSlice] or [TryAllocSlice].
// If the allocator is nil, uses the system allocator.
// Calling FreeSlice on an empty or nil slice is a no-op.
//
//so:inline
func FreeSlice[T any](a Allocator, slice []T) {
	_a := a
	if _a == nil {
		_a = System
	}

	_s := slice
	_cap := cap(_s)
	if _cap > 0 {
		_a.Free(unsafe.SliceData(_s), c.Sizeof[T]()*_cap, c.Alignof[T]())
	}
}

// FreeString frees a heap-allocated string.
// If the allocator is nil, uses the system allocator.
// Calling FreeString on an empty string is a no-op.
func FreeString(a Allocator, s string) {
	if len(s) == 0 {
		return
	}

	_a := a
	if _a == nil {
		_a = System
	}
	_a.Free(unsafe.StringData(s), len(s), c.Alignof[byte]())
}

// Clear zeroes size bytes starting at ptr.
//
//so:inline
func Clear(ptr any, size int) {
	c.Assert(ptr != nil, "mem: nil pointer")
	c.Assert(size >= 0, "mem: negative size")
	memset(ptr, 0, uintptr(size))
}

// Compare compares size bytes at a and b.
// Returns an integer comparing the bytes at a and b.
// The result will be 0 if the bytes are equal, -1 if a < b, and +1 if a > b.
// Panics if either a or b is nil.
//
//so:inline
func Compare(a any, b any, size int) int {
	c.Assert(a != nil, "mem: nil pointer")
	c.Assert(b != nil, "mem: nil pointer")
	c.Assert(size >= 0, "mem: negative size")
	res := memcmp(a, b, uintptr(size))
	if res < 0 {
		return -1
	} else if res > 0 {
		return 1
	}
	return 0
}

// Copy copies n bytes from src to dst. Returns dst.
// The memory areas must not overlap.
// Panics if either dst or src is nil.
//
//so:inline
func Copy(dst any, src any, n int) any {
	c.Assert(dst != nil, "mem: nil pointer")
	c.Assert(src != nil, "mem: nil pointer")
	c.Assert(n >= 0, "mem: negative size")
	return memcpy(dst, src, uintptr(n))
}

// Move copies n bytes from src to dst. Returns dst.
// The memory areas may overlap.
// Panics if either dst or src is nil.
//
//so:inline
func Move(dst any, src any, n int) any {
	c.Assert(dst != nil, "mem: nil pointer")
	c.Assert(src != nil, "mem: nil pointer")
	c.Assert(n >= 0, "mem: negative size")
	return memmove(dst, src, uintptr(n))
}

// Swap swaps the values pointed to by a and b.
// Panics if either a or b is nil.
//
//so:inline
func Swap[T any](a *T, b *T) {
	c.Assert(a != nil, "mem: nil pointer")
	c.Assert(b != nil, "mem: nil pointer")
	_tmp := *a
	*a = *b
	*b = _tmp
}

// SwapByte swaps n bytes between a and b.
// Panics if either a or b is nil.
//
// SwapByte temporarily allocates a buffer of size n
// on the stack, so it's not suitable for large n.
//
//so:extern
func SwapByte(a any, b any, n int) {
	// Has to be implemented as extern because it uses VLA.
	pa := unsafe.Slice((*byte)(ptrVal(a)), n)
	pb := unsafe.Slice((*byte)(ptrVal(b)), n)
	tmp := make([]byte, n)
	copy(tmp, pb)
	copy(pb, pa)
	copy(pa, tmp)
}

// void* memset(void *dest, int ch, size_t count);
//
//so:extern
func memset(ptr any, ch c.Int, count uintptr) any {
	s := unsafe.Slice((*byte)(ptrVal(ptr)), int(count))
	for i := range s {
		s[i] = byte(ch)
	}
	return ptr
}

// int memcmp(const void *s1, const void *s2, size_t n);
//
//so:extern
func memcmp(s1 any, s2 any, n uintptr) c.Int {
	slice1 := unsafe.Slice((*byte)(ptrVal(s1)), int(n))
	slice2 := unsafe.Slice((*byte)(ptrVal(s2)), int(n))
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return c.Int(slice1[i]) - c.Int(slice2[i])
		}
	}
	return 0
}

// void* memcpy(void* dest, const void* src, size_t count);
//
//so:extern
func memcpy(dst any, src any, count uintptr) any {
	return memmove(dst, src, count)
}

// void* memmove(void* dest, const void* src, size_t count);
//
//so:extern
func memmove(dst any, src any, count uintptr) any {
	dstSlice := unsafe.Slice((*byte)(ptrVal(dst)), int(count))
	srcSlice := unsafe.Slice((*byte)(ptrVal(src)), int(count))
	copy(dstSlice, srcSlice)
	return dst
}
