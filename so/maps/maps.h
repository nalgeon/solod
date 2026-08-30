// wymum performs 128-bit multiply-and-mix.
// Uses hardware 128-bit multiply on 64-bit, software fallback on 32-bit.
#if so_int_bits == 64
static inline uint64_t maps_wymum(uint64_t a, uint64_t b) {
    __uint128_t r = (__uint128_t)a * b;
    return (uint64_t)(r >> 64) ^ (uint64_t)r;
}
#else
static inline uint64_t maps_wymum(uint64_t a, uint64_t b) {
    so_R_u64_u64 r = bits_Mul64(a, b);
    return r.val ^ r.val2;
}
#endif

// keyHash hashes a key, dispatching to string or inline hash.
#define maps_keyHash(K, key_ptr, seed) _Generic((K){}, \
    so_String: maps_hashString(key_ptr, seed),         \
    default: maps_hash(key_ptr, sizeof(K), seed))

// stringEqual compares two string keys through untyped pointers.
// Extracted from maps_keyEqual to avoid strict-aliasing warnings.
static inline bool maps_stringEqual(const void* a, const void* b) {
    return so_string_eq(*(const so_String*)a, *(const so_String*)b);
}

// equal compares two typed key pointers for equality.
#define maps_keyEqual(K, a, b)                  \
    _Generic((K){},                             \
        so_String: maps_stringEqual((a), (b)),  \
        default: memcmp((a), (b), sizeof(K)) == 0)
