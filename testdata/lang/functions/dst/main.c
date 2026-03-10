#include "main.h"

// -- Forward declarations (types) --
typedef so_int (*sum3Fn)(so_int, so_int, so_int);

// -- Forward declarations (functions and methods) --
static so_int sum3(so_int a, so_int b, so_int c);

// -- Implementation --

typedef so_int (*sum3Fn)(so_int, so_int, so_int);

static so_int sum3(so_int a, so_int b, so_int c) {
    return a + b + c;
}

int main(void) {
    so_int s0 = sum3(1, 2, 3);
    (void)s0;
    main_Sum3Fn fn1 = sum3;
    so_int s1 = fn1(4, 5, 6);
    (void)s1;
    main_Sum3Fn fn2 = sum3;
    so_int s2 = fn2(7, 8, 9);
    (void)s2;
    sum3Fn fn3 = sum3;
    so_int s3 = fn3(3, 3, 3);
    (void)s3;
}
