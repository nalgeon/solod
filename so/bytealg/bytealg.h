#include "so/builtin/builtin.h"

static inline so_int bytealg_Compare(so_Slice a, so_Slice b) {
    size_t n = a.len;
    if (b.len < n) n = b.len;
    int cmp = memcmp(a.ptr, b.ptr, n);
    if (cmp != 0) return cmp;
    if (a.len < b.len) return -1;
    if (a.len > b.len) return +1;
    return 0;
}

static inline so_int bytealg_IndexByte(so_Slice b, so_byte c) {
    void* at = memchr(b.ptr, (int)c, b.len);
    if (at == NULL) return -1;
    return (so_int)((char*)at - (char*)b.ptr);
}
