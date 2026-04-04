#pragma once

#ifdef _WIN32
#include <malloc.h>
#else
#include <alloca.h>
#endif

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
typedef uint64_t so_uint;

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
#define so_cstr(s) ({                             \
    so_String _s = (s);                           \
    char* _buf = so_alloca(_s.len + 1);           \
    if (_s.len > 0) memcpy(_buf, _s.ptr, _s.len); \
    _buf[_s.len] = '\0';                          \
    _buf;                                         \
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

// string_add concatenates two strings.
// Allocates memory on the stack until the calling function returns.
#define so_string_add(s1, s2) ({                               \
    so_String _s1 = (s1);                                      \
    so_String _s2 = (s2);                                      \
    size_t _total = _s1.len + _s2.len;                         \
    char* _buf = so_alloca(_total);                            \
    if (_s1.len > 0) memcpy(_buf, _s1.ptr, _s1.len);           \
    if (_s2.len > 0) memcpy(_buf + _s1.len, _s2.ptr, _s2.len); \
    (so_String){_buf, _total};                                 \
})

// string_eq returns true if two strings are equal.
static inline bool so_string_eq(so_String s1, so_String s2) {
    return s1.len == s2.len && (s1.len == 0 || memcmp(s1.ptr, s2.ptr, s1.len) == 0);
}

// string_ne returns true if two strings are not equal.
static inline bool so_string_ne(so_String s1, so_String s2) {
    return !so_string_eq(s1, s2);
}

// string_lt returns true if s1 < s2 in lexicographical order.
static inline bool so_string_lt(so_String s1, so_String s2) {
    size_t n = s1.len < s2.len ? s1.len : s2.len;
    int cmp = n > 0 ? memcmp(s1.ptr, s2.ptr, n) : 0;
    return cmp < 0 || (cmp == 0 && s1.len < s2.len);
}

// string_lte returns true if s1 <= s2 in lexicographical order.
static inline bool so_string_lte(so_String s1, so_String s2) {
    return so_string_lt(s1, s2) || so_string_eq(s1, s2);
}

// string_gt returns true if s1 > s2 in lexicographical order.
static inline bool so_string_gt(so_String s1, so_String s2) {
    size_t n = s1.len < s2.len ? s1.len : s2.len;
    int cmp = n > 0 ? memcmp(s1.ptr, s2.ptr, n) : 0;
    return cmp > 0 || (cmp == 0 && s1.len > s2.len);
}

// string_gte returns true if s1 >= s2 in lexicographical order.
static inline bool so_string_gte(so_String s1, so_String s2) {
    return so_string_gt(s1, s2) || so_string_eq(s1, s2);
}

// utf8_decode decodes one UTF-8 rune from string s at byte offset i.
// Stores the byte width in *w.
// Returns the decoded rune, or 0xFFFD for invalid UTF-8.
so_rune so_utf8_decode(so_String s, so_int i, so_int* w);

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

// array_slice3 creates a slice from a C array with an explicit capacity.
#define so_array_slice3(T, arr, from, to, max) \
    ((so_Slice){(T*)(arr) + (from), (to) - (from), (max) - (from)})

// --- Slice type ---

// Slice is a pointer to array of elements plus a length.
typedef struct {
    void* ptr;
    size_t len;
    size_t cap;
} so_Slice;

// Nil sentinel: address used as the pointer for nil/empty slices.
// Non-NULL to satisfy static analyzers; never dereferenced.
extern so_byte so_Nil[];

// make_slice creates a zero-initialized slice on the stack.
// Allocates memory on the stack until the calling function returns.
#define so_make_slice(T, len, cap) ({        \
    size_t _cap = (cap);                     \
    size_t _n = sizeof(T) * _cap;            \
    void* _p = _n ? so_alloca(_n) : &so_Nil; \
    if (_n) memset(_p, 0, _n);               \
    (so_Slice){_p, (len), _cap};             \
})

