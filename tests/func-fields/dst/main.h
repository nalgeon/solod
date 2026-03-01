#include "solod.h"

typedef struct main_Movie {
    so_int year;
    so_int (*ratingFn)(struct main_Movie m);
} main_Movie;
