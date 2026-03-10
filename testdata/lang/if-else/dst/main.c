#include "main.h"

// -- Implementation --

int main(void) {
    if (7 % 2 == 0) {
        so_panic("want 7%2 != 0");
    } else {
        so_println("%s", "7 is odd");
    }
    if (8 % 2 == 0 || 7 % 2 == 0) {
        so_println("%s", "either 8 or 7 are even");
    }
    if (1 == 2 - 1 && (2 == 1 + 1 || 3 == 6 / 2) && !(4 != 2 * 2)) {
        so_println("%s", "all conditions are true");
    }
    if (9 % 3 == 0) {
        so_println("%s", "9 is divisible by 3");
    } else if (9 % 2 == 0) {
        so_panic("want 9%2 != 0");
    } else {
        so_panic("want 9%3 == 0");
    }
    {
        so_int num = 9;
        if (num < 0) {
            so_panic("want num >= 0");
        } else if (num < 10) {
            so_println("%" PRId64 " %s", num, "has 1 digit");
        } else {
            so_panic("want 0 <= num < 10");
        }
    }
}