// slice creates a slice from another slice
// from index 'from' (inclusive) to index 'to' (exclusive).
#define so_slice(T, s, from, to) ({                              \
    so_Slice _s = (s);                                           \
    size_t _from = (size_t)(from);                               \
    size_t _to = (size_t)(to);                                   \
    if (_to > _s.cap || _from > _to)                             \
        so_panic("slice bounds out of range");                   \
    (so_Slice){(T*)_s.ptr + _from, _to - _from, _s.cap - _from}; \
})

// slice3 creates a slice from another slice with an explicit capacity.
#define so_slice3(T, s, from, to, max) ({                      \
    so_Slice _s = (s);                                         \
    size_t _from = (size_t)(from);                             \
    size_t _to = (size_t)(to);                                 \
    size_t _max = (size_t)(max);                               \
    if (_max > _s.cap || _to > _max || _from > _to)            \
        so_panic("slice bounds out of range");                 \
    (so_Slice){(T*)_s.ptr + _from, _to - _from, _max - _from}; \
})

// string_bytes reinterprets a string as a byte slice (zero-copy).
#define so_string_bytes(s) ({                  \
    so_String _s = (s);                        \
    (so_Slice){(void*)_s.ptr, _s.len, _s.len}; \
})

// string_runes decodes a string's UTF-8 bytes into a rune slice.
// Allocates memory on the stack until the calling function returns.
#define so_string_runes(s) ({                              \
    so_String _s = (s);                                    \
    so_rune* _buf = so_alloca((_s.len) * sizeof(so_rune)); \
    so_string_runes_impl(_s, _buf);                        \
})
so_Slice so_string_runes_impl(so_String s, so_rune* buf);

// bytes_string reinterprets a byte slice as a string (zero-copy).
#define so_bytes_string(bs) ({                  \
    so_Slice _bs = (bs);                        \
    (so_String){(const char*)_bs.ptr, _bs.len}; \
})

// runes_string encodes a rune slice into a UTF-8 string.
// Allocates memory on the stack until the calling function returns.
#define so_runes_string(rs) ({           \
    so_Slice _rs = (rs);                 \
    char* _buf = so_alloca(_rs.len * 4); \
    so_runes_string_impl(_rs, _buf);     \
})
so_String so_runes_string_impl(so_Slice rs, char* buf);

// utf8_encode encodes a single rune into buf (up to 4 bytes).
// Returns the number of bytes written.
size_t so_utf8_encode(so_rune r, char* buf);

// byte_string creates a string from a single byte.
// Allocates memory on the stack until the calling function returns.
#define so_byte_string(b) ({   \
    char* _buf = so_alloca(1); \
    _buf[0] = (char)(b);       \
    (so_String){_buf, 1};      \
})

// rune_string creates a UTF-8 string from a single rune.
// Allocates memory on the stack until the calling function returns.
#define so_rune_string(r) ({                        \
    char* _buf = so_alloca(4);                      \
    size_t _n = so_utf8_encode((so_rune)(r), _buf); \
    (so_String){_buf, _n};                          \
})

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
#define so_extend(T, dst, src) ({                             \
    so_Slice _dst = (dst);                                    \
    so_Slice _src = (src);                                    \
    if (_dst.len + _src.len > _dst.cap)                       \
        so_panic("extend: out of capacity");                  \
    if (_src.len > 0) memcpy((T*)_dst.ptr + _dst.len,         \
                             _src.ptr, _src.len * sizeof(T)); \
    _dst.len += _src.len;                                     \
    _dst;                                                     \
})

// copy copies elements from src to dst. Returns the number of elements copied
// (which is the minimum of dst.len and src.len).
#define so_copy(T, dst, src) so_copy_impl(dst, src, sizeof(T))
static inline so_int so_copy_impl(so_Slice dst, so_Slice src, size_t elem_size) {
    size_t n = dst.len < src.len ? dst.len : src.len;
    if (n > 0) memmove(dst.ptr, src.ptr, n * elem_size);
    return (so_int)n;
}

