#include "main.h"

// -- Implementation --

int main(void) {
    double x = math_Sqrt(49);
    if (x != 7) {
        so_panic("want x == 7");
    }
}
