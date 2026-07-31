#include "main.h"

// -- Forward declarations --
static so_int add(so_int a, so_int b);

// -- Variables and constants --
static so_int x = 11;
so_int main_Y = 22;
static const so_unused so_int z = 33;

// -- add.go --

static so_int add(so_int a, so_int b) {
    return a + b + x + main_Y + z;
}

// -- main.go --

int main(void) {
    so_println("%" PRIdINT, add(1, 2));
    return 0;
}
