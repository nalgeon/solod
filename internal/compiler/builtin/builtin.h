#pragma once

// The builtin is globally available in Clang and GCC. It's safe to define
// because any system header that defines alloca undefines it first.
#define alloca __builtin_alloca

#if __STDC_HOSTED__

#include <assert.h>
#include <inttypes.h>
#include <stdalign.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define so_build_hosted

#else

#include <stdbool.h>
#include <stdint.h>
#include <stdalign.h>
#include <stddef.h>

#define memcmp __builtin_memcmp
#define memcpy __builtin_memcpy
#define memmove __builtin_memmove
#define memset __builtin_memset

#define PRId64 "lld"
#define PRIu64 "llu"

#if defined(NDEBUG)
#define assert(cond) ((void)0)
#else
#define assert(cond)                   \
    do {                               \
        if (!(cond)) __builtin_trap(); \
    } while (0)
#endif

#endif  // __STDC_HOSTED__

// --- Build metadata ---

#if !defined(so_version)
#define so_version "(devel)"
#endif

#if defined(__APPLE__)
#define so_build_darwin
#define so_build_posix
#elif defined(__linux__)
#define so_build_linux
#define so_build_posix
#elif defined(__FreeBSD__)
#define so_build_freebsd
#define so_build_posix
#elif defined(__NetBSD__)
#define so_build_netbsd
#define so_build_posix
#elif defined(__OpenBSD__)
#define so_build_openbsd
#define so_build_posix
#elif defined(__DragonFly__)
#define so_build_dragonfly
#define so_build_posix
#elif defined(__wasm__)
#define so_build_wasm
#elif defined(_WIN32)
#define so_build_windows
#endif

#if defined(__x86_64__) || defined(_M_X64)
#define so_build_amd64
#elif defined(__i386__) || defined(_M_IX86)
#define so_build_i386
#elif defined(__aarch64__) || defined(_M_ARM64)
#define so_build_arm64
#elif defined(__riscv) && __riscv_xlen == 64
#define so_build_riscv64
#elif defined(__wasm32__)
#define so_build_wasm32
#endif

// --- General utilities ---

#define SO_CONCAT(a, b) a##b
#define SO_NAME(a, b) SO_CONCAT(a, b)

#define so_auto __auto_type
#define so_typeof __typeof__
#define so_unreachable() __builtin_unreachable()
#define so_unused __attribute__((unused))

typedef uint8_t so_byte;
typedef int32_t so_rune;

typedef struct {
    uint64_t hi;
    uint64_t lo;
} so_uint128;

#if SIZE_MAX == 0xFFFFFFFFu
#define so_int_bits 32
#define so_max_int INT32_MAX
#define PRIdINT "d"
#define PRIuINT "u"
typedef int32_t so_int;
typedef uint32_t so_uint;
#else
#define so_int_bits 64
#define so_max_int INT64_MAX
#define PRIdINT PRId64
#define PRIuINT PRIu64
typedef int64_t so_int;
typedef uint64_t so_uint;
#endif

// --- Panic and assert ---

#if !defined(so_build_hosted)

static inline size_t strlen(const char* s) {
    const char* p = s;
    while (*p) p++;
    return p - s;
}

// so_write_out is the target's output hook for a freestanding environment.
// The default definition in builtin.c drops the bytes and reports a full write.
so_int so_write_out(const uint8_t* buf, so_int size);

// so_write_str writes a null-terminated string through so_write_out.
static inline void so_write_str(const char* s) {
    so_write_out((const uint8_t*)s, (so_int)strlen(s));
}

// so_stringify expands x and makes a string literal of the result.
#define so_stringify_(x) #x
#define so_stringify(x) so_stringify_(x)

#endif  // !so_build_hosted

