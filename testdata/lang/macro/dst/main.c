#include "main.h"

// -- Variables and constants --

// -- Implementation --

int main(void) {
    so_println("%s", "lang/macro - start");
    {
        so_print("%s", "lang/macro: Function with return");
        so_int x = identity(so_int, 42);
        if (x != (so_int)(42)) {
            so_panic("x != 42");
        }
        so_println("%s", " - ok");
    }
    {
        so_print("%s", "lang/macro: Function w/o return");
        so_int y = 0;
        setPtr(so_int, &y, 42);
        if (y != 42) {
            so_panic("y != 42");
        }
        so_println("%s", " - ok");
    }
    {
        so_print("%s", "lang/macro: Pass an expression as an argument");
        so_int x = increment(so_int, 1 + 1);
        if (x != 4) {
            so_panic("x != 4");
        }
        so_println("%s", " - ok");
    }
    {
        so_print("%s", "lang/macro: Nested calls with variable shadowing");
        so_int z = a(so_int, 42);
        if (z != 45) {
            so_panic("z != 45");
        }
        so_println("%s", " - ok");
    }
    {
        so_print("%s", "lang/macro: Generic method");
        main_Box b = {0};
        main_Box_set(so_int, &b, 42);
        if (b.val != 42) {
            so_panic("b.val != 42");
        }
        so_println("%s", " - ok");
    }
    {
        so_print("%s", "lang/macro: Multi-return");
        so_int v = 42;
        so_R_ptr_err _res1 = work(so_int, &v);
        so_int* res = _res1.val;
        so_Error err = _res1.err;
        if (err != NULL) {
            so_panic("err != nil");
        }
        if (*res != 42) {
            so_panic("res != 42");
        }
        so_println("%s", " - ok");
    }
    so_println("%s", "lang/macro - ok");
}
