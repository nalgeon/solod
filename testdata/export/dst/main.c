#include "main.h"

// -- Forward declarations --
static bool someFunc(so_int x, so_int y);

// -- Variables and constants --
static const so_unused int64_t someConst = 7;
static so_unused so_int someVar = 42;
so_int main_SomeVar = 42;

// -- Implementation --

static bool someFunc(so_int x, so_int y) {
    return x > y + someConst;
}

bool main_SomeFunc(so_int x, so_int y) {
    return x > y + someVar;
}

int main(void) {
    (void)someFunc(1, 2);
    main_SomeFunc(3, 4);
    return 0;
}
