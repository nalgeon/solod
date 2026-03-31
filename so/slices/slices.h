#include <stddef.h>
#include "so/builtin/builtin.h"
#include "so/mem/mem.h"

// nextcap computes the capacity for a grown slice using Go's growth
// formula: 2x for small slices (< 256 elements), transitioning to ~1.25x
// for larger ones.
static inline size_t slices_nextcap(size_t newLen, size_t oldCap) {
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

// grow grows a slice's backing allocation to hold at least newLen elements.
// Returns a result with the updated slice or an error if reallocation fails.
// If the allocator is nil, uses the system allocator.
static inline so_R_slice_err slices_grow(mem_Allocator a, so_Slice s, size_t newLen,
                                         size_t elemSize, so_int elemAlign) {
    if (!a.self) a = mem_System;
    so_R_slice_err res = {.val = s, .err = NULL};
    if (newLen > s.cap) {
        size_t newcap = slices_nextcap(newLen, s.cap);
        so_R_ptr_err rr = a.Realloc(a.self, s.ptr,
                                    (so_int)(s.cap * elemSize),
                                    (so_int)(newcap * elemSize), elemAlign);
        if (rr.err != NULL) {
            res.err = rr.err;
        } else {
            s.ptr = rr.val;
            s.cap = newcap;
            res.val = s;
        }
    }
    return res;
}

// Make allocates a slice of type T with given length using allocator a.
// Returns an allocated slice or panics on failure.
// If the allocator is nil, uses the system allocator.
#define slices_Make(T, a, len) mem_AllocSlice(T, (a), (len), (len))

// MakeCap allocates a slice of type T with given length and capacity using allocator a.
// Returns an allocated slice or panics on failure.
// If the allocator is nil, uses the system allocator.
#define slices_MakeCap(T, a, len, cap) mem_AllocSlice(T, (a), (len), (cap))

// Free frees a previously allocated slice.
// If the allocator is nil, uses the system allocator.
#define slices_Free(T, a, s) mem_FreeSlice(T, (a), (s))

// TryAppend appends elements to a heap-allocated slice, growing it if needed.
// Returns a result with the updated slice or an error if reallocation fails.
// If the allocator is nil, uses the system allocator.
#define slices_TryAppend(T, a, s, ...) ({                               \
    so_Slice _s = (s);                                                  \
    T _vals[] = {__VA_ARGS__};                                          \
    size_t _n = sizeof(_vals) / sizeof(T);                              \
    so_R_slice_err _gr = slices_grow((a), _s, _s.len + _n,              \
                                     sizeof(T), alignof(so_typeof(T))); \
    if (_gr.err == NULL) {                                              \
        _s = _gr.val;                                                   \
        memcpy((T*)_s.ptr + _s.len, _vals, sizeof(_vals));              \
        _s.len += _n;                                                   \
        _gr.val = _s;                                                   \
    }                                                                   \
    _gr;                                                                \
})

// Append appends elements to a heap-allocated slice, growing it if needed.
// Returns the updated slice or panics on allocation failure.
// If the allocator is nil, uses the system allocator.
#define slices_Append(T, a, s, ...) ({                                \
    so_R_slice_err _res = slices_TryAppend(T, (a), (s), __VA_ARGS__); \
    if (_res.err != NULL)                                             \
        so_panic(_res.err->msg);                                      \
    _res.val;                                                         \
})

// TryExtend appends all elements from another slice, growing if needed.
// Returns a result with the updated slice or an error if reallocation fails.
// If the allocator is nil, uses the system allocator.
#define slices_TryExtend(T, a, s, other) ({                             \
    so_Slice _s = (s);                                                  \
    so_Slice _src = (other);                                            \
    so_R_slice_err _gr = slices_grow((a), _s, _s.len + _src.len,        \
                                     sizeof(T), alignof(so_typeof(T))); \
    if (_gr.err == NULL) {                                              \
        _s = _gr.val;                                                   \
        memcpy((T*)_s.ptr + _s.len, _src.ptr, _src.len * sizeof(T));    \
        _s.len += _src.len;                                             \
        _gr.val = _s;                                                   \
    }                                                                   \
    _gr;                                                                \
})

// Extend appends all elements from another slice, growing if needed.
// Returns the updated slice or panics on allocation failure.
// If the allocator is nil, uses the system allocator.
#define slices_Extend(T, a, s, other) ({                          \
    so_R_slice_err _res = slices_TryExtend(T, (a), (s), (other)); \
    if (_res.err != NULL)                                         \
        so_panic(_res.err->msg);                                  \
    _res.val;                                                     \
})

// Clone returns a shallow copy of the slice.
// The returned slice is heap-allocated; the caller owns it.
#define slices_Clone(T, a, s) ({                                 \
    so_Slice _s = (s);                                           \
    so_Slice _newSlice = mem_AllocSlice(T, (a), _s.len, _s.len); \
    memcpy(_newSlice.ptr, _s.ptr, _s.len * sizeof(T));           \
    _newSlice;                                                   \
})

// Equal reports whether two slices are equal: the same length and all
// elements equal. Empty and nil slices are considered equal.
#define slices_Equal(T, s1, s2) ({                                                \
    so_Slice _s1 = (s1);                                                          \
    so_Slice _s2 = (s2);                                                          \
    size_t _size = _s1.len * sizeof(T);                                           \
    (_s1.len == _s2.len) && (_size == 0 || memcmp(_s1.ptr, _s2.ptr, _size) == 0); \
})
