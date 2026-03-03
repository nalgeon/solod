#include "main.h"

// -- Forward declarations (functions and methods) --
static so_int freshness(main_Movie m);

// -- Implementation --

static so_int freshness(main_Movie m) {
    return m.year - 1970;
}

int main(void) {
    main_Movie m1 = (main_Movie){.year = 2020, .ratingFn = freshness};
    // 50
    so_int s1 = m1.ratingFn(m1);
    if (s1 != 50) {
        so_panic("unexpected s1");
    }
    main_Movie m2 = (main_Movie){.year = 1995, .ratingFn = freshness};
    // 25
    so_int s2 = m2.ratingFn(m2);
    if (s2 != 25) {
        so_panic("unexpected s2");
    }
}
