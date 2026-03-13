#include "main.h"

// -- Implementation --

int main(void) {
    so_int fails = 0;
    for (so_int i = 0; i < 10; i++) {
        if (i % 2 == 0) {
            goto next;
        }
        next:
        fails++;
        if (fails > 2) {
            goto fallback;
        }
    }
    fallback:
    if (fails != 3) {
        so_panic("want fails == 3");
    }
}
