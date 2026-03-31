#include "so/builtin/builtin.h"

// Forwards declarations.
struct mem_Allocator;
so_R_slice_err mem_tryAllocSlice(
    const struct mem_Allocator* a,
    size_t elemSize, size_t align,
    so_int len, so_int cap);
so_R_slice_err mem_tryReallocSlice(
    const struct mem_Allocator* a, so_Slice s,
    so_int newLen, so_int newCap,
    size_t elemSize, size_t align);

// Alloc allocates a single value of type T using allocator a.
// Returns a pointer to the allocated memory or panics on failure.
// Whether new memory is zeroed depends on the allocator.
// If the allocator is nil, uses the system allocator.
#define mem_Alloc(T, a) ({                        \
    so_R_ptr_err _mem_res = mem_TryAlloc(T, (a)); \
    if (_mem_res.err != NULL)                     \
        so_panic(_mem_res.err->msg);              \
    _mem_res.val;                                 \
})

// TryAlloc is like [Alloc] but returns an error
// instead of panicking on failure.
#define mem_TryAlloc(T, a) ({                            \
    mem_Allocator _a = (a);                              \
    if (!_a.self) _a = mem_System;                       \
    _a.Alloc(_a.self, sizeof(T), alignof(so_typeof(T))); \
})

// Free frees a value previously allocated with [Alloc] or [TryAlloc].
// If the allocator is nil, uses the system allocator.
#define mem_Free(T, a, ptr) ({                                 \
    mem_Allocator _a = (a);                                    \
    if (!_a.self) _a = mem_System;                             \
    _a.Free(_a.self, (ptr), sizeof(T), alignof(so_typeof(T))); \
})

// AllocSlice allocates a slice of type T with given length
// and capacity using allocator a.
// Returns a slice of the allocated memory or panics on failure.
// Whether new memory is zeroed depends on the allocator.
// If the allocator is nil, uses the system allocator.
#define mem_AllocSlice(T, a, len, cap) ({                          \
    so_R_slice_err _res = mem_TryAllocSlice(T, (a), (len), (cap)); \
    if (_res.err != NULL)                                          \
        so_panic(_res.err->msg);                                   \
    _res.val;                                                      \
})

// TryAllocSlice is like [AllocSlice] but returns an error
// instead of panicking on allocation failure.
#define mem_TryAllocSlice(T, a, slen, scap) ({                \
    mem_Allocator _a = (a);                                   \
    mem_tryAllocSlice(&_a, sizeof(T),                         \
                      alignof(so_typeof(T)), (slen), (scap)); \
})

// ReallocSlice reallocates a slice of type T with new length and capacity
// using allocator a. Preserves contents up to the old capacity.
// Returns the reallocated slice or panics on failure.
// Whether new memory is zeroed depends on the allocator.
// If the allocator is nil, uses the system allocator.
#define mem_ReallocSlice(T, a, s, newLen, newCap) ({                            \
    so_R_slice_err _res = mem_TryReallocSlice(T, (a), (s), (newLen), (newCap)); \
    if (_res.err != NULL)                                                       \
        so_panic(_res.err->msg);                                                \
    _res.val;                                                                   \
})

// TryReallocSlice is like [ReallocSlice] but returns an error
// instead of panicking on allocation failure.
#define mem_TryReallocSlice(T, a, s, newLen, newCap) ({    \
    mem_Allocator _a = (a);                                \
    mem_tryReallocSlice(&_a, (s), (newLen), (newCap),      \
                        sizeof(T), alignof(so_typeof(T))); \
})

// FreeSlice frees a slice previously allocated with [AllocSlice] or [TryAllocSlice].
// If the allocator is nil, uses the system allocator.
#define mem_FreeSlice(T, a, s) ({                                        \
    mem_Allocator _a = (a);                                              \
    so_Slice _s = (s);                                                   \
    if (!_a.self) _a = mem_System;                                       \
    _a.Free(_a.self, _s.ptr, sizeof(T) * _s.cap, alignof(so_typeof(T))); \
})

// Clear zeroes size bytes starting at ptr + offset.
static inline void mem_Clear(void* ptr, so_int offset, so_int size) {
    if (ptr == NULL) so_panic("mem: nil pointer");
    if (offset < 0) so_panic("mem: negative offset");
    if (size < 0) so_panic("mem: negative size");
    memset((char*)ptr + offset, 0, (size_t)size);
}

// Move copies n bytes from src to dst. Returns dst.
// The memory areas may overlap.
// Panics if either dst or src is nil.
static inline void* mem_Move(void* dst, const void* src, so_int n) {
    if (dst == NULL || src == NULL) so_panic("mem: nil pointer");
    if (n < 0) so_panic("mem: negative size");
    return memmove(dst, src, (size_t)n);
}