// panic aborts the program with the given message.
//
// SO_PANIC_MODE selects how a hosted build terminates after printing
// the message. The `so` toolchain defaults to SO_PANIC_TRACE (and adds
// the -rdynamic it needs); compiling this C by hand without -D falls
// back to SO_PANIC_EXIT, which needs no extra flags.
//   - SO_PANIC_EXIT:  exit(1). Clean, deterministic exit code.
//   - SO_PANIC_ABORT: abort(). Raises SIGABRT for a core dump or debugger.
//   - SO_PANIC_TRACE: print a backtrace, then exit(1).
// A freestanding build ignores the mode and always traps.
#define SO_PANIC_EXIT 0
#define SO_PANIC_ABORT 1
#define SO_PANIC_TRACE 2
#if !defined(SO_PANIC_MODE)
#define SO_PANIC_MODE SO_PANIC_EXIT
#endif

#if defined(so_build_hosted)

#if SO_PANIC_MODE == SO_PANIC_TRACE
// print_trace writes a symbolized backtrace of the current call stack
// to stderr. Enabled only in trace panic mode.
void so_print_trace(void);
#define so_panic_tail()   \
    do {                  \
        so_print_trace(); \
        exit(1);          \
    } while (0)
#elif SO_PANIC_MODE == SO_PANIC_ABORT
#define so_panic_tail() abort()
#else
#define so_panic_tail() exit(1)
#endif

#define so_panic(msg)                                     \
    do {                                                  \
        fprintf(stderr, "panic: %s\n  %s:%d (func %s)\n", \
                msg, __FILE__, __LINE__, __func__);       \
        so_panic_tail();                                  \
    } while (0)

#else

#define so_panic(msg)                                                       \
    do {                                                                    \
        so_write_str("panic: ");                                            \
        so_write_str(msg);                                                  \
        so_write_str("\n  " __FILE__ ":" so_stringify(__LINE__) " (func "); \
        so_write_str(__func__);                                             \
        so_write_str(")\n");                                                \
        __builtin_trap();                                                   \
    } while (0)

#endif  // so_build_hosted

// assert panics with the given message if the condition is false.
// SO_NO_ASSERT removes the check entirely, so cond must be free of side
// effects. NDEBUG has no effect here.
#if defined(SO_NO_ASSERT)
#define so_assert(cond, msg) ((void)0)
#else
#define so_assert(cond, msg) \
    do {                     \
        if (!(cond)) {       \
            so_panic(msg);   \
        }                    \
    } while (0)
#endif

// assume tells the C compiler that cond is always true. It generates no code
// in any build, and SO_NO_ASSERT has no effect on it. The behavior is
// undefined if cond is false.
//
// Use assume only for conditions that are provably true, such as when
// a pointer is known to be non-null but the compiler cannot see it.
// Use assert for all other conditions.
#define so_assume(cond)       \
    do {                      \
        if (!(cond)) {        \
            so_unreachable(); \
        }                     \
    } while (0)

// --- Alloca safety ---

// MaxAllocaSize is the maximum size that can be
// allocated with alloca (64 KB by default).
#if !defined(SO_MAX_ALLOCA_SIZE)
#define SO_MAX_ALLOCA_SIZE (64 << 10)  // in bytes
#endif

#define so_alloca(size) ({                                \
    size_t _size = (size_t)(size);                        \
    if (_size > SO_MAX_ALLOCA_SIZE)                       \
        so_panic("alloca: size exceeds maximum allowed"); \
    _size ? alloca(_size) : NULL;                         \
})

// --- Integer division ---

// is_signed reports whether the type of x is signed.
#define so_is_signed(x) ((so_typeof(x)) - 1 < 0)

// negate returns the two's complement negation of a signed x. Can't use -x
// because the negation of the minimum value overflows, and signed overflow
// is undefined.
#define so_negate(x) ((so_typeof(x))(0 - (uintmax_t)(x)))

// div and mod evaluate a/b and a%b, with a guard for the two divisors that
// C leaves undefined: (1) zero divisor panics, (2) a divisor of -1 overflows
// for the minimum value.
#define so_div(a, b) ({                                       \
    so_typeof(a) _div_a = (a);                                \
    so_typeof(b) _div_b = (b);                                \
    so_assert(_div_b != 0, "integer divide by zero");         \
    so_is_signed(_div_b) && _div_b == (so_typeof(_div_b)) - 1 \
        ? so_negate(_div_a)                                   \
        : _div_a / _div_b;                                    \
})

