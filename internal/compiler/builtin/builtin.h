#pragma once

#include <alloca.h>
#include <inttypes.h>
#include <stdalign.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// --- General utilities ---
#define SO_CONCAT(a, b) a##b
#define SO_NAME(a, b) SO_CONCAT(a, b)

#define so_typeof __typeof__
#define so_auto __auto_type

typedef uint8_t so_byte;
typedef int32_t so_rune;
typedef int64_t so_int;

// --- Alloca safety ---

// MaxAllocaSize is the maximum size that can be
// allocated with alloca (64 KB by default).
#ifndef so_MaxAllocaSize
#define so_MaxAllocaSize (64 << 10)
#endif

#define so_alloca(size) ({                                \
    size_t _size = (size);                                \
    if (_size > so_MaxAllocaSize)                         \
        so_panic("alloca: size exceeds maximum allowed"); \
    _size ? alloca(_size) : NULL;                         \
})

// --- String type ---

// String is a pointer to array of bytes plus a length.
typedef struct {
    const char* ptr;
    size_t len;
} so_String;

// strlit creates a String from a string literal.
#define so_str(s) ((so_String){s, sizeof(s) - 1})

// cstr returns a null-terminated C string copy on the stack.
#define so_cstr(s) ({                   \
    so_String _s = (s);                 \
    char* _buf = so_alloca(_s.len + 1); \
    memcpy(_buf, _s.ptr, _s.len);       \
    _buf[_s.len] = '\0';                \
    _buf;                               \
})

// string_slice creates a substring [from, to).
#define so_string_slice(s, from, to) ({        \
    so_String _s = (s);                        \
    size_t _from = (size_t)(from);             \
    size_t _to = (size_t)(to);                 \
    if (_to > _s.len || _from > _to)           \
        so_panic("slice bounds out of range"); \
    (so_String){_s.ptr + _from, _to - _from};  \
})

// string_eq returns true if two strings are equal.
static inline bool so_string_eq(so_String s1, so_String s2) {
    return s1.len == s2.len && memcmp(s1.ptr, s2.ptr, s1.len) == 0;
}

// string_ne returns true if two strings are not equal.
static inline bool so_string_ne(so_String s1, so_String s2) {
    return !so_string_eq(s1, s2);
}

// string_lt returns true if s1 < s2 in lexicographical order.
static inline bool so_string_lt(so_String s1, so_String s2) {
    size_t n = s1.len < s2.len ? s1.len : s2.len;
    int cmp = memcmp(s1.ptr, s2.ptr, n);
    return cmp < 0 || (cmp == 0 && s1.len < s2.len);
}

// string_lte returns true if s1 <= s2 in lexicographical order.
static inline bool so_string_lte(so_String s1, so_String s2) {
    return so_string_lt(s1, s2) || so_string_eq(s1, s2);
}

// string_gt returns true if s1 > s2 in lexicographical order.
static inline bool so_string_gt(so_String s1, so_String s2) {
    size_t n = s1.len < s2.len ? s1.len : s2.len;
    int cmp = memcmp(s1.ptr, s2.ptr, n);
    return cmp > 0 || (cmp == 0 && s1.len > s2.len);
}

// string_gte returns true if s1 >= s2 in lexicographical order.
static inline bool so_string_gte(so_String s1, so_String s2) {
    return so_string_gt(s1, s2) || so_string_eq(s1, s2);
}

// utf8_decode decodes one UTF-8 rune from string s at byte offset i.
// Stores the byte width in *w.
// Returns the decoded rune, or 0xFFFD for invalid UTF-8.
so_rune so_utf8_decode(so_String s, so_int i, int* w);

// --- Arrays ---

// array_eq returns true if two arrays are equal.
static inline bool so_array_eq(const void* a, const void* b, size_t size) {
    return memcmp(a, b, size) == 0;
}

// array_ne returns true if two arrays are not equal.
static inline bool so_array_ne(const void* a, const void* b, size_t size) {
    return memcmp(a, b, size) != 0;
}

