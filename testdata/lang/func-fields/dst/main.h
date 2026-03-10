#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_Movie {
    so_int year;
    so_int (*ratingFn)(struct main_Movie m);
} main_Movie;

// Must define a named function type to use it
// as function argument or return value.
typedef so_int (*main_RatingFn)(main_Movie);