#define so_mod(a, b) ({                                       \
    so_typeof(a) _div_a = (a);                                \
    so_typeof(b) _div_b = (b);                                \
    so_assert(_div_b != 0, "integer divide by zero");         \
    so_is_signed(_div_b) && _div_b == (so_typeof(_div_b)) - 1 \
        ? (so_typeof(_div_a))0                                \
        : _div_a % _div_b;                                    \
})

// --- Comparison ---

// mem_eq returns true if two memory blocks are equal.
static inline bool so_mem_eq(const void* a, const void* b, size_t size) {
    return memcmp(a, b, size) == 0;
}

// mem_ne returns true if two memory blocks are not equal.
static inline bool so_mem_ne(const void* a, const void* b, size_t size) {
    return memcmp(a, b, size) != 0;
}

// --- String type ---

// String is a pointer to array of bytes plus a length.
typedef struct {
    const char* ptr;
    so_int len;
} so_String;

// strlit creates a String from a string literal.
#define so_str(s) ((so_String){s, (so_int)(sizeof(s) - 1)})

// cstr returns a null-terminated C string copy on the stack.
// alloca returns NULL for a zero size, but the size here is len + 1,
// so the buffer is never NULL. The assume signals this to the compiler.
#define so_cstr(s) ({                                     \
    so_String _s = (s);                                   \
    char* _buf = so_alloca(_s.len + 1);                   \
    so_assume(_buf != NULL);                              \
    if (_s.len > 0) memcpy(_buf, _s.ptr, (size_t)_s.len); \
    _buf[_s.len] = '\0';                                  \
    _buf;                                                 \
})

// string_slice creates a substring [from, to).
#define so_string_slice(s, from, to) ({                    \
    so_String _s = (s);                                    \
    so_int _from = (so_int)(from);                         \
    so_int _to = (so_int)(to);                             \
    so_assert(0 <= _from && _from <= _to && _to <= _s.len, \
              "slice bounds out of range");                \
    (so_String){_s.ptr + _from, _to - _from};              \
})

// string_add concatenates two strings.
// Allocates memory on the stack until the calling function returns.
#define so_string_add(s1, s2) ({                                       \
    so_String _s1 = (s1);                                              \
    so_String _s2 = (s2);                                              \
    so_int _total = _s1.len + _s2.len;                                 \
    char* _buf = so_alloca(_total);                                    \
    so_assume(_buf != NULL || _total == 0);                            \
    if (_s1.len > 0) memcpy(_buf, _s1.ptr, (size_t)_s1.len);           \
    if (_s2.len > 0) memcpy(_buf + _s1.len, _s2.ptr, (size_t)_s2.len); \
    (so_String){_buf, _total};                                         \
})

// string_eq returns true if two strings are equal.
static inline bool so_string_eq(so_String s1, so_String s2) {
    return s1.len == s2.len &&
           (s1.len == 0 || memcmp(s1.ptr, s2.ptr, (size_t)s1.len) == 0);
}

// string_ne returns true if two strings are not equal.
static inline bool so_string_ne(so_String s1, so_String s2) {
    return !so_string_eq(s1, s2);
}

// string_lt returns true if s1 < s2 in lexicographical order.
static inline bool so_string_lt(so_String s1, so_String s2) {
    so_int n = s1.len < s2.len ? s1.len : s2.len;
    int cmp = n > 0 ? memcmp(s1.ptr, s2.ptr, (size_t)n) : 0;
    return cmp < 0 || (cmp == 0 && s1.len < s2.len);
}

// string_lte returns true if s1 <= s2 in lexicographical order.
static inline bool so_string_lte(so_String s1, so_String s2) {
    return so_string_lt(s1, s2) || so_string_eq(s1, s2);
}