// array_slice creates a slice from a C array.
// 'size' is the total array size (known at compile time).
#define so_array_slice(T, arr, from, to, size) \
    ((so_Slice){(T*)(arr) + (from), (to) - (from), (size) - (from)})

// --- Slice type ---

// Slice is a pointer to array of elements plus a length.
typedef struct {
    void* ptr;
    size_t len;
    size_t cap;
} so_Slice;

// make_slice creates a zero-initialized slice on the stack.
// Allocates memory on the stack until the calling function returns.
#define so_make_slice(T, len, cap) ({ \
    size_t _cap = (cap);              \
    size_t _n = sizeof(T) * _cap;     \
    void* _p = so_alloca(_n);         \
    if (_p) memset(_p, 0, _n);        \
    (so_Slice){_p, (len), _cap};      \
})

// slice creates a slice from another slice
// from index 'from' (inclusive) to index 'to' (exclusive).
#define so_slice(T, s, from, to) ({                                \
    size_t _from = (size_t)(from);                                 \
    size_t _to = (size_t)(to);                                     \
    if (_to > (s).len || _from > _to)                              \
        so_panic("slice bounds out of range");                     \
    (so_Slice){(T*)(s).ptr + _from, _to - _from, (s).cap - _from}; \
})

// string_bytes copies a string's bytes into a byte slice.
// Allocates memory on the stack until the calling function returns.
#define so_string_bytes(s) ({               \
    so_String _s = (s);                     \
    void* _buf = so_alloca(_s.len);         \
    if (_buf) memcpy(_buf, _s.ptr, _s.len); \
    (so_Slice){_buf, _s.len, _s.len};       \
})

// string_runes decodes a string's UTF-8 bytes into a rune slice.
// Allocates memory on the stack until the calling function returns.
#define so_string_runes(s, maxlen) ({                      \
    int32_t* _buf = so_alloca((maxlen) * sizeof(int32_t)); \
    so_string_runes_impl((s), _buf);                       \
})
so_Slice so_string_runes_impl(so_String s, int32_t* buf);

// bytes_string copies a byte slice into a string.
// Allocates memory on the stack until the calling function returns.
#define so_bytes_string(bs) ({                \
    so_Slice _bs = (bs);                      \
    char* _buf = so_alloca(_bs.len);          \
    if (_buf) memcpy(_buf, _bs.ptr, _bs.len); \
    (so_String){_buf, _bs.len};               \
})

// runes_string encodes a rune slice into a UTF-8 string.
// Allocates memory on the stack until the calling function returns.
#define so_runes_string(rs) ({           \
    so_Slice _rs = (rs);                 \
    char* _buf = so_alloca(_rs.len * 4); \
    so_runes_string_impl(_rs, _buf);     \
})
so_String so_runes_string_impl(so_Slice rs, char* buf);

// append appends elements to a slice without resizing.
// Returns the new slice with updated length.
// Panics if the new length exceeds the capacity.
#define so_append(T, s, ...) ({                                    \
    so_Slice _s = (s);                                             \
    T _vals[] = {__VA_ARGS__};                                     \
    size_t _n = sizeof(_vals) / sizeof(T);                         \
    if (_s.len + _n > _s.cap) so_panic("append: out of capacity"); \
    memcpy((T*)_s.ptr + _s.len, _vals, sizeof(_vals));             \
    _s.len += _n;                                                  \
    _s;                                                            \
})

// extend appends all elements from a source slice to a destination slice.
// Returns the new slice with updated length.
// Panics if the new length exceeds the capacity.
#define so_extend(T, dst, src) ({                                            \
    so_Slice _dst = (dst);                                                   \
    so_Slice _src = (src);                                                   \
    if (_dst.len + _src.len > _dst.cap) so_panic("extend: out of capacity"); \
    memcpy((T*)_dst.ptr + _dst.len, _src.ptr, _src.len * sizeof(T));         \
    _dst.len += _src.len;                                                    \
    _dst;                                                                    \
})

