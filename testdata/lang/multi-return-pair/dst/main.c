#include "main.h"

// -- Forward declarations (functions and methods) --
static so_Result divmod(so_int a, so_int b);
static so_Result check(so_int n);
static so_Result greet(so_String name);
static so_Result forwardDivmod(void);

// -- Implementation --

// Same-type pair.
static so_Result divmod(so_int a, so_int b) {
    return (so_Result){.val.as_int = a / b, .val2.as_int = a % b};
}

// Mixed types.
static so_Result check(so_int n) {
    return (so_Result){.val.as_bool = n > 0, .val2.as_int = n * 2};
}

// String pair.
static so_Result greet(so_String name) {
    return (so_Result){.val.as_string = so_str("hello"), .val2.as_string = name};
}

// Forwarding.
static so_Result forwardDivmod(void) {
    return divmod(10, 3);
}

int main(void) {
    {
        // Destructure into new variables.
        so_Result _res1 = divmod(10, 3);
        so_int q = _res1.val.as_int;
        so_int r = _res1.val2.as_int;
        (void)q;
        (void)r;
        // Blank identifiers.
        so_Result _res2 = divmod(10, 3);
        so_int r2 = _res2.val2.as_int;
        (void)r2;
        so_Result _res3 = divmod(10, 3);
        so_int q3 = _res3.val.as_int;
        (void)q3;
        // Partial reassignment.
        so_Result _res4 = divmod(20, 7);
        so_int q4 = _res4.val.as_int;
        r2 = _res4.val2.as_int;
        (void)q4;
        // Assign to existing variables.
        q = 0;
        r = 0;
        so_Result _res5 = divmod(20, 7);
        q = _res5.val.as_int;
        r = _res5.val2.as_int;
    }
    {
        // Mixed types.
        so_Result _res6 = check(5);
        bool ok = _res6.val.as_bool;
        so_int doubled = _res6.val2.as_int;
        (void)ok;
        (void)doubled;
    }
    {
        // String pair.
        so_Result _res7 = greet(so_str("world"));
        so_String greeting = _res7.val.as_string;
        so_String name = _res7.val2.as_string;
        (void)greeting;
        (void)name;
    }
    {
        // If-init with multi-return.
        {
            so_Result _res8 = divmod(10, 3);
            so_int q = _res8.val.as_int;
            so_int r = _res8.val2.as_int;
            if (r > 0) {
                (void)q;
            }
        }
    }
    {
        // Forwarding.
        so_Result _res9 = forwardDivmod();
        so_int q = _res9.val.as_int;
        so_int r = _res9.val2.as_int;
        (void)q;
        (void)r;
    }
}
