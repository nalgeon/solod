#pragma once
#include "so/builtin/builtin.h"

// -- Embeds --

typedef struct {
    int val;
} main_Box;

// -- Functions and methods --

#define identity(T, val_) ({ \
    val_; \
})

#define setPtr(T, ptr_, val_) do { \
    *ptr_ = val_; \
} while (0)

// increment adds two to n.
//
#define increment(T, n_) ({ \
    /* A line comment in a macro becomes a block comment. */ \
    T _n = n_; \
    /* A * / in the text does not close the block comment. */ \
    _n = _n + 1; \
    /* A block comment stays as it is. */ \
    _n = _n + 1; \
    _n; \
})

#define a(T, n_) ({ \
    so_int _some = 11; \
    (void)_some; \
    T _x = b(T, (n_)) + 1; \
    _x; \
})

#define b(T, n_) ({ \
    double _some = 22.2; \
    (void)_some; \
    T _x = c(T, (n_)) + 1; \
    _x; \
})

#define c(T, n_) ({ \
    so_String _some = so_str("33"); \
    (void)_some; \
    T _x = n_ + 1; \
    _x; \
})

// Blank parameters.
//
#define pickThird(T, _1_, _2_, c_) ({ \
    c_; \
})

#define work(T, v_) ({ \
    (so_R_ptr_err){.val = v_, .err = (so_Error){}}; \
})

#define main_Box_set(T, b_, val_) do { \
    b_->val = val_; \
} while (0)
