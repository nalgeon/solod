// Alloc allocates memory for a single value of type T using allocator a.
// Returns a pointer to the allocated memory or an error if allocation fails.
#define mem_Alloc(T, a) \
    (a).Alloc((a).self, sizeof(T), alignof(so_typeof(T)))

// Dealloc frees a value previously allocated with Alloc.
#define mem_Dealloc(T, a, ptr) \
    (a).Dealloc((a).self, (ptr), sizeof(T), alignof(so_typeof(T)))

// AllocSlice allocates a slice of type T with given length and capacity using allocator a.
// Returns a slice of the allocated memory or an error if allocation fails.
#define mem_AllocSlice(T, a, slen, scap) ({                                       \
    if ((slen) > (scap)) so_panic("mem: length exceeds capacity");                \
    so_Result _mem_res = (a).Alloc((a).self, sizeof(T) * (scap),                  \
                                   alignof(so_typeof(T)));                        \
    so_Slice _slice = {.ptr = _mem_res.val.as_ptr, .len = (slen), .cap = (scap)}; \
    so_Result _slice_res = {.val.as_slice = _slice, .err = _mem_res.err};         \
    _slice_res;                                                                   \
})

// DeallocSlice frees a slice previously allocated with AllocSlice.
#define mem_DeallocSlice(T, a, s) \
    (a).Dealloc((a).self, (s).ptr, sizeof(T) * (s).cap, alignof(so_typeof(T)))

// New allocates a single value of type T using the system allocator.
// Returns a pointer to the allocated memory or panics on failure.
#define mem_New(T) ({                              \
    so_Result _mem_res = mem_Alloc(T, mem_System); \
    if (_mem_res.err != NULL)                      \
        so_panic(_mem_res.err->msg);               \
    _mem_res.val.as_ptr;                           \
})

// Free frees a value previously allocated with New.
#define mem_Free(T, ptr) mem_Dealloc(T, mem_System, ptr)

// NewSlice allocates a slice of type T with given length
// and capacity using the system allocator.
// Returns a slice of the allocated memory or panics on failure.
#define mem_NewSlice(T, len, cap) ({                          \
    so_Result _res = mem_AllocSlice(T, mem_System, len, cap); \
    if (_res.err != NULL)                                     \
        so_panic(_res.err->msg);                              \
    _res.val.as_slice;                                        \
})

// FreeSlice frees a slice previously allocated with NewSlice.
#define mem_FreeSlice(T, s) mem_DeallocSlice(T, mem_System, s)
