#include "main.h"

// -- Implementation --

int main(void) {
    {
        // 2 int args.
        so_int a = 3;
        so_int b = 7;
        if (so_min(a, b) != 3) {
            so_panic("2 int args: min failed");
        }
        if (so_max(a, b) != 7) {
            so_panic("2 int args: max failed");
        }
    }
    {
        // 3 int args.
        so_int a = 5;
        so_int b = 2;
        so_int c = 8;
        if (so_min(so_min(a, b), c) != 2) {
            so_panic("3 int args: min failed");
        }
        if (so_max(so_max(a, b), c) != 8) {
            so_panic("3 int args: max failed");
        }
    }
    {
        // float64 args.
        double x = 1.5;
        double y = 2.5;
        if (so_min(x, y) != 1.5) {
            so_panic("float64 args: min failed");
        }
        if (so_max(x, y) != 2.5) {
            so_panic("float64 args: max failed");
        }
    }
    {
        // string args.
        so_String s1 = so_str("apple");
        so_String s2 = so_str("banana");
        if (so_string_ne(so_string_min(s1, s2), so_str("apple"))) {
            so_panic("string args: min failed");
        }
        if (so_string_ne(so_string_max(s1, s2), so_str("banana"))) {
            so_panic("string args: max failed");
        }
    }
}