// copy_string copies bytes from a string to a byte slice. Returns the number
// of bytes copied (which is the minimum of dst.len and src.len).
static inline so_int so_copy_string(so_Slice dst, so_String src) {
    size_t n = dst.len < src.len ? dst.len : src.len;
    if (n > 0) memmove(dst.ptr, src.ptr, n);
    return (so_int)n;
}

// clear sets all elements up to the length
// of the slice to their zero value.
#define so_clear(T, s) ({                  \
    so_Slice _s = (s);                     \
    memset(_s.ptr, 0, _s.len * sizeof(T)); \
    _s;                                    \
})

// at returns a reference to the element at index i in a slice or string.
#define so_at(T, s, i) (*so_at_ptr(T, s, i))
#define so_at_ptr(T, s, i) ({            \
    so_auto _s_at = (s);                 \
    size_t _i = (size_t)(i);             \
    if (_i >= _s_at.len)                 \
        so_panic("index out of bounds"); \
    (T*)_s_at.ptr + _i;                  \
})

// len returns the length of a slice or string.
#define so_len(s) ((so_int)(s).len)

// cap returns the capacity of a slice.
#define so_cap(s) ((so_int)(s).cap)

// --- Map type ---

// Map is an open-addressed hash table with MSI (mask-step-index) probing.
typedef struct {
    void* keys;
    void* vals;
    uint8_t* used;  // 0=empty, 1=occupied
    size_t len;
    size_t cap;  // always power of 2
} so_Map;

// key_hash hashes a map key to a 64-bit value (FNV-1a).
// The seed is the map's own address (randomized by ASLR).
static inline uint64_t so_key_hash_default(const void* ptr, size_t n, uint64_t seed) {
    const uint8_t* p = (const uint8_t*)ptr;
    uint64_t h = seed;
    for (size_t i = 0; i < n; i++) {
        h ^= p[i];
        h *= 0x100000001b3ULL;
    }
    return h;
}
static inline uint64_t so_key_hash_string(const void* ptr, size_t n, uint64_t seed) {
    (void)n;
    const so_String* s = (const so_String*)ptr;
    return so_key_hash_default(s->ptr, s->len, seed);
}

#define so_key_hash(k, seed) ({                                             \
    so_typeof(k) _kh = (k);                                                 \
    _Generic((_kh),                                                         \
        so_String: so_key_hash_string,                                      \
        default: so_key_hash_default)(&_kh, sizeof(_kh), (uint64_t)(seed)); \
})

// key_eq compares two map keys for equality.
// Uses so_string_eq for strings, memcmp for everything else.
static inline bool so_key_eq_default(const void* a, const void* b, size_t n) {
    return memcmp(a, b, n) == 0;
}
static inline bool so_key_eq_string(const void* a, const void* b, size_t n) {
    (void)n;
    return so_string_eq(*(const so_String*)a, *(const so_String*)b);
}

#define so_key_eq(a, b) ({                                    \
    so_typeof(a) _ka = (a);                                   \
    so_typeof(a) _kb = (b);                                   \
    _Generic((_ka),                                           \
        so_String: so_key_eq_string,                          \
        default: so_key_eq_default)(&_ka, &_kb, sizeof(_ka)); \
})

// map_nextpow2 rounds up to the next power of 2.
static inline size_t so_map_nextpow2(size_t n) {
    if (n == 0) return 1;
    n--;
    n |= n >> 1;
    n |= n >> 2;
    n |= n >> 4;
    n |= n >> 8;
    n |= n >> 16;
    n |= n >> 32;
    return n + 1;
}

// map_cap computes the internal capacity for n elements (keeps load <= 75%).
static inline size_t so_map_cap(size_t n) {
    if (n == 0) return 0;
    return so_map_nextpow2(n + n / 3 + 1);
}

