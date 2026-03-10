#include "main.h"

// -- Embeds --

//go:build ignore

// begin include
#include <stdint.h>
#include "main.h"

int64_t getCSecret() {
    return C_SECRET;
}
// end include

// -- Implementation --
int64_t main_GoSecret = 42;

int main(void) {
    int64_t cSecret = getCSecret();
    if (cSecret != main_GoSecret) {
        so_panic("secret mismatch");
    }
}
