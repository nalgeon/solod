#include "main.h"

// -- Types --

typedef so_int (*sum3Fn)(so_int, so_int, so_int);

// -- Forward declarations --
static so_int sum3(so_int a, so_int b, so_int c);
static so_int pickThird(so_int _0 so_unused, float _1 so_unused, so_int c);
static so_int countUp(so_int _0 so_unused, float _1 so_unused);

// -- Variables and constants --

// stop guards an early return in main.
static so_unused bool stop = false;

// -- Implementation --

static so_int sum3(so_int a, so_int b, so_int c) {
    return a + b + c;
}

// Blank parameters.
static so_int pickThird(so_int _0 so_unused, float _1 so_unused, so_int c) {
    return c;
}

// Unnamed parameters.
static so_int countUp(so_int _0 so_unused, float _1 so_unused) {
    return 1;
}

int main(void) {
    so_int s0 = sum3(1, 2, 3);
    (void)s0;
    so_int (*fn1)(so_int, so_int, so_int) = sum3;
    so_int s1 = fn1(4, 5, 6);
    (void)s1;
    main_Sum3Fn fn2 = sum3;
    so_int s2 = fn2(7, 8, 9);
    (void)s2;
    sum3Fn fn3 = sum3;
    so_int s3 = fn3(3, 3, 3);
    (void)s3;
    // Function literals (anonymous functions) are not supported.
    // fn4 := func(a, b, c int) int {
    // 	return a * b * c
    // }
    // s4 := fn4(2, 3, 4)
    // _ = s4
    main_Sum3Fn fn5 = sub_Sum;
    so_int s5 = fn5(10, 20, 30);
    (void)s5;
    so_int s6 = pickThird(1, 2.0f, 3);
    (void)s6;
    so_int s7 = countUp(1, 2.0f);
    (void)s7;
    // A bare return in main translate to return 0 in C.
    if (stop) {
        return 0;
    }
    so_println("%s", "done");
    return 0;
}
