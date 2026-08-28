#include "main.h"

// -- Implementation --

int main(void) {
    so_int i = 1;
    for (; i <= 3;) {
        so_println("%" PRIdINT, i);
        i = i + 1;
    }
    for (so_int j = 0; j < 3; j++) {
        so_println("%" PRIdINT, j);
    }
    so_int start = 5;
    for (start--; start >= 0; start--) {
        if (start == 2) {
            break;
        }
    }
    for (start = 5; start >= 0; start--) {
    }
    for (so_int k = 0; k < 3; k++) {
        so_println("%s %" PRIdINT, "range", k);
    }
    // The loop assigns to k2, so it keeps the value of the last iteration.
    so_int k2 = 0;
    for (so_int _k2i = 0; _k2i < 3; _k2i++) {
        k2 = _k2i;
    }
    if (k2 != 2) {
        so_panic("want k2 == 2");
    }
    for (so_int _i = 0; _i < 3; _i++) {
    }
    for (;;) {
        so_println("%s", "loop");
        break;
    }
    for (so_int n = 0; n < 6; n++) {
        if (n % 2 == 0) {
            continue;
        }
        so_println("%" PRIdINT, n);
    }
    return 0;
}