// string_gt returns true if s1 > s2 in lexicographical order.
static inline bool so_string_gt(so_String s1, so_String s2) {
    so_int n = s1.len < s2.len ? s1.len : s2.len;
    int cmp = n > 0 ? memcmp(s1.ptr, s2.ptr, (size_t)n) : 0;
    return cmp > 0 || (cmp == 0 && s1.len > s2.len);
}

// string_gte returns true if s1 >= s2 in lexicographical order.
static inline bool so_string_gte(so_String s1, so_String s2) {
    return so_string_gt(s1, s2) || so_string_eq(s1, s2);
}

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
    so_int _n = so_utf8_encode((so_rune)(r), _buf); \
    (so_String){_buf, _n};                          \
})

// utf8_encode encodes a single rune into buf (up to 4 bytes).
// Encodes 0xFFFD for a rune that UTF-8 cannot represent: a negative rune,
// a surrogate half, or a rune above 0x10FFFF.
// Returns the number of bytes written.
so_int so_utf8_encode(so_rune r, char* buf);

// utf8_decode decodes one UTF-8 rune from string s at byte offset i.
// Stores the byte width in *w.
// Returns the decoded rune, or 0xFFFD for invalid UTF-8.
so_rune so_utf8_decode(so_String s, so_int i, so_int* w);

// --- Arrays ---

// slice_array converts a slice to an array pointer with bounds checking.
#define so_slice_array(s, n) ({                                  \
    so_Slice _s = (s);                                           \
    so_assert(_s.len >= (so_int)(n), "slice-to-array mismatch"); \
    _s.ptr;                                                      \
})

// array_slice creates a slice from a C array.
// 'size' is the total array size (known at compile time).
#define so_array_slice(T, arr, from, to, size) ({                \
    so_int _from = (so_int)(from);                               \
    so_int _to = (so_int)(to);                                   \
    so_int _size = (so_int)(size);                               \
    ((so_Slice){(T*)(arr) + _from, _to - _from, _size - _from}); \
})

// array_slice3 creates a slice from a C array with an explicit capacity.
#define so_array_slice3(T, arr, from, to, max) ({               \
    so_int _from = (so_int)(from);                              \
    so_int _to = (so_int)(to);                                  \
    so_int _max = (so_int)(max);                                \
    ((so_Slice){(T*)(arr) + _from, _to - _from, _max - _from}); \
})

// --- Slice type ---

// Slice is a pointer to array of elements plus a length.
// The slice is nil if and only if its capacity is 0.
typedef struct {
    void* ptr;
    so_int len;
    so_int cap;
} so_Slice;

// make_slice creates a zero-initialized slice on the stack.
// Allocates memory on the stack until the calling function returns.
//
// n is 0 for an element type with size 0 (struct{}). The macro still
// allocates one byte because so_at_ptr requires a non-NULL pointer for
// any slice with a capacity greater than 0.
#define so_make_slice(T, len, cap) ({                \
    so_int _cap = (so_int)(cap);                     \
    size_t _n = sizeof(T) * (size_t)_cap;            \
    void* _p = _cap ? so_alloca(_n ? _n : 1) : NULL; \
    if (_n) memset(_p, 0, _n);                       \
    (so_Slice){_p, (len), _cap};                     \
})

// slice creates a slice from another slice
// from index 'from' (inclusive) to index 'to' (exclusive).
// A NULL source pointer means a zero capacity, so the bounds check
// forces a zero length. The assume signals this to the compiler.
#define so_slice(T, s, from, to) ({                        \
    so_Slice _s = (s);                                     \
    so_int _from = (so_int)(from);                         \
    so_int _to = (so_int)(to);                             \
    so_assert(0 <= _from && _from <= _to && _to <= _s.cap, \
              "slice bounds out of range");                \
    T* _ptr = _s.ptr == NULL ? NULL : (T*)_s.ptr + _from;  \
    so_int _len = _to - _from;                             \
    so_assume(_ptr != NULL || _len == 0);                  \
    (so_Slice){_ptr, _len, _s.cap - _from};                \
})