// make_map creates a zero-initialized map on the stack.
#define so_make_map(K, V, n) ({              \
    size_t _n = (n);                         \
    if (_n == 0)                             \
        so_panic("map: zero capacity");      \
    size_t _cap = so_map_cap(_n);            \
    size_t _ksz = sizeof(K) * _cap;          \
    size_t _vsz = sizeof(V) * _cap;          \
    size_t _usz = sizeof(uint8_t) * _cap;    \
    void* _kp = so_alloca(_ksz);             \
    void* _vp = so_alloca(_vsz);             \
    uint8_t* _up = so_alloca(_usz);          \
    if (_kp) memset(_kp, 0, _ksz);           \
    if (_vp) memset(_vp, 0, _vsz);           \
    if (_up) memset(_up, 0, _usz);           \
    so_Map* _mp = so_alloca(sizeof(so_Map)); \
    *_mp = (so_Map){_kp, _vp, _up, 0, _cap}; \
    _mp;                                     \
})

// map_set inserts or updates a key-value pair in the map.
// Panics if the map is full and the key is not found.
#define so_map_set(K, V, m, key, val)                  \
    do {                                               \
        K _k = (key);                                  \
        V _v = (val);                                  \
        so_Map* _m = (m);                              \
        uint64_t _h = so_key_hash(_k, (uintptr_t)_m);  \
        size_t _mask = _m->cap - 1;                    \
        size_t _step = (size_t)(_h >> 32) | 1;         \
        size_t _idx = (size_t)_h & _mask;              \
        for (size_t _p = 0;; _p++) {                   \
            if (_p >= _m->cap)                         \
                so_panic("map: out of capacity");      \
            if (!_m->used[_idx]) {                     \
                ((K*)_m->keys)[_idx] = _k;             \
                ((V*)_m->vals)[_idx] = _v;             \
                _m->used[_idx] = 1;                    \
                _m->len++;                             \
                break;                                 \
            }                                          \
            if (so_key_eq(((K*)_m->keys)[_idx], _k)) { \
                ((V*)_m->vals)[_idx] = _v;             \
                break;                                 \
            }                                          \
            _idx = (_idx + _step) & _mask;             \
        }                                              \
    } while (0)

// map_get returns the value for the given key, or zero if not found.
#define so_map_get(K, V, m, key) ({                    \
    K _k = (key);                                      \
    const so_Map* _m = (m);                            \
    V _v = {0};                                        \
    if (_m->cap > 0) {                                 \
        uint64_t _h = so_key_hash(_k, (uintptr_t)_m);  \
        size_t _mask = _m->cap - 1;                    \
        size_t _step = (size_t)(_h >> 32) | 1;         \
        size_t _idx = (size_t)_h & _mask;              \
        for (size_t _p = 0; _p < _m->cap; _p++) {      \
            if (!_m->used[_idx]) break;                \
            if (so_key_eq(((K*)_m->keys)[_idx], _k)) { \
                _v = ((V*)_m->vals)[_idx];             \
                break;                                 \
            }                                          \
            _idx = (_idx + _step) & _mask;             \
        }                                              \
    }                                                  \
    _v;                                                \
})

// map_has returns true if the map contains the given key.
#define so_map_has(K, m, key) ({                       \
    K _k = (key);                                      \
    const so_Map* _m = (m);                            \
    bool _found = false;                               \
    if (_m->cap > 0) {                                 \
        uint64_t _h = so_key_hash(_k, (uintptr_t)_m);  \
        size_t _mask = _m->cap - 1;                    \
        size_t _step = (size_t)(_h >> 32) | 1;         \
        size_t _idx = (size_t)_h & _mask;              \
        for (size_t _p = 0; _p < _m->cap; _p++) {      \
            if (!_m->used[_idx]) break;                \
            if (so_key_eq(((K*)_m->keys)[_idx], _k)) { \
                _found = true;                         \
                break;                                 \
            }                                          \
            _idx = (_idx + _step) & _mask;             \
        }                                              \
    }                                                  \
    _found;                                            \
})

// map_lit creates a map from literal key/value arrays.
#define so_map_lit(K, V, n, keys, vals) ({                     \
    size_t _ml_n = (n);                                        \
    so_Map* _ml_m = so_make_map(K, V, _ml_n);                  \
    K* _ml_ks = (keys);                                        \
    V* _ml_vs = (vals);                                        \
    for (size_t _ml_i = 0; _ml_i < _ml_n; _ml_i++)             \
        so_map_set(K, V, _ml_m, _ml_ks[_ml_i], _ml_vs[_ml_i]); \
    _ml_m;                                                     \
})

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

