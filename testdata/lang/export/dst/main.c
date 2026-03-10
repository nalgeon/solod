#include "main.h"

// -- Forward declarations (functions and methods) --
static bool someFunc(so_int x, so_int y);

// -- Implementation --
static const so_int someConst = 7;
const so_int main_SomeConst = 7;
static so_int someVar = 42;
so_int main_SomeVar = 42;

static bool someFunc(so_int x, so_int y) {
    return x > y + someConst;
}

bool main_SomeFunc(so_int x, so_int y) {
    return x > y + someVar;
}

int main(void) {
    (void)someFunc(1, 2);
    main_SomeFunc(3, 4);
}