// slice3 creates a slice from another slice with an explicit capacity.
#define so_slice3(T, s, from, to, max) ({                      \
    so_Slice _s = (s);                                         \
    so_int _from = (so_int)(from);                             \
    so_int _to = (so_int)(to);                                 \
    so_int _max = (so_int)(max);                               \
    so_assert(0 <= _from && _from <= _to &&                    \
                  _to <= _max && _max <= _s.cap,               \
              "slice bounds out of range");                    \
    (so_Slice){(T*)_s.ptr + _from, _to - _from, _max - _from}; \
})

// decay extracts the pointer from a slice for passing to C functions.
// Returns NULL for empty/nil slices.
#define so_decay(s) ({ so_Slice _s = (s); _s.cap ? _s.ptr : NULL; })

// append appends n elements to a slice without resizing.
// Returns the new slice with updated length.
// Panics if the new length exceeds the capacity.
//
// The generator passes n. The macro cannot calculate n with sizeof(T),
// because sizeof(T) is 0 for a struct type with no fields.
//
// sizeof(_vals) is zero for an append of no values, and for an element
// type with no fields. Both cases skip memcpy to avoid UB.
#define so_append(T, s, n, ...) ({                         \
    so_Slice _s = (s);                                     \
    T _vals[] = {__VA_ARGS__};                             \
    so_int _n = (so_int)(n);                               \
    if (_s.len + _n > _s.cap)                              \
        so_panic("append: out of capacity");               \
    if (sizeof(_vals) > 0)                                 \
        memcpy((T*)_s.ptr + _s.len, _vals, sizeof(_vals)); \
    _s.len += _n;                                          \
    _s;                                                    \
})

// extend appends all elements from a source slice to a destination slice.
// Returns the new slice with updated length.
// Panics if the new length exceeds the capacity.
#define so_extend(T, dst, src) ({                       \
    so_Slice _dst = (dst);                              \
    so_Slice _src = (src);                              \
    if (_dst.len + _src.len > _dst.cap)                 \
        so_panic("extend: out of capacity");            \
    if (_src.len > 0)                                   \
        memcpy((T*)_dst.ptr + _dst.len,                 \
               _src.ptr, (size_t)_src.len * sizeof(T)); \
    _dst.len += _src.len;                               \
    _dst;                                               \
})

// copy copies elements from src to dst. Returns the number of elements copied
// (which is the minimum of dst.len and src.len).
#define so_copy(T, dst, src) so_copy_impl(dst, src, sizeof(T))
static inline so_int so_copy_impl(so_Slice dst, so_Slice src, size_t elem_size) {
    so_int _n = dst.len < src.len ? dst.len : src.len;
    if (_n > 0) memmove(dst.ptr, src.ptr, (size_t)(_n)*elem_size);
    return _n;
}

// clear sets all elements up to the length
// of the slice to their zero value.
#define so_clear(T, s) ({                                \
    so_Slice _s = (s);                                   \
    if (_s.len > 0) {                                    \
        memset(_s.ptr, 0, (size_t)(_s.len) * sizeof(T)); \
    }                                                    \
    _s;                                                  \
})

// --- String/slice operations ---

// at returns a reference to the element at index i in a slice or string.
#define so_at(T, s, i) (*so_at_ptr(T, s, i))
#define so_at_ptr(T, s, i) ({                   \
    so_auto _s_at = (s);                        \
    so_int _i = (so_int)(i);                    \
    so_assert((so_uint)_i < (so_uint)_s_at.len, \
              "index out of range");            \
    if (_s_at.ptr == NULL) so_unreachable();    \
    (T*)_s_at.ptr + _i;                         \
})

// len returns the length of a slice or string.
#define so_len(s) ((s).len)

// cap returns the capacity of a slice.
#define so_cap(s) ((s).cap)

// string_bytes reinterprets a string as a byte slice (zero-copy).
#define so_string_bytes(s) ({                  \
    so_String _s = (s);                        \
    (so_Slice){(void*)_s.ptr, _s.len, _s.len}; \
})

