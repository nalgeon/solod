#include "main.h"

// -- Forward declarations --
static so_R_int_int divmod(so_int a, so_int b);
static so_R_bool_int check(so_int n);
static so_R_str_str greet(so_String name);
static so_R_int_int forwardDivmod(void);

// -- Implementation --

// Same-type pair.
static so_R_int_int divmod(so_int a, so_int b) {
    return (so_R_int_int){.val = a / b, .val2 = a % b};
}

// Mixed types.
static so_R_bool_int check(so_int n) {
    return (so_R_bool_int){.val = n > 0, .val2 = n * 2};
}

// String pair.
static so_R_str_str greet(so_String name) {
    return (so_R_str_str){.val = so_str("hello"), .val2 = name};
}

// Forwarding.
static so_R_int_int forwardDivmod(void) {
    return divmod(10, 3);
}

int main(void) {
    {
        // Destructure into new variables.
        so_R_int_int _res1 = divmod(10, 3);
        so_int q = _res1.val;
        so_int r = _res1.val2;
        (void)q;
        (void)r;
        // Blank identifiers.
        so_R_int_int _res2 = divmod(10, 3);
        so_int r2 = _res2.val2;
        (void)r2;
        so_R_int_int _res3 = divmod(10, 3);
        so_int q3 = _res3.val;
        (void)q3;
        // Partial reassignment.
        so_R_int_int _res4 = divmod(20, 7);
        so_int q4 = _res4.val;
        r2 = _res4.val2;
        (void)q4;
        // Assign to existing variables.
        q = 0;
        r = 0;
        so_R_int_int _res5 = divmod(20, 7);
        q = _res5.val;
        r = _res5.val2;
    }
    {
        // Mixed types.
        so_R_bool_int _res6 = check(5);
        bool ok = _res6.val;
        so_int doubled = _res6.val2;
        (void)ok;
        (void)doubled;
    }
    {
        // String pair.
        so_R_str_str _res7 = greet(so_str("world"));
        so_String greeting = _res7.val;
        so_String name = _res7.val2;
        (void)greeting;
        (void)name;
    }
    {
        // If-init with multi-return.
        {
            so_R_int_int _res8 = divmod(10, 3);
            so_int q = _res8.val;
            so_int r = _res8.val2;
            if (r > 0) {
                (void)q;
            }
        }
    }
    {
        // Forwarding.
        so_R_int_int _res9 = forwardDivmod();
        so_int q = _res9.val;
        so_int r = _res9.val2;
        (void)q;
        (void)r;
    }
}
