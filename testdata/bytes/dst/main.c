#include "main.h"

// -- Implementation --

int main(void) {
    {
        // Byte literals.
        so_byte b1 = 'a', b2 = 'b', b3 = 'c';
        if (b1 != 'a' || b2 != 'b' || b3 != 'c') {
            so_panic("unexpected byte");
        }
    }
    {
        // Rune literals.
        so_rune r1 = 0x4e16, r2 = 0x754c, r3 = '!';
        if (r1 != 0x4e16 || r2 != 0x754c || r3 != '!') {
            so_panic("unexpected rune");
        }
    }
    {
        // Byte slices and strings.
        so_Slice b = (so_Slice){(so_byte[5]){'h', 'e', 'l', 'l', 'o'}, 5, 5};
        so_String s = so_bytes_string(b);
        if (so_string_ne(s, so_str("hello"))) {
            so_panic("want s == hello");
        }
    }
    {
        // Rune slices and strings.
        so_Slice r = (so_Slice){(so_rune[2]){0x4e16, 0x754c}, 2, 2};
        so_String s = so_runes_string(r);
        if (so_string_ne(s, so_str("世界"))) {
            so_panic("want s == 世界");
        }
    }
    return 0;
}
