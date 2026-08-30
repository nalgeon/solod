#include "main.h"

// -- Forward declarations --
static so_Error makeTea(so_int arg);
static so_R_int_err work(so_int n);

// -- Variables and constants --
so_Error main_ErrOutOfTea = errors_New("no more tea available");

// -- Implementation --

static so_Error makeTea(so_int arg) {
    if (arg == 42) {
        return main_ErrOutOfTea;
    }
    return (so_Error){};
}

static so_R_int_err work(so_int n) {
    if (n == 42) {
        return (so_R_int_err){.val = 0, .err = main_ErrOutOfTea};
    }
    return (so_R_int_err){.val = n, .err = (so_Error){}};
}

int main(void) {
    {
        // Nil and non-nil errors.
        so_Error err = makeTea(7);
        if (err.self != NULL) {
            so_panic("err != nil");
        }
        err = makeTea(42);
        if (err.self == NULL) {
            so_panic("err == nil");
        }
        if (err.self != main_ErrOutOfTea.self) {
            so_panic("err != ErrOutOfTea");
        }
    }
    {
        // Variable of type error.
        so_Error err = {};
        if (err.self != NULL) {
            so_panic("err != nil");
        }
        err = makeTea(42);
        if (err.self == NULL) {
            so_panic("err == nil");
        }
    }
    {
        // Multiple returns with error.
        so_R_int_err _res1 = work(11);
        so_int r1 = _res1.val;
        so_Error err = _res1.err;
        if (r1 != 11) {
            so_panic("unexpected result");
        }
        if (err.self != NULL) {
            so_panic("unexpected error");
        }
        (void)r1;
        so_R_int_err _res2 = work(42);
        so_int r2 = _res2.val;
        err = _res2.err;
        if (r2 != 0) {
            so_panic("unexpected result");
        }
        if (err.self != main_ErrOutOfTea.self) {
            so_panic("expected ErrOutOfTea");
        }
        (void)r2;
    }
    {
        // Printing errors.
        so_Error err = makeTea(42);
        so_println("%s %s", "err =", so_error_cstr(err));
        so_println("%s %.*s", "err text =", so_Error_Error(err).len, so_Error_Error(err).ptr);
        so_Error nilErr = {};
        so_println("%s %s", "err =", so_error_cstr(nilErr));
    }
    {
        // Error comparison.
        so_Error err1 = makeTea(42);
        so_Error err2 = makeTea(42);
        if (err1.self != err2.self) {
            so_panic("want err1 == err2");
        }
        if (err1.self != main_ErrOutOfTea.self) {
            so_panic("want err1 == ErrOutOfTea");
        }
    }
    {
        // Error method call.
        if (so_string_ne(so_Error_Error(main_ErrOutOfTea), so_str("no more tea available"))) {
            so_panic("unexpected .Error() result");
        }
    }
    {
        // Error method expression.
        if (so_string_ne(so_Error_Error(main_ErrOutOfTea), so_str("no more tea available"))) {
            so_panic("unexpected error.Error() result");
        }
    }
    return 0;
}
