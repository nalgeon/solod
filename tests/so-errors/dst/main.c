#include "main.h"

// -- Forward declarations (functions and methods) --
static so_Result work(so_int n);

// -- Implementation --
so_Error main_Err42 = errors_New(so_strlit("42"));

static so_Result work(so_int n) {
    if (n == 42) {
        return (so_Result){.val.as_int = 0, .err = main_Err42};
    }
    return (so_Result){.val.as_int = 42, .err = NULL};
}

int main(void) {
    so_Result _res1 = work(11);
    so_int r1 = _res1.val.as_int;
    so_Error err = _res1.err;
    if (err != NULL) {
        so_panic("unexpected error");
    }
    (void)r1;
    so_Result _res2 = work(42);
    so_int r2 = _res2.val.as_int;
    err = _res2.err;
    if (err != main_Err42) {
        so_panic("expected Err42");
    }
    (void)r2;
}
