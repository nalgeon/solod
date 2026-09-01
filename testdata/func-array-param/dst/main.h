#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_Calc main_Calc;

// A named function type with array parameters.
typedef so_int (*main_SumFn)(so_int*);

typedef so_int (*main_SumFn2D)(so_int (*)[3]);

typedef so_int (*main_SumFn3D)(so_int (*)[3][4]);

// A named array type as a function parameter.
typedef so_int main_Matrix[2][3];

typedef so_int (*main_SumFnNamed)(main_Matrix);

// A struct field of a function type with array parameters.
typedef struct main_Calc {
    so_int (*sum)(so_int*);
    so_int (*sum2D)(so_int (*)[3]);
    so_int (*sum3D)(so_int (*)[3][4]);
} main_Calc;

// An interface method with array parameters.
typedef struct main_Summer {
    void* self;
    so_int (*Sum)(void* self, so_int*);
    so_int (*Sum2D)(void* self, so_int (*)[3]);
    so_int (*Sum3D)(void* self, so_int (*)[3][4]);
} main_Summer;

static inline so_int main_Summer_Sum(main_Summer self, so_int a[3]) {
    return self.Sum(self.self, a);
}

static inline so_int main_Summer_Sum2D(main_Summer self, so_int m[2][3]) {
    return self.Sum2D(self.self, m);
}

static inline so_int main_Summer_Sum3D(main_Summer self, so_int m[2][3][4]) {
    return self.Sum3D(self.self, m);
}
