#include "main.h"

// -- Forward declarations (functions and methods) --
static void xopen(so_int* x);
static void xclose(void* a);

// -- Implementation --

static void xopen(so_int* x) {
    so_println("%s %" PRId64, "open", *x);
}

static void xclose(void* a) {
    so_int* x = (so_int*)a;
    so_println("%s %" PRId64, "close", *x);
}

int main(void) {
    so_int x = 42;
    xopen(&x);
    so_defer(xclose, &x);
    so_println("%s", "working...");
}
