#include "main.h"

// -- Implementation --

int main(void) {
    // if-else
    if (7 % 2 == 0) {
        so_panic("want 7%2 != 0");
    } else {
        so_println("%s", "7 is odd");
    }
    // if without else
    if (8 % 2 == 0 || 7 % 2 == 0) {
        so_println("%s", "either 8 or 7 are even");
    }
    // if with a complex condition
    if (1 == 2 - 1 && (2 == 1 + 1 || 3 == 6 / 2) && !(4 != 2 * 2)) {
        so_println("%s", "all conditions are true");
    }
    // if-elseif-else
    if (9 % 3 == 0) {
        so_println("%s", "9 is divisible by 3");
    } else if (9 % 2 == 0) {
        so_panic("want 9%2 != 0");
    } else {
        so_panic("want 9%3 == 0");
    }
    // if with init
    {
        so_int num = 9;
        if (num < 0) {
            so_panic("want num >= 0");
        } else if (num < 10) {
            so_println("%" PRIdINT " %s", num, "has 1 digit");
        } else {
            so_panic("want 0 <= num < 10");
        }
    }
    // else-if init
    so_int n = 0;
    if (n == 1) {
        so_panic("want n == 0");
    } else {
        so_int m = n + 1;
        if (m == 1) {
            so_println("%s", "m == 1");
        } else {
            so_panic("want m == 1");
        }
    }
    // else-if init that shadows an outer variable
    so_int v = 100;
    if (v == 1) {
        so_panic("want v == 100");
    } else {
        so_int v = n + 1;
        if (v == 1) {
            so_println("%s", "shadowed v == 1");
        } else {
            so_panic("want shadowed v == 1");
        }
    }
    if (v != 100) {
        so_panic("want outer v == 100");
    }
    // chained else-if with init
    {
        so_int k = 1;
        if (k == 0) {
            so_panic("want k == 1");
        } else {
            so_int j = k + 1;
            if (j == 0) {
                so_panic("want j == 2");
            } else {
                so_int i = j + 1;
                if (i == 3) {
                    so_println("%s", "i == 3");
                } else {
                    so_panic("want i == 3");
                }
            }
        }
    }
    return 0;
}