// string_runes decodes a string's UTF-8 bytes into a rune slice.
// Allocates memory on the stack until the calling function returns.
#define so_string_runes(s) ({                                      \
    so_String _s = (s);                                            \
    so_rune* _buf = so_alloca((size_t)(_s.len) * sizeof(so_rune)); \
    so_string_runes_impl(_s, _buf);                                \
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

// copy_string copies bytes from a string to a byte slice. Returns the number
// of bytes copied (which is the minimum of dst.len and src.len).
static inline so_int so_copy_string(so_Slice dst, so_String src) {
    so_int _n = dst.len < src.len ? dst.len : src.len;
    if (_n > 0) memmove(dst.ptr, src.ptr, (size_t)_n);
    return _n;
}

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
typedef struct {
    void* self;
    so_String (*Error)(void* self);
} so_Error;

// errors_New creates a new error with the given message string.
// so_Error errors_New(const char* s)
#define errors_New(s) ((so_Error){.self = s, .Error = so_error_error})

// so_error_error implements the Error method for errors
// created with errors_New (.self is a C string pointer).
// Returns the error message as a string, or "<nil>" for nil errors.
static inline so_String so_error_error(void* self) {
    if (!self) return so_str("<nil>");
    return (so_String){self, (so_int)strlen(self)};
}

// so_error_cstr returns the error message as a C string.
// An error with a non-null self always carries a method pointer.
// The assume signals this to the compiler.
#define so_error_cstr(err) ({                          \
    so_Error _error = (err);                           \
    const char* _err_str;                              \
    if (!_error.self) {                                \
        _err_str = "<nil>";                            \
    } else if (_error.Error == so_error_error) {       \
        _err_str = (const char*)_error.self;           \
    } else {                                           \
        so_assume(_error.Error != NULL);               \
        _err_str = so_cstr(_error.Error(_error.self)); \
    }                                                  \
    _err_str;                                          \
})

// --- Result types ---

// clang-format off

// Result types for (T, error):
typedef struct { bool val; so_Error err; } so_R_bool_err;
typedef struct { float val; so_Error err; } so_R_f32_err;
typedef struct { double val; so_Error err; } so_R_f64_err;
typedef struct { so_int val; so_Error err; } so_R_int_err;
typedef struct { int8_t val; so_Error err; } so_R_i8_err;
typedef struct { int16_t val; so_Error err; } so_R_i16_err;
typedef struct { int32_t val; so_Error err; } so_R_i32_err;
typedef struct { int64_t val; so_Error err; } so_R_i64_err;
typedef struct { so_uint val; so_Error err; } so_R_uint_err;
typedef struct { uint16_t val; so_Error err; } so_R_u16_err;
typedef struct { uint32_t val; so_Error err; } so_R_u32_err;
typedef struct { uint64_t val; so_Error err; } so_R_u64_err;
typedef struct { so_byte val; so_Error err; } so_R_byte_err;
typedef struct { so_rune val; so_Error err; } so_R_rune_err;
typedef struct { so_String val; so_Error err; } so_R_str_err;
typedef struct { so_Slice val; so_Error err; } so_R_slice_err;
typedef struct { void* val; so_Error err; } so_R_ptr_err;

// Result types for (T, T):
typedef struct { bool val; bool val2; } so_R_bool_bool;
typedef struct { bool val; so_int val2; } so_R_bool_int;
typedef struct { double val; bool val2; } so_R_f64_bool;
typedef struct { double val; double val2; } so_R_f64_f64;
typedef struct { double val; so_int val2; } so_R_f64_int;
typedef struct { float val; bool val2; } so_R_f32_bool;
typedef struct { float val; float val2; } so_R_f32_f32;
typedef struct { int64_t val; int32_t val2; } so_R_i64_i32;
typedef struct { so_byte val; so_int val2; } so_R_byte_int;
typedef struct { so_int val; bool val2; } so_R_int_bool;
typedef struct { so_int val; so_int val2; } so_R_int_int;
typedef struct { so_int val; uint64_t val2; } so_R_int_u64;
typedef struct { so_rune val; bool val2; } so_R_rune_bool;
typedef struct { so_rune val; so_int val2; } so_R_rune_int;
typedef struct { so_String val; bool val2; } so_R_str_bool;
typedef struct { so_String val; so_String val2; } so_R_str_str;
typedef struct { so_uint val; so_uint val2; } so_R_uint_uint;
typedef struct { uint32_t val; bool val2; } so_R_u32_bool;
typedef struct { uint32_t val; so_int val2; } so_R_u32_int;
typedef struct { uint32_t val; uint32_t val2; } so_R_u32_u32;
typedef struct { uint64_t val; bool val2; } so_R_u64_bool;
typedef struct { uint64_t val; so_int val2; } so_R_u64_int;
typedef struct { uint64_t val; uint64_t val2; } so_R_u64_u64;
typedef struct { void* val; bool val2; } so_R_ptr_bool;
typedef struct { void* val; so_int val2; } so_R_ptr_int;

