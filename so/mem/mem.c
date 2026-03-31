//go:build ignore
#include "mem.h"

so_R_slice_err mem_tryAllocSlice(const struct mem_Allocator* a, size_t elemSize, size_t align, so_int len, so_int cap) {
    if (len < 0) so_panic("mem: negative length");
    if (cap <= 0) so_panic("mem: invalid capacity");
    if (len > cap) so_panic("mem: length exceeds capacity");
    if (INT64_MAX / (so_int)elemSize < cap) so_panic("mem: capacity overflow");
    if (!a->self) a = &mem_System;

    so_R_ptr_err res = a->Alloc(a->self, elemSize * cap, align);
    if (res.err != NULL) return (so_R_slice_err){.err = res.err};
    so_Slice s = {.ptr = res.val, .len = len, .cap = cap};
    return (so_R_slice_err){.val = s};
}

so_R_slice_err mem_tryReallocSlice(const struct mem_Allocator* a, so_Slice s, so_int newLen, so_int newCap, size_t elemSize, size_t align) {
    if (newLen < 0) so_panic("mem: negative length");
    if (newCap <= 0) so_panic("mem: invalid capacity");
    if (newLen > newCap) so_panic("mem: length exceeds capacity");
    if (INT64_MAX / (so_int)elemSize < newCap) so_panic("mem: capacity overflow");
    if (!a->self) a = &mem_System;

    so_R_ptr_err res;
    if (s.cap == 0) {
        res = a->Alloc(a->self, elemSize * newCap, align);
    } else {
        res = a->Realloc(a->self, s.ptr,
                         elemSize * s.cap,
                         elemSize * newCap, align);
    }

    if (res.err != NULL) return (so_R_slice_err){.err = res.err};
    so_Slice ns = {.ptr = res.val, .len = newLen, .cap = newCap};
    return (so_R_slice_err){.val = ns};
}
