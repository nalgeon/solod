#include "so/builtin/builtin.h"

#if !defined(so_build_hosted)
// Pure C memchr implementation for freestanding environments.
static inline void* memchr(const void* s, int c, size_t n) {
    const unsigned char* p = s;
    unsigned char target = (unsigned char)c;
    while (n--) {
        if (*p == target) return (void*)p;
        p++;
    }
    return NULL;
}
#endif

// An empty slice can have a NULL pointer. memcmp and memchr reject a NULL
// pointer even for a zero length, so both functions below skip the call.
static inline so_int bytealg_Compare(so_Slice a, so_Slice b) {
    so_int n = a.len;
    if (b.len < n) n = b.len;
    int cmp = n > 0 ? memcmp(a.ptr, b.ptr, (size_t)n) : 0;
    if (cmp != 0) return cmp < 0 ? -1 : +1;
    if (a.len < b.len) return -1;
    if (a.len > b.len) return +1;
    return 0;
}

static inline so_int bytealg_IndexByte(so_Slice b, so_byte c) {
    if (b.len == 0) return -1;
    void* at = memchr(b.ptr, (int)c, (size_t)b.len);
    if (at == NULL) return -1;
    return (so_int)((char*)at - (char*)b.ptr);
}