// clang-format on

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
static inline so_String unsafe_String(void* ptr, so_int len) {
    if (ptr == NULL) {
        return (so_String){};
    }
    return (so_String){(char*)ptr, len};
}
static inline so_byte* unsafe_StringData(so_String s) {
    if (s.len == 0) {
        return NULL;
    }
    return (so_byte*)s.ptr;
}
static inline so_Slice unsafe_Slice(void* ptr, so_int len) {
    if (ptr == NULL) {
        return (so_Slice){};
    }
    return (so_Slice){ptr, len, len};
}
static inline void* unsafe_SliceData(so_Slice s) {
    if (s.cap == 0) {
        // A slice length is never above the capacity.
        so_assume(s.len == 0);
        return NULL;
    }
    return s.ptr;
}

// --- Map type ---

// Map is an open-addressed hash table with MSI (mask-step-index) probing.
typedef struct {
    void* keys;
    void* vals;
    uint8_t* used;  // 0=empty, 1=occupied
    so_int len;
    so_int cap;  // always power of 2
} so_Map;

// key_hash hashes a map key to a 64-bit value (FNV-1a).
// The seed is the map's own address (randomized by ASLR).
static inline uint64_t so_key_hash_def(const void* ptr, so_int n, uint64_t seed) {
    const uint8_t* p = (const uint8_t*)ptr;
    uint64_t h = seed;
    for (so_int i = 0; i < n; i++) {
        h ^= p[i];
        h *= 0x100000001b3ULL;
    }
    return h;
}
static inline uint64_t so_key_hash_str(const void* ptr, so_int n, uint64_t seed) {
    (void)n;
    const so_String* s = (const so_String*)ptr;
    return so_key_hash_def(s->ptr, s->len, seed);
}

#define so_key_hash(key) \
    _Generic((key), so_String: so_key_hash_str, default: so_key_hash_def)

// key_eq compares two map keys for equality.
// Uses so_string_eq for strings, memcmp for everything else.
static inline bool so_key_eq_def(const void* a, const void* b, size_t n) {
    return memcmp(a, b, n) == 0;
}
static inline bool so_key_eq_str(const void* a, const void* b, size_t n) {
    (void)n;
    return so_string_eq(*(const so_String*)a, *(const so_String*)b);
}

#define so_key_eq(key) \
    _Generic((key), so_String: so_key_eq_str, default: so_key_eq_def)

// map_nextpow2 rounds up to the next power of 2.
so_int so_map_nextpow2(so_int n);

// map_find looks up a key in the map.
void so_map_find(const so_Map* m, const void* key, size_t key_size,
                 void* out_val, size_t val_size,
                 uint64_t hash, bool* found,
                 bool (*eq)(const void*, const void*, size_t));

// map_set_impl inserts or updates a key-value pair in the map.
void so_map_set_impl(so_Map* m, const void* key, size_t key_size,
                     const void* val, size_t val_size,
                     uint64_t hash,
                     bool (*eq)(const void*, const void*, size_t));

// map_cap computes the internal capacity for n elements (keeps load <= 75%).
static inline so_int so_map_cap(so_int n) {
    if (n == 0) return 0;
    return so_map_nextpow2(n + n / 3 + 1);
}

