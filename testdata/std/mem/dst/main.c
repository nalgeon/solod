#include "main.h"

// -- Forward declarations (functions and methods) --
static void withDefer(void);

// -- Implementation --

static void withDefer(void) {
    main_Point* p = mem_Alloc(main_Point, (mem_Allocator){0});
    p->x = 11;
    p->y = 22;
    if (p->x != 11 || p->y != 22) {
        mem_Free(main_Point, (mem_Allocator){0}, p);
        so_panic("unexpected value");
    }
    mem_Free(main_Point, (mem_Allocator){0}, p);
}

int main(void) {
    {
        // TryAlloc and Free.
        so_Result _res1 = mem_TryAlloc(main_Point, mem_System);
        main_Point* p = _res1.val.as_ptr;
        so_Error err = _res1.err;
        if (err != NULL) {
            so_panic("Alloc: allocation failed");
        }
        p->x = 11;
        p->y = 22;
        if (p->x != 11 || p->y != 22) {
            so_panic("Alloc: unexpected value");
        }
        mem_Free(main_Point, mem_System, p);
    }
    {
        // TryAllocSlice and FreeSlice.
        so_Result _res2 = mem_TryAllocSlice(so_int, mem_System, 3, 3);
        so_Slice slice = _res2.val.as_slice;
        so_Error err = _res2.err;
        if (err != NULL) {
            so_panic("AllocSlice: allocation failed");
        }
        so_at(so_int, slice, 0) = 11;
        so_at(so_int, slice, 1) = 22;
        so_at(so_int, slice, 2) = 33;
        if (so_at(so_int, slice, 0) != 11 || so_at(so_int, slice, 1) != 22 || so_at(so_int, slice, 2) != 33) {
            so_panic("AllocSlice: unexpected value");
        }
        mem_FreeSlice(so_int, mem_System, slice);
    }
    {
        // Alloc/Free with default allocator.
        main_Point* p = mem_Alloc(main_Point, (mem_Allocator){0});
        p->x = 11;
        p->y = 22;
        if (p->x != 11 || p->y != 22) {
            so_panic("New: unexpected value");
        }
        mem_Free(main_Point, (mem_Allocator){0}, p);
    }
    {
        // AllocSlice/FreeSlice with default allocator.
        so_Slice slice = mem_AllocSlice(so_int, (mem_Allocator){0}, 3, 3);
        so_at(so_int, slice, 0) = 11;
        so_at(so_int, slice, 1) = 22;
        so_at(so_int, slice, 2) = 33;
        if (so_at(so_int, slice, 0) != 11 || so_at(so_int, slice, 1) != 22 || so_at(so_int, slice, 2) != 33) {
            so_panic("NewSlice: unexpected value");
        }
        mem_FreeSlice(so_int, (mem_Allocator){0}, slice);
    }
    withDefer();
    {
        // Append within capacity.
        so_Slice s = mem_AllocSlice(so_int, (mem_Allocator){0}, 0, 8);
        s = mem_Append(so_int, (mem_Allocator){0}, s, 10, 20, 30);
        if (so_len(s) != 3 || so_at(so_int, s, 0) != 10 || so_at(so_int, s, 1) != 20 || so_at(so_int, s, 2) != 30) {
            so_panic("Append: unexpected value");
        }
        mem_FreeSlice(so_int, (mem_Allocator){0}, s);
    }
    {
        // Append that triggers growth.
        so_Slice s = mem_AllocSlice(so_int, (mem_Allocator){0}, 0, 2);
        s = mem_Append(so_int, (mem_Allocator){0}, s, 1, 2);
        s = mem_Append(so_int, (mem_Allocator){0}, s, 3, 4, 5);
        if (so_len(s) != 5 || so_at(so_int, s, 0) != 1 || so_at(so_int, s, 4) != 5) {
            so_panic("Append grow: unexpected value");
        }
        mem_FreeSlice(so_int, (mem_Allocator){0}, s);
    }
    {
        // Extend from another slice.
        so_Slice s = mem_AllocSlice(so_int, (mem_Allocator){0}, 0, 8);
        so_Slice other = (so_Slice){(so_int[3]){100, 200, 300}, 3, 3};
        s = mem_Extend(so_int, (mem_Allocator){0}, s, other);
        if (so_len(s) != 3 || so_at(so_int, s, 0) != 100 || so_at(so_int, s, 2) != 300) {
            so_panic("Extend: unexpected value");
        }
        mem_FreeSlice(so_int, (mem_Allocator){0}, s);
    }
    {
        // TryAppend success.
        so_Slice s = mem_AllocSlice(so_int, (mem_Allocator){0}, 0, 4);
        so_Result _res3 = mem_TryAppend(so_int, (mem_Allocator){0}, s, 42);
        s = _res3.val.as_slice;
        so_Error err = _res3.err;
        if (err != NULL) {
            so_panic("TryAppend: unexpected error");
        }
        if (so_len(s) != 1 || so_at(so_int, s, 0) != 42) {
            so_panic("TryAppend: unexpected value");
        }
        mem_FreeSlice(so_int, (mem_Allocator){0}, s);
    }
}
