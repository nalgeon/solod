#include "main.h"

// -- Forward declarations --
static void xopen(so_int* x);
static void xclose(void* a);
static void funcScope(void);
static so_int funcWithReturn(void);
static so_R_int_err funcReturnCall(void);
static so_int funcReturnVar(void);
static so_R_int_err funcCalc(void);

// -- Variables and constants --
static so_int state = 0;

// -- Implementation --

static void xopen(so_int* x) {
    (*x)++;
}

static void xclose(void* a) {
    so_int* x = (so_int*)a;
    (*x)--;
}

static void funcScope(void) {
    xopen(&state);
    if (state != 1) {
        xclose(&state);
        so_panic("unexpected state");
    }
    xclose(&state);
}

static so_int funcWithReturn(void) {
    xopen(&state);
    if (state != 1) {
        xclose(&state);
        so_panic("unexpected state");
    }
    xclose(&state);
    return 42;
}

static so_R_int_err funcReturnCall(void) {
    xopen(&state);
    so_R_int_err _res1 = funcCalc();
    xclose(&state);
    return _res1;
}

static so_int funcReturnVar(void) {
    xopen(&state);
    so_int _res1 = state;
    xclose(&state);
    return _res1;
}

static so_R_int_err funcCalc(void) {
    if (state != 1) {
        so_panic("unexpected state");
    }
    return (so_R_int_err){.val = 42, .err = (so_Error){}};
}

int main(void) {
    funcScope();
    if (state != 0) {
        so_panic("unexpected state");
    }
    funcWithReturn();
    if (state != 0) {
        so_panic("unexpected state");
    }
    funcReturnCall();
    if (state != 0) {
        so_panic("unexpected state");
    }
    if (funcReturnVar() != 1) {
        so_panic("unexpected return value");
    }
    if (state != 0) {
        so_panic("unexpected state");
    }
    return 0;
}