// make_map creates a zero-initialized map on the stack.
#define so_make_map(K, V, n) ({                   \
    so_int _n = (n);                              \
    so_assert(_n != 0, "map: zero capacity");     \
    so_int _cap = so_map_cap(_n);                 \
    size_t _ksz = sizeof(K) * (size_t)_cap;       \
    size_t _vsz = sizeof(V) * (size_t)_cap;       \
    size_t _usz = sizeof(uint8_t) * (size_t)_cap; \
    void* _kp = so_alloca(_ksz ? _ksz : 1);       \
    void* _vp = so_alloca(_vsz ? _vsz : 1);       \
    uint8_t* _up = so_alloca(_usz);               \
    if (_kp) memset(_kp, 0, _ksz);                \
    if (_vp) memset(_vp, 0, _vsz);                \
    if (_up) memset(_up, 0, _usz);                \
    so_Map* _mp = so_alloca(sizeof(so_Map));      \
    *_mp = (so_Map){_kp, _vp, _up, 0, _cap};      \
    _mp;                                          \
})

// map_set inserts or updates a key-value pair in the map.
#define so_map_set(K, V, m, key, val)                            \
    do {                                                         \
        so_Map* _m = (m);                                        \
        K _k = (key);                                            \
        V _v = (val);                                            \
        uint64_t _seed = so_map_seed(_m);                        \
        uint64_t _hash = so_key_hash(_k)(&_k, sizeof(K), _seed); \
        so_map_set_impl(_m, &_k, sizeof(K), &_v, sizeof(V),      \
                        _hash, so_key_eq(_k));                   \
    } while (0)

// map_get returns the value for the given key, or zero if not found.
#define so_map_get(K, V, m, key) ({                          \
    const so_Map* _m = (m);                                  \
    K _k = (key);                                            \
    V _v = {};                                               \
    bool _found = false;                                     \
    uint64_t _seed = so_map_seed(_m);                        \
    uint64_t _hash = so_key_hash(_k)(&_k, sizeof(K), _seed); \
    so_map_find(_m, &_k, sizeof(K), &_v, sizeof(V),          \
                _hash, &_found, so_key_eq(_k));              \
    _v;                                                      \
})

// map_has returns true if the map contains the given key.
#define so_map_has(K, m, key) ({                             \
    const so_Map* _m = (m);                                  \
    K _k = (key);                                            \
    bool _found = false;                                     \
    uint64_t _seed = so_map_seed(_m);                        \
    uint64_t _hash = so_key_hash(_k)(&_k, sizeof(K), _seed); \
    so_map_find(_m, &_k, sizeof(K), NULL, 0,                 \
                _hash, &_found, so_key_eq(_k));              \
    _found;                                                  \
})

// map_lit creates a map from literal key/value arrays.
#define so_map_lit(K, V, n, keys, vals) ({                     \
    so_int _ml_n = (n);                                        \
    so_Map* _ml_m = so_make_map(K, V, _ml_n);                  \
    K* _ml_ks = (keys);                                        \
    V* _ml_vs = (vals);                                        \
    for (so_int _ml_i = 0; _ml_i < _ml_n; _ml_i++)             \
        so_map_set(K, V, _ml_m, _ml_ks[_ml_i], _ml_vs[_ml_i]); \
    _ml_m;                                                     \
})

#define so_map_seed(m) ((uint64_t)(uintptr_t)(m))

// --- OS intrinsics ---

// Command-line arguments, populated by main().
extern so_Slice os_Args;

#if defined(so_build_hosted)
// so_args_init populates os_Args from C argc/argv.
// buf must be a so_String array of at least argc elements (VLA on main's stack).
static inline void so_args_init(int argc, char* argv[], so_String* buf) {
    for (int i = 0; i < argc; i++) {
        buf[i] = (so_String){argv[i], (so_int)strlen(argv[i])};
    }
    os_Args = (so_Slice){buf, (so_int)argc, (so_int)argc};
}
#endif
