#include "main.h"

// -- Forward declarations --
static void withDefer(void);
static void allocTest(void);
static void arenaTest(void);

// -- alloc.go --

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

static void allocTest(void) {
    {
        // TryAlloc and Free.
        so_R_ptr_err _res1 = mem_TryAlloc(main_Point, mem_System);
        main_Point* p = _res1.val;
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
        so_R_slice_err _res2 = mem_TryAllocSlice(so_int, mem_System, 3, 3);
        so_Slice slice = _res2.val;
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
    {
        // TryReallocSlice with explicit allocator.
        so_R_slice_err _res3 = mem_TryAllocSlice(so_int, mem_System, 3, 3);
        so_Slice slice = _res3.val;
        so_Error err = _res3.err;
        if (err != NULL) {
            so_panic("ReallocSlice: initial allocation failed");
        }
        so_at(so_int, slice, 0) = 11;
        so_at(so_int, slice, 1) = 22;
        so_at(so_int, slice, 2) = 33;
        so_R_slice_err _res4 = mem_TryReallocSlice(so_int, mem_System, slice, 3, 6);
        slice = _res4.val;
        err = _res4.err;
        if (err != NULL) {
            so_panic("ReallocSlice: reallocation failed");
        }
        if (so_len(slice) != 3 || so_cap(slice) != 6) {
            so_panic("ReallocSlice: unexpected len/cap");
        }
        if (so_at(so_int, slice, 0) != 11 || so_at(so_int, slice, 1) != 22 || so_at(so_int, slice, 2) != 33) {
            so_panic("ReallocSlice: data not preserved");
        }
        mem_FreeSlice(so_int, mem_System, slice);
    }
    {
        // ReallocSlice with default allocator.
        so_Slice slice = mem_AllocSlice(so_int, (mem_Allocator){0}, 2, 2);
        so_at(so_int, slice, 0) = 44;
        so_at(so_int, slice, 1) = 55;
        slice = mem_ReallocSlice(so_int, (mem_Allocator){0}, slice, 4, 8);
        if (so_len(slice) != 4 || so_cap(slice) != 8) {
            so_panic("ReallocSlice default: unexpected len/cap");
        }
        if (so_at(so_int, slice, 0) != 44 || so_at(so_int, slice, 1) != 55) {
            so_panic("ReallocSlice default: data not preserved");
        }
        // New elements should be zeroed.
        if (so_at(so_int, slice, 2) != 0 || so_at(so_int, slice, 3) != 0) {
            so_panic("ReallocSlice default: new elements not zeroed");
        }
        mem_FreeSlice(so_int, (mem_Allocator){0}, slice);
    }
    {
        // ReallocSlice from empty slice.
        so_Slice empty = {0};
        so_Slice slice = mem_ReallocSlice(so_int, (mem_Allocator){0}, empty, 3, 4);
        if (so_len(slice) != 3 || so_cap(slice) != 4) {
            so_panic("ReallocSlice empty: unexpected len/cap");
        }
        if (so_at(so_int, slice, 0) != 0 || so_at(so_int, slice, 1) != 0 || so_at(so_int, slice, 2) != 0) {
            so_panic("ReallocSlice empty: not zeroed");
        }
        mem_FreeSlice(so_int, (mem_Allocator){0}, slice);
    }
    {
        // Free with nil or an empty slice.
        main_Point* p = NULL;
        mem_Free(main_Point, (mem_Allocator){0}, p);
        so_Slice empty = {0};
        mem_FreeSlice(so_int, (mem_Allocator){0}, empty);
    }
    {
        // Free string.
        so_Slice b = mem_AllocSlice(so_byte, (mem_Allocator){0}, 3, 3);
        so_at(so_byte, b, 0) = 'h';
        so_at(so_byte, b, 1) = 'i';
        so_at(so_byte, b, 2) = '!';
        so_String s1 = so_bytes_string(b);
        mem_FreeString((mem_Allocator){0}, s1);
        so_String s2 = so_str("");
        mem_FreeString((mem_Allocator){0}, s2);
    }
    {
        // Free with defer.
        withDefer();
    }
}

// -- arena.go --

static void arenaTest(void) {
    {
        // Arena allocator.
        so_Slice buf = so_make_slice(so_byte, 1024, 1024);
        mem_Arena arena = mem_NewArena(buf);
        mem_Allocator a = (mem_Allocator){.self = &arena, .Alloc = mem_Arena_Alloc, .Free = mem_Arena_Free, .Realloc = mem_Arena_Realloc};
        // Allocate a Point.
        so_R_ptr_err _res1 = mem_TryAlloc(main_Point, a);
        main_Point* p = _res1.val;
        so_Error err = _res1.err;
        if (err != NULL) {
            so_panic("initial allocation failed");
        }
        p->x = 11;
        p->y = 22;
        if (p->x != 11 || p->y != 22) {
            so_panic("unexpected p.x or p.y");
        }
        // Free is a no-op.
        mem_Free(main_Point, a, p);
        // Reset and reallocate.
        mem_Arena_Reset(&arena);
        so_R_ptr_err _res2 = mem_TryAlloc(main_Point, a);
        main_Point* p2 = _res2.val;
        err = _res2.err;
        if (err != NULL) {
            so_panic("allocation after reset failed");
        }
        // Memory should be zeroed.
        if (p2->x != 0 || p2->y != 0) {
            so_panic("memory not zeroed after reset");
        }
        p2->x = 33;
        p2->y = 44;
    }
}

// -- main.go --

int main(void) {
    allocTest();
    arenaTest();
}