// --- Result types ---

// Result types for (T, error):
typedef struct {
    bool val;
    so_Error err;
} so_R_bool_err;
typedef struct {
    double val;
    so_Error err;
} so_R_f64_err;
typedef struct {
    float val;
    so_Error err;
} so_R_f32_err;
typedef struct {
    int32_t val;
    so_Error err;
} so_R_i32_err;
typedef struct {
    int64_t val;
    so_Error err;
} so_R_i64_err;
typedef struct {
    so_byte val;
    so_Error err;
} so_R_byte_err;
typedef struct {
    so_int val;
    so_Error err;
} so_R_int_err;
typedef struct {
    so_rune val;
    so_Error err;
} so_R_rune_err;
typedef struct {
    so_Slice val;
    so_Error err;
} so_R_slice_err;
typedef struct {
    so_String val;
    so_Error err;
} so_R_str_err;
typedef struct {
    so_uint val;
    so_Error err;
} so_R_uint_err;
typedef struct {
    uint32_t val;
    so_Error err;
} so_R_u32_err;
typedef struct {
    uint64_t val;
    so_Error err;
} so_R_u64_err;
typedef struct {
    void* val;
    so_Error err;
} so_R_ptr_err;

// Result types for (T, T):
typedef struct {
    bool val;
    bool val2;
} so_R_bool_bool;
typedef struct {
    bool val;
    so_int val2;
} so_R_bool_int;
typedef struct {
    double val;
    bool val2;
} so_R_f64_bool;
typedef struct {
    double val;
    double val2;
} so_R_f64_f64;
typedef struct {
    double val;
    so_int val2;
} so_R_f64_int;
typedef struct {
    float val;
    bool val2;
} so_R_f32_bool;
typedef struct {
    int64_t val;
    int32_t val2;
} so_R_i64_i32;
typedef struct {
    so_int val;
    bool val2;
} so_R_int_bool;
typedef struct {
    so_int val;
    so_int val2;
} so_R_int_int;
typedef struct {
    so_int val;
    uint64_t val2;
} so_R_int_u64;
typedef struct {
    so_rune val;
    bool val2;
} so_R_rune_bool;
typedef struct {
    so_rune val;
    so_int val2;
} so_R_rune_int;
typedef struct {
    so_String val;
    bool val2;
} so_R_str_bool;
typedef struct {
    so_String val;
    so_String val2;
} so_R_str_str;
typedef struct {
    so_uint val;
    so_uint val2;
} so_R_uint_uint;
typedef struct {
    uint32_t val;
    bool val2;
} so_R_u32_bool;
typedef struct {
    uint32_t val;
    so_int val2;
} so_R_u32_int;
typedef struct {
    uint32_t val;
    uint32_t val2;
} so_R_u32_u32;
typedef struct {
    uint64_t val;
    bool val2;
} so_R_u64_bool;
typedef struct {
    uint64_t val;
    so_int val2;
} so_R_u64_int;
typedef struct {
    uint64_t val;
    uint64_t val2;
} so_R_u64_u64;

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

static inline void* unsafe_Add(void* ptr, size_t offset) {
    return (char*)ptr + offset;
}
static inline so_String unsafe_String(void* ptr, size_t len) {
    if (ptr == NULL) {
        return (so_String){(char*)&so_Nil, 0};
    }
    return (so_String){(char*)ptr, len};
}
static inline so_byte* unsafe_StringData(so_String s) {
    return (so_byte*)s.ptr;
}
static inline so_Slice unsafe_Slice(void* ptr, size_t len) {
    if (ptr == NULL) {
        return (so_Slice){&so_Nil, 0, 0};
    }
    return (so_Slice){ptr, len, len};
}
static inline void* unsafe_SliceData(so_Slice s) {
    return s.ptr;
}
