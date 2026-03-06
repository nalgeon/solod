#include "main.h"

// -- Forward declarations (functions and methods) --
static so_int add(so_int a, so_int b);

// -- add.go --

static so_int add(so_int a, so_int b) {
    return a + b;
}

// -- main.go --

int main(void) {
    so_println("%" PRId64, add(1, 2));
}
