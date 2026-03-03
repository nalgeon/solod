#include "main.h"

// -- Forward declarations (functions and methods) --
static so_Error makeTea(so_int arg);

// -- Implementation --
so_Error main_ErrOutOfTea = errors_New(so_strlit("no more tea available"));

static so_Error makeTea(so_int arg) {
    if (arg == 42) {
        return main_ErrOutOfTea;
    }
    return NULL;
}

int main(void) {
    {
        // Nil and non-nil errors.
        so_Error err = makeTea(7);
        if (err != NULL) {
            so_panic("err != nil");
        }
        err = makeTea(42);
        if (err == NULL) {
            so_panic("err == nil");
        }
        if (err != main_ErrOutOfTea) {
            so_panic("err != ErrOutOfTea");
        }
    }
    {
        // Variable of type error.
        so_Error err = NULL;
        if (err != NULL) {
            so_panic("err != nil");
        }
        err = makeTea(42);
        if (err == NULL) {
            so_panic("err == nil");
        }
    }
}
