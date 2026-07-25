#include "main.h"

// -- Forward declarations --
static so_int freshness(main_Movie m);
static main_RatingFn getRatingFn(void);
static so_int rateMovie(main_Movie m, so_int (*f)(main_Movie));

// -- Variables and constants --

// Global function variable.
void (*main_Global1)(void) = NULL;
void (*main_Global2)(void) = NULL;
static so_int (*rate)(main_Movie) = freshness;

// -- Implementation --

static so_int freshness(main_Movie m) {
    return m.year - 1970;
}

static main_RatingFn getRatingFn(void) {
    return freshness;
}

// Anonymous function types can be passed as arguments.
static so_int rateMovie(main_Movie m, so_int (*f)(main_Movie)) {
    return f(m);
}

// Returning anonymous function types is not supported.
// func getRatingFn() func(m Movie) int {
// 	return freshness
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
        so_int (*fn1)(main_Movie) = freshness;
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
        // Anonymous function type variable.
        so_int (*fn3)(main_Movie) = freshness;
        so_int r4b = fn3(m);
        if (r4b != 50) {
            so_panic("unexpected r4b");
        }
        so_int r5 = rate(m);
        if (r5 != 50) {
            so_panic("unexpected r5");
        }
        // Function variable without an initializer is nil.
        so_int (*fn4)(main_Movie) = NULL;
        if (fn4 != NULL) {
            so_panic("unexpected fn4");
        }
        fn4 = freshness;
        if (fn4(m) != 50) {
            so_panic("unexpected fn4 result");
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
    return 0;
}
