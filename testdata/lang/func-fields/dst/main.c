#include "main.h"

// -- Forward declarations (functions and methods) --
static so_int freshness(main_Movie m);
static main_RatingFn getRatingFn(void);
static so_int rateMovie(main_Movie m, main_RatingFn f);

// -- Implementation --

static so_int freshness(main_Movie m) {
    return m.year - 1970;
}

static main_RatingFn getRatingFn(void) {
    return freshness;
}

static so_int rateMovie(main_Movie m, main_RatingFn f) {
    return f(m);
}

// Returning anonymous function types is not supported.
// func getRatingFn() func(m Movie) int {
// 	return freshness
// }
// Passing anonymous function types is not supported.
// func rateMovie(m Movie, f func(m Movie) int) int {
// 	return f(m)
// }
int main(void) {
    {
        // Function struct field.
        main_Movie m1 = (main_Movie){.year = 2020, .ratingFn = freshness};
        so_int r1 = m1.ratingFn(m1);
        if (r1 != 50) {
            so_panic("unexpected r1");
        }
        main_Movie m2 = (main_Movie){.year = 1995, .ratingFn = freshness};
        so_int r2 = m2.ratingFn(m2);
        if (r2 != 25) {
            so_panic("unexpected r2");
        }
    }
    {
        // Function variable.
        main_RatingFn fn1 = freshness;
        main_Movie m = (main_Movie){.year = 2020};
        so_int r3 = fn1(m);
        if (r3 != 50) {
            so_panic("unexpected r3");
        }
        main_RatingFn fn2 = freshness;
        so_int r4 = fn2(m);
        if (r4 != 50) {
            so_panic("unexpected r4");
        }
    }
    {
        // Function argument.
        main_Movie m = (main_Movie){.year = 2020};
        so_int r5 = rateMovie(m, freshness);
        if (r5 != 50) {
            so_panic("unexpected r5");
        }
    }
    {
        // Function return value.
        main_Movie m = (main_Movie){.year = 2020};
        so_int r6 = getRatingFn()(m);
        if (r6 != 50) {
            so_panic("unexpected r6");
        }
    }
}
