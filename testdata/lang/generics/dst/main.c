#include "main.h"

// -- Implementation --

int main(void) {
    so_int* v = newObj(so_int);
    *v = 42;
    if (*v != 42) {
        so_panic("unexpected value");
    }
    freeObj(so_int, v);
}
