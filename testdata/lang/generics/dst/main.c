#include "main.h"

// -- Variables and constants --

// -- Implementation --

int main(void) {
    {
        // Generic extern function (single type parameter).
        so_int* v = newObj(so_int);
        *v = 42;
        if (*v != 42) {
            so_panic("unexpected value");
        }
        freeObj(so_int, v);
    }
    {
        // Generic extern function (multiple type parameters),
        // generic extern type, generic extern method.
        main_Map m = newMap(so_String, so_int, 10);
        if (main_Map_Len(so_String, so_int, &m) != 10) {
            so_panic("unexpected map size");
        }
    }
}
