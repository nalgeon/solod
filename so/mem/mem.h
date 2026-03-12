#include <stddef.h>

// Alloc allocates a single value of type T using allocator a.
// Returns a pointer to the allocated memory or panics on failure.
// If the allocator is nil, uses the system allocator.
#define mem_Alloc(T, a) ({                     \
    so_Result _mem_res = mem_TryAlloc(T, (a)); \
    if (_mem_res.err != NULL)                  \
        so_panic(_mem_res.err->msg);           \
    _mem_res.val.as_ptr;                       \
})

// TryAlloc allocates memory for a single value of type T using allocator a.
// Returns a pointer to the allocated memory or an error if allocation fails.
// If the allocator is nil, uses the system allocator.
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
// If the allocator is nil, uses the system allocator.
#define mem_AllocSlice(T, a, len, cap) ({                     \
    so_Result _res = mem_TryAllocSlice(T, (a), (len), (cap)); \
    if (_res.err != NULL)                                     \
        so_panic(_res.err->msg);                              \
    _res.val.as_slice;                                        \
})

// TryAllocSlice allocates a slice of type T with given length and capacity using allocator a.
// Returns a slice of the allocated memory or an error if allocation fails.
// If the allocator is nil, uses the system allocator.
#define mem_TryAllocSlice(T, a, slen, scap) ({                                    \
    mem_Allocator _a = (a);                                                       \
    if (!_a.self) _a = mem_System;                                                \
    if ((slen) > (scap)) so_panic("mem: length exceeds capacity");                \
    so_Result _mem_res = _a.Alloc(_a.self, sizeof(T) * (scap),                    \
                                  alignof(so_typeof(T)));                         \
    so_Slice _slice = {.ptr = _mem_res.val.as_ptr, .len = (slen), .cap = (scap)}; \
    so_Result _slice_res = {.val.as_slice = _slice, .err = _mem_res.err};         \
    _slice_res;                                                                   \
})

// FreeSlice frees a slice previously allocated with [AllocSlice] or [TryAllocSlice].
// If the allocator is nil, uses the system allocator.
#define mem_FreeSlice(T, a, s) ({                                          \
    mem_Allocator _a = (a);                                                \
    if (!_a.self) _a = mem_System;                                         \
    _a.Free(_a.self, (s).ptr, sizeof(T) * (s).cap, alignof(so_typeof(T))); \
})

// nextslicecap computes the capacity for a grown slice using Go's growth
// formula: 2x for small slices (< 256 elements), transitioning to ~1.25x
// for larger ones.
static inline size_t mem_nextslicecap(size_t newLen, size_t oldCap) {
    size_t newcap = oldCap;
    size_t doublecap = newcap + newcap;
    if (newLen > doublecap) return newLen;
    const size_t threshold = 256;
    if (oldCap < threshold) return doublecap;
    for (;;) {
        newcap += (newcap + 3 * threshold) >> 2;
        if (newcap >= newLen) break;
    }
    return newcap;
}

// slicegrow grows a slice's backing allocation to hold at least newLen elements.
// Returns a result with the updated slice or an error if reallocation fails.
// If the allocator is nil, uses the system allocator.
#define mem_slicegrow(a, s, newLen, elemSize, elemAlign) ({                               \
    mem_Allocator _sg_a = (a);                                                            \
    if (!_sg_a.self) _sg_a = mem_System;                                                  \
    so_Slice _sg_s = (s);                                                                 \
    size_t _sg_newLen = (newLen);                                                         \
    so_Result _sg_res = {.val.as_slice = _sg_s, .err = NULL};                             \
    if (_sg_newLen > _sg_s.cap) {                                                         \
        size_t _sg_newcap = mem_nextslicecap(_sg_newLen, _sg_s.cap);                      \
        so_Result _sg_rr = _sg_a.Realloc(_sg_a.self, _sg_s.ptr,                           \
                                         (so_int)(_sg_s.cap * (elemSize)),                \
                                         (so_int)(_sg_newcap * (elemSize)), (elemAlign)); \
        if (_sg_rr.err != NULL) {                                                         \
            _sg_res.err = _sg_rr.err;                                                     \
        } else {                                                                          \
            _sg_s.ptr = _sg_rr.val.as_ptr;                                                \
            _sg_s.cap = _sg_newcap;                                                       \
            _sg_res.val.as_slice = _sg_s;                                                 \
        }                                                                                 \
    }                                                                                     \
    _sg_res;                                                                              \
})

// TryAppend appends elements to a heap-allocated slice, growing it if needed.
// Returns a result with the updated slice or an error if reallocation fails.
// If the allocator is nil, uses the system allocator.
#define mem_TryAppend(T, a, s, ...) ({                               \
    so_Slice _s = (s);                                               \
    T _vals[] = {__VA_ARGS__};                                       \
    size_t _n = sizeof(_vals) / sizeof(T);                           \
    so_Result _gr = mem_slicegrow((a), _s, _s.len + _n,              \
                                  sizeof(T), alignof(so_typeof(T))); \
    if (_gr.err == NULL) {                                           \
        _s = _gr.val.as_slice;                                       \
        memcpy((T*)_s.ptr + _s.len, _vals, sizeof(_vals));           \
        _s.len += _n;                                                \
        _gr.val.as_slice = _s;                                       \
    }                                                                \
    _gr;                                                             \
})

// Append appends elements to a heap-allocated slice, growing it if needed.
// Returns the updated slice or panics on allocation failure.
// If the allocator is nil, uses the system allocator.
#define mem_Append(T, a, s, ...) ({                           \
    so_Result _res = mem_TryAppend(T, (a), (s), __VA_ARGS__); \
    if (_res.err != NULL)                                     \
        so_panic(_res.err->msg);                              \
    _res.val.as_slice;                                        \
})

// TryExtend appends all elements from another slice, growing if needed.
// Returns a result with the updated slice or an error if reallocation fails.
// If the allocator is nil, uses the system allocator.
#define mem_TryExtend(T, a, s, other) ({                             \
    so_Slice _s = (s);                                               \
    so_Slice _src = (other);                                         \
    so_Result _gr = mem_slicegrow((a), _s, _s.len + _src.len,        \
                                  sizeof(T), alignof(so_typeof(T))); \
    if (_gr.err == NULL) {                                           \
        _s = _gr.val.as_slice;                                       \
        memcpy((T*)_s.ptr + _s.len, _src.ptr, _src.len * sizeof(T)); \
        _s.len += _src.len;                                          \
        _gr.val.as_slice = _s;                                       \
    }                                                                \
    _gr;                                                             \
})

// Extend appends all elements from another slice, growing if needed.
// Returns the updated slice or panics on allocation failure.
// If the allocator is nil, uses the system allocator.
#define mem_Extend(T, a, s, other) ({                     \
    so_Result _res = mem_TryExtend(T, (a), (s), (other)); \
    if (_res.err != NULL)                                 \
        so_panic(_res.err->msg);                          \
    _res.val.as_slice;                                    \
})

// MaxAllocaSize is the maximum size that can be allocated with Alloca.
#define mem_MaxAllocaSize so_MaxAllocaSize

// Alloca allocates a block of memory of the given size on the stack.
// The memory is automatically freed when the function that called Alloca returns.
// Panics if the requested size exceeds [MaxAllocaSize].
#define mem_Alloca(size) ({               \
    size_t _size = (size);                \
    so_make_slice(uint8_t, _size, _size); \
})