// copy copies elements from src to dst. Returns the number of elements copied
// (which is the minimum of dst.len and src.len).
#define so_copy(T, dst, src) so_copy_impl(dst, src, sizeof(T))
static inline so_int so_copy_impl(so_Slice dst, so_Slice src, size_t elem_size) {
    size_t n = dst.len < src.len ? dst.len : src.len;
    memmove(dst.ptr, src.ptr, n * elem_size);
    return (so_int)n;
}

// at returns a reference to the element at index i in a slice or string.
#define so_at(T, s, i) (*so_at_ptr(T, s, i))
#define so_at_ptr(T, s, i) ({            \
    size_t _i = (size_t)(i);             \
    if (_i >= (s).len)                   \
        so_panic("index out of bounds"); \
    (T*)(s).ptr + _i;                    \
})

// len returns the length of a slice or string.
#define so_len(s) ((so_int)(s).len)

// cap returns the capacity of a slice.
#define so_cap(s) ((so_int)(s).cap)

// --- Min/Max ---

// min returns the smaller of two values.
#define so_min(a, b) ({    \
    so_typeof(a) _a = (a); \
    so_typeof(b) _b = (b); \
    _a < _b ? _a : _b;     \
})

// max returns the larger of two values.
#define so_max(a, b) ({    \
    so_typeof(a) _a = (a); \
    so_typeof(b) _b = (b); \
    _a > _b ? _a : _b;     \
})

// string_min returns the lexicographically smaller string.
static inline so_String so_string_min(so_String a, so_String b) {
    return so_string_lt(a, b) ? a : b;
}

// string_max returns the lexicographically larger string.
static inline so_String so_string_max(so_String a, so_String b) {
    return so_string_gt(a, b) ? a : b;
}

// --- Error type ---

// Error is a pointer to an error message string, or NULL for no error.
// Errors are immutable and compared by pointer equality.
struct so_Error_ {
    const char* msg;
};
typedef struct so_Error_* so_Error;

// errors_New creates a new error with the given message string.
// so_Error errors_New(const char* s)
#define errors_New(s) (&(struct so_Error_){s})

// panic aborts the program with the given message.
#define so_panic(msg)                                     \
    do {                                                  \
        fprintf(stderr, "panic: %s\n  %s:%d (func %s)\n", \
                msg, __FILE__, __LINE__, __func__);       \
        exit(1);                                          \
    } while (0)

// --- Result type ---

// so_Result is the return type for functions that return (T, error).
typedef struct {
    union {
        bool as_bool;
        uint8_t as_byte;
        int32_t as_rune;
        so_int as_int;
        double as_double;
        so_String as_string;
        so_Slice as_slice;
        void* as_ptr;
    } val;
    so_Error err;
} so_Result;

// --- Defer ---

// Deferred is a deferred function and its argument.
struct so_Deferred {
    void (*fn)(void*);
    void* arg;
};

// defer_cleanup calls the deferred function with its argument.
static inline void so_defer_cleanup(struct so_Deferred* ctx) {
    if (ctx->fn) ctx->fn(ctx->arg);
}

// defer creates a deferred function call for the current scope.
#define so_defer(fn, ptr)                                \
    struct so_Deferred SO_NAME(_defer_var_, __COUNTER__) \
        __attribute__((cleanup(so_defer_cleanup))) =     \
            {(void (*)(void*))(fn), (void*)(ptr)}

// --- Printing ---

// print writes the formatted string to stdout.
// Returns the number of bytes written.
int so_print(const char* format, ...);

// println writes the formatted string to stdout with a newline.
// Returns the number of bytes written.
int so_println(const char* format, ...);

// --- Unsafe ---

#define unsafe_Alignof(x) alignof(so_typeof(x))
#define unsafe_Sizeof(x) sizeof(x)

static inline so_String unsafe_String(void* ptr, size_t len) {
    return (so_String){(const char*)ptr, len};
}
static inline uint8_t* unsafe_StringData(so_String s) {
    return (uint8_t*)s.ptr;
}
static inline so_Slice unsafe_Slice(void* ptr, size_t len) {
    return (so_Slice){ptr, len, len};
}
static inline void* unsafe_SliceData(so_Slice s) {
    return s.ptr;
}
