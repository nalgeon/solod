// Append appends elements to a heap-allocated slice, growing it if needed.
// Returns the updated slice or panics on allocation failure.
// If the allocator is nil, uses the system allocator.
#define slices_Append(T, a, s, ...) ({                               \
    T _vals[] = {__VA_ARGS__};                                       \
    so_int _n = (so_int)(sizeof(_vals) / sizeof(T));                 \
    slices_extend((a), (s),                                          \
                  (so_Slice){(so_byte*)_vals, _n, _n},               \
                  (so_int)sizeof(T), (so_int)alignof(so_typeof(T))); \
})

// Extend appends all elements from another slice, growing if needed.
// Returns the updated slice or panics on allocation failure.
// If the allocator is nil, uses the system allocator.
#define slices_Extend(T, a, s, other)                     \
    slices_extend((a), (s), (other), (so_int)(sizeof(T)), \
                  (so_int)(alignof(so_typeof(T))))

// Header returns the Slice header for a given slice.
#define slices_Header(T, s) (s)
