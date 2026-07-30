#include "main.h"

// -- Types --
typedef so_int step;

// -- Forward declarations --
static so_String step_String(step s);

// -- Implementation --

static so_String step_String(step s) {
    if (s == 0) {
        return so_str("zero");
    }
    return so_str("step");
}

int main(void) {
    {
        // Generic extern function (single type parameter).
        so_int* v = newObj(so_int);
        *v = 42;
        if (*v != 42) {
            so_panic("unexpected value");
        }
        freeObj(so_int, (v));
    }
    {
        // Generic extern function (multiple type parameters),
        // generic extern type, generic extern method.
        main_Map m = newMap(so_String, so_int, (10));
        if (main_Map_Len(so_String, so_int, (&m)) != 10) {
            so_panic("unexpected map size");
        }
    }
    {
        // Generic inline functions with named constraints.
        if (add(so_int, (2), (3)) != 5) {
            so_panic("unexpected sum");
        }
        if (so_string_ne(first(so_String, (so_str("a")), (so_str("b"))), so_str("a"))) {
            so_panic("unexpected value");
        }
        step s = same(step, ((step)(1)));
        if (s != 1 || so_string_ne(step_String(s), so_str("step"))) {
            so_panic("unexpected step");
        }
    }
    {
        // A constraint declared inside a function body is not emitted.
        if (add(so_int, (1), (1)) != 2) {
            so_panic("unexpected sum");
        }
    }
    return 0;
}
