#pragma once
#include "so/builtin/builtin.h"

// -- Functions and methods --

#define identity(T, val_) ({ \
    (val_); \
})

#define setPtr(T, ptr_, val_) do { \
    *(ptr_) = (val_); \
} while (0)

#define increment(T, n_) ({ \
    T _n = (n_); \
    _n = _n + 1; \
    _n = _n + 1; \
    _n; \
})

#define a(T, n_) ({ \
    so_int some = 11; \
    (void)some; \
    T x = b(T, (n_)) + 1; \
    x; \
})

#define b(T, n_) ({ \
    double some = 22.2; \
    (void)some; \
    T x = c(T, (n_)) + 1; \
    x; \
})

#define c(T, n_) ({ \
    so_String some = so_str("33"); \
    (void)some; \
    T x = (n_) + 1; \
    x; \
})

#define work(T, v_) ({ \
    (so_R_ptr_err){.val = (v_), .err = NULL}; \
})

#define main_Box_set(T, b_, val_) do { \
    (b_)->val = (val_); \
} while (0)

// -- Embeds --

typedef struct {
    int val;
} main_Box;
