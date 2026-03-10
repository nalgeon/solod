#include "main.h"

// -- Forward declarations (functions and methods) --
static so_Error makeTea(so_int arg);
static so_Result work(so_int n);

// -- Implementation --
so_Error main_ErrOutOfTea = errors_New("no more tea available");

static so_Error makeTea(so_int arg) {
    if (arg == 42) {
        return main_ErrOutOfTea;
    }
    return NULL;
}

static so_Result work(so_int n) {
    if (n == 42) {
        return (so_Result){.val.as_int = 0, .err = main_ErrOutOfTea};
    }
    return (so_Result){.val.as_int = n, .err = NULL};
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
    {
        // Multiple returns with error.
        so_Result _res1 = work(11);
        so_int r1 = _res1.val.as_int;
        so_Error err = _res1.err;
        if (r1 != 11) {
            so_panic("unexpected result");
        }
        if (err != NULL) {
            so_panic("unexpected error");
        }
        (void)r1;
        so_Result _res2 = work(42);
        so_int r2 = _res2.val.as_int;
        err = _res2.err;
        if (r2 != 0) {
            so_panic("unexpected result");
        }
        if (err != main_ErrOutOfTea) {
            so_panic("expected ErrOutOfTea");
        }
        (void)r2;
    }
    {
        // Printing errors.
        so_Error err = makeTea(42);
        so_println("%s %s", "err =", err->msg);
    }
}
