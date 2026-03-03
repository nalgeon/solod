#include "main.h"

// -- Implementation --

int main(void) {
    {
        // mem.Alloc and mem.Dealloc
        so_Result _res1 = mem_Alloc(main_Point, mem_System);
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
        mem_Dealloc(main_Point, mem_System, p);
    }
    {
        // mem.AllocSlice and mem.DeallocSlice
        so_Result _res2 = mem_AllocSlice(so_int, mem_System, 3, 3);
        so_Slice slice = _res2.val.as_slice;
        so_Error err = _res2.err;
        if (err != NULL) {
            so_panic("AllocSlice: allocation failed");
        }
        so_index(so_int, slice, 0) = 11;
        so_index(so_int, slice, 1) = 22;
        so_index(so_int, slice, 2) = 33;
        if (so_index(so_int, slice, 0) != 11 || so_index(so_int, slice, 1) != 22 || so_index(so_int, slice, 2) != 33) {
            so_panic("AllocSlice: unexpected value");
        }
        mem_DeallocSlice(so_int, mem_System, slice);
    }
    {
        // mem.New and mem.Free
        main_Point* p = mem_New(main_Point);
        p->x = 11;
        p->y = 22;
        if (p->x != 11 || p->y != 22) {
            so_panic("New: unexpected value");
        }
        mem_Free(main_Point, p);
    }
    {
        // mem.NewSlice and mem.FreeSlice
        so_Slice slice = mem_NewSlice(so_int, 3, 3);
        so_index(so_int, slice, 0) = 11;
        so_index(so_int, slice, 1) = 22;
        so_index(so_int, slice, 2) = 33;
        if (so_index(so_int, slice, 0) != 11 || so_index(so_int, slice, 1) != 22 || so_index(so_int, slice, 2) != 33) {
            so_panic("NewSlice: unexpected value");
        }
        mem_FreeSlice(so_int, slice);
    }
}
