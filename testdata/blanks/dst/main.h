#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_Point main_Point;
typedef struct main_Wrapper main_Wrapper;
typedef struct main_Value main_Value;

// A named struct with blank fields.
typedef struct main_Point {
    so_int x;
    so_int _1;
    so_int y;
    so_int _3;
} main_Point;

// An inner struct field with blank fields.
typedef struct main_Wrapper {
    struct {
        so_int n;
        so_String _1;
        so_String _2;
    } inner;
} main_Wrapper;

// Struct with blank and unnamed parameters.
typedef struct main_Value {
    so_int x;
} main_Value;

// A named array type.
typedef so_int main_Nums[3];

// Interface with unnamed and blank parameters.
typedef struct main_Valuer {
    void* self;
    so_int (*Decr2)(void* self, so_int, float);
    so_int (*Incr3)(void* self, so_int, float, so_int);
} main_Valuer;

static inline so_int main_Valuer_Decr2(main_Valuer self, so_int _0, float _1) {
    return self.Decr2(self.self, _0, _1);
}

static inline so_int main_Valuer_Incr3(main_Valuer self, so_int _0, float _1, so_int n) {
    return self.Incr3(self.self, _0, _1, n);
}

// -- Functions and methods --

// Unnamed receiver.
so_int main_Value_One(main_Value self);

// Blank receiver.
so_int main_Value_Two(main_Value self);

// Unnamed method parameters.
so_int main_Value_Decr1(main_Value v, so_int _1 so_unused);
so_int main_Value_Decr2(void* self, so_int _1 so_unused, float _2 so_unused);

// Blank method parameters.
so_int main_Value_Incr1(main_Value v, so_int _1 so_unused);
so_int main_Value_Incr2(main_Value v, so_int n, float _2 so_unused);
so_int main_Value_Incr3(void* self, so_int _1 so_unused, float _2 so_unused, so_int n);

// Unnamed generic function parameters.
#define unnamedGen1(T, _1_) ({ \
    1; \
})

#define unnamedGen2(T, _1_, _2_) ({ \
    2; \
})

// Blank generic function parameters.
#define blankGen1(T, _1_) ({ \
    1; \
})

#define blankGen2(T, n_, _2_) ({ \
    n_; \
})

#define blankGen3(T, _1_, _2_, n_) ({ \
    n_; \
})
