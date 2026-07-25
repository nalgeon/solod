#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_Movie main_Movie;

typedef struct main_Movie {
    so_int year;
    so_int (*ratingFn)(struct main_Movie);
    void (*updateFn)(struct main_Movie*);
} main_Movie;

// A named function type can be used as a function argument or return value.
typedef so_int (*main_RatingFn)(main_Movie);

typedef void (*main_UpdateFn)(main_Movie*);

// -- Variables and constants --

// Global function variable.
extern void (*main_Global1)();
extern void (*main_Global2)();
