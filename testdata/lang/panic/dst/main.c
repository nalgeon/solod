#include "main.h"

// -- Forward declarations (functions and methods) --
static void panicLiteral(void);
static void panicString(void);
static void panicError(void);

// -- Implementation --
so_Error main_ErrNotFound = errors_New("not found");

static void panicLiteral(void) {
    so_panic("something went wrong");
}

static void panicString(void) {
    so_String msg = so_str("runtime error");
    so_panic(so_cstr(msg));
}

static void panicError(void) {
    so_Error err = main_ErrNotFound;
    so_panic(err->msg);
}

int main(void) {
    if (false) {
        panicLiteral();
    }
    if (false) {
        panicString();
    }
    if (false) {
        panicError();
    }
}
