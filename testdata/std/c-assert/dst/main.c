#include "main.h"

// -- Implementation --

int main(void) {
    if (assert_Enabled) {
        assert_Assert(1 + 1 == 2);
        assert_Assertf(1 + 1 == 2, "math is broken");
    } else {
        so_println("%s", "assertions are disabled");
    }
}
