#include "main.h"

// -- Types --

typedef struct calc calc;

typedef struct calc {
} calc;

// -- Forward declarations --
static so_int calc_Sum(void* self, so_int a[3]);
static so_int calc_Sum2D(void* self, so_int m[2][3]);
static so_int calc_Sum3D(void* self, so_int m[2][3][4]);
static so_int sum(so_int a[3]);
static so_int sum2D(so_int m[2][3]);
static so_int sum3D(so_int m[2][3][4]);
static so_int sumNamed(main_Matrix m);
static so_int apply(so_int (*f)(so_int*), so_int a[3]);

// -- Implementation --

static so_int calc_Sum(void* self, so_int a[3]) {
    (void)self;
    return sum(a);
}

static so_int calc_Sum2D(void* self, so_int m[2][3]) {
    (void)self;
    return sum2D(m);
}

static so_int calc_Sum3D(void* self, so_int m[2][3][4]) {
    (void)self;
    return sum3D(m);
}

static so_int sum(so_int a[3]) {
    return a[0] + a[1] + a[2];
}

static so_int sum2D(so_int m[2][3]) {
    so_int total = 0;
    for (so_int i = 0; i < 2; i++) {
        for (so_int j = 0; j < 3; j++) {
            total += m[i][j];
        }
    }
    return total;
}

static so_int sum3D(so_int m[2][3][4]) {
    return m[0][0][0] + m[1][2][3];
}

static so_int sumNamed(main_Matrix m) {
    return m[0][0] + m[1][2];
}

// An anonymous function type with array parameters as an argument.
static so_int apply(so_int (*f)(so_int*), so_int a[3]) {
    return f(a);
}

int main(void) {
    so_int a[3] = {1, 2, 3};
    so_int m[2][3] = {{1, 2, 3}, {4, 5, 6}};
    so_int m3D[2][3][4] = {};
    m3D[0][0][0] = 10;
    m3D[1][2][3] = 11;
    {
        // Named function type.
        main_SumFn f = sum;
        if (f(a) != 6) {
            so_panic("unexpected f");
        }
        main_SumFn2D f2D = sum2D;
        if (f2D(m) != 21) {
            so_panic("unexpected f2D");
        }
        main_SumFn3D f3D = sum3D;
        if (f3D(m3D) != 21) {
            so_panic("unexpected f3D");
        }
        main_SumFnNamed fn = sumNamed;
        if (fn((so_int[2][3]){{1, 2, 3}, {4, 5, 6}}) != 7) {
            so_panic("unexpected fn");
        }
    }
    {
        // Anonymous function type variable.
        so_int (*g)(so_int*) = sum;
        if (g(a) != 6) {
            so_panic("unexpected g");
        }
        so_int (*g2D)(so_int (*)[3]) = sum2D;
        if (g2D(m) != 21) {
            so_panic("unexpected g2D");
        }
        so_int (*g3D)(so_int (*)[3][4]) = sum3D;
        if (g3D(m3D) != 21) {
            so_panic("unexpected g3D");
        }
    }
    {
        // Struct field.
        main_Calc c = (main_Calc){.sum = sum, .sum2D = sum2D, .sum3D = sum3D};
        if (c.sum(a) != 6) {
            so_panic("unexpected c.sum");
        }
        if (c.sum2D(m) != 21) {
            so_panic("unexpected c.sum2D");
        }
        if (c.sum3D(m3D) != 21) {
            so_panic("unexpected c.sum3D");
        }
    }
    {
        // Interface method.
        calc c = (calc){};
        main_Summer s = (main_Summer){.self = &c, .Sum = calc_Sum, .Sum2D = calc_Sum2D, .Sum3D = calc_Sum3D};
        if (main_Summer_Sum(s, a) != 6) {
            so_panic("unexpected s.Sum");
        }
        if (main_Summer_Sum2D(s, m) != 21) {
            so_panic("unexpected s.Sum2D");
        }
        if (main_Summer_Sum3D(s, m3D) != 21) {
            so_panic("unexpected s.Sum3D");
        }
    }
    {
        // Anonymous function type as an argument.
        if (apply(sum, a) != 6) {
            so_panic("unexpected apply");
        }
    }
    return 0;
}
