#pragma once

#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#define SO_CONCAT(a, b) a##b
#define SO_NAME(a, b) SO_CONCAT(a, b)

#define so_auto __auto_type
#define so_byte uint8_t
#define so_rune int32_t
typedef int64_t so_int;

// so_String is a pointer to array of bytes plus a length.
typedef struct {
    const char* ptr;
    size_t len;
} so_String;

// so_strlit creates a String from a string literal.
#define so_strlit(s) ((so_String){s, sizeof(s) - 1})

// so_string_eq returns true if two strings are equal.
#define so_string_eq(s1, s2) so_string_eq_impl(s1, s2)
static inline bool so_string_eq_impl(so_String s1, so_String s2) {
    return s1.len == s2.len && memcmp(s1.ptr, s2.ptr, s1.len) == 0;
}
// so_string_ne returns true if two strings are not equal.
#define so_string_ne(s1, s2) (!so_string_eq(s1, s2))
// so_string_lt returns true if s1 < s2 in lexicographical order.
#define so_string_lt(s1, s2) so_string_lt_impl(s1, s2)
static inline bool so_string_lt_impl(so_String s1, so_String s2) {
    size_t n = s1.len < s2.len ? s1.len : s2.len;
    int cmp = memcmp(s1.ptr, s2.ptr, n);
    return cmp < 0 || (cmp == 0 && s1.len < s2.len);
}
// so_string_lte returns true if s1 <= s2 in lexicographical order.
#define so_string_lte(s1, s2) (so_string_lt(s1, s2) || so_string_eq(s1, s2))
// so_string_gt returns true if s1 > s2 in lexicographical order.
#define so_string_gt(s1, s2) so_string_gt_impl(s1, s2)
static inline bool so_string_gt_impl(so_String s1, so_String s2) {
    size_t n = s1.len < s2.len ? s1.len : s2.len;
    int cmp = memcmp(s1.ptr, s2.ptr, n);
    return cmp > 0 || (cmp == 0 && s1.len > s2.len);
}
// so_string_gte returns true if s1 >= s2 in lexicographical order.
#define so_string_gte(s1, s2) (so_string_gt(s1, s2) || so_string_eq(s1, s2))

// so_utf8_decode decodes one UTF-8 rune from string s at byte offset i.
// Stores the byte width in *w.
// Returns the decoded rune, or 0xFFFD for invalid UTF-8.
so_rune so_utf8_decode(so_String s, so_int i, int* w);

// so_Slice is a pointer to array of elements plus a length.
typedef struct {
    void* ptr;
    size_t len;
    size_t cap;
} so_Slice;

// so_make_slice creates a slice of the given type, length, and capacity.
// Uses separate declaration and initialization to avoid
// variable-length array initialization, which is not allowed in C.
#define so_make_slice(type, len, cap) ({ \
    type _arr[(cap)];                    \
    memset(_arr, 0, sizeof(_arr));       \
    (so_Slice){_arr, (len), (cap)};      \
})

// so_slice creates a slice from an array or another slice
// from index 'from' (inclusive) to index 'to' (exclusive).
#define so_slice(s, T, from, to) ((so_Slice){(T*)(s).ptr + (from), (to) - (from), (s).cap - (from)})

// so_string_bytes wraps a string's raw bytes as a byte slice.
#define so_string_bytes(s) ((so_Slice){(void*)(s).ptr, (s).len, (s).len})

// so_string_runes decodes a string's UTF-8 bytes into a rune slice.
#define so_string_runes(s, maxlen) ({ \
    int32_t _buf[(maxlen)];           \
    so_string_runes_impl((s), _buf);  \
})
so_Slice so_string_runes_impl(so_String s, int32_t* buf);

// so_append appends elements to a slice without resizing.
// Returns the new slice with updated length.
// Panics if the new length exceeds the capacity.
#define so_append(s, T, ...) ({                                    \
    so_Slice _s = (s);                                             \
    T _vals[] = {__VA_ARGS__};                                     \
    size_t _n = sizeof(_vals) / sizeof(T);                         \
    if (_s.len + _n > _s.cap) so_panic("append: out of capacity"); \
    memcpy((T*)_s.ptr + _s.len, _vals, sizeof(_vals));             \
    _s.len += _n;                                                  \
    _s;                                                            \
})

// so_index returns a reference to the element at index i in a slice or string.
#define so_index(s, T, i) (((T*)(s).ptr)[i])

// so_len returns the length of a slice or string.
#define so_len(s) ((so_int)(s).len)

// so_cap returns the capacity of a slice.
#define so_cap(s) ((so_int)(s).cap)

// so_copy copies elements from src to dst. Returns the number of elements copied
// (which is the minimum of dst.len and src.len).
static inline so_int so_copy_impl(so_Slice dst, so_Slice src, size_t elem_size) {
    size_t n = dst.len < src.len ? dst.len : src.len;
    memmove(dst.ptr, src.ptr, n * elem_size);
    return (so_int)n;
}
#define so_copy(dst, src, T) so_copy_impl(dst, src, sizeof(T))

// so_Error is a pointer to an error message string, or NULL for no error.
// Errors are immutable and compared by pointer equality.
struct so_Error_ {
    const char* msg;
};
typedef struct so_Error_* so_Error;

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

// errors_New creates a new error with the given message string.
// so_Error errors_New(so_String s)
#define errors_New(s) (&(struct so_Error_){s.ptr})

// so_panic aborts the program with the given message.
void so_panic(const char* msg);

// Defer.

// so_Deferred is a deferred function and its argument.
struct so_Deferred {
    void (*fn)(void*);
    void* arg;
};

// so_defer_cleanup calls the deferred function with its argument.
static inline void so_defer_cleanup(struct so_Deferred* ctx) {
    if (ctx->fn) ctx->fn(ctx->arg);
}

// so_defer creates a deferred function call for the current scope.
#define so_defer(fn, ptr)                                \
    struct so_Deferred SO_NAME(_defer_var_, __COUNTER__) \
        __attribute__((cleanup(so_defer_cleanup))) =     \
            {(void (*)(void*))(fn), (void*)(ptr)}

// Printing.

// so_print writes the formatted string to stdout.
// Returns the number of bytes written.
int so_print(const char* format, ...);

// so_println writes the formatted string to stdout with a newline.
// Returns the number of bytes written.
int so_println(const char* format, ...);
