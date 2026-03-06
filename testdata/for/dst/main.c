#include "main.h"

// -- Implementation --

int main(void) {
    so_int i = 1;
    for (; i <= 3;) {
        so_println("%" PRId64, i);
        i = i + 1;
    }
    for (so_int j = 0; j < 3; j++) {
        so_println("%" PRId64, j);
    }
    for (so_int k = 0; k < 3; k++) {
        so_println("%s %" PRId64, "range", k);
    }
    for (;;) {
        so_println("%s", "loop");
        break;
    }
    for (so_int n = 0; n < 6; n++) {
        if (n % 2 == 0) {
            continue;
        }
        so_println("%" PRId64, n);
    }
}
