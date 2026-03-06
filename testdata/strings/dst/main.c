#include "main.h"

// -- Implementation --

int main(void) {
    so_String str = so_str("Hi 世界!");
    // Loop over bytes.
    for (so_int i = 0; i < so_len(str); i++) {
        uint8_t chr = so_at(uint8_t, str, i);
        so_println("%s %" PRId64 " %s %u", "i =", i, "chr =", chr);
    }
    // Loop over runes.
    for (so_int i = 0; i < so_len(str);) {
        int _iw = 0;
        so_rune r = so_utf8_decode(str, i, &_iw);
        so_println("%s %" PRId64 " %s %d", "i =", i, "r =", r);
        i += _iw;
    }
    for (so_int i = 0; i < so_len(str);) {
        int _iw = 0;
        so_utf8_decode(str, i, &_iw);
        so_println("%s %" PRId64, "i =", i);
        i += _iw;
    }
    for (so_int _ = 0; _ < so_len(str);) {
        int __w = 0;
        so_rune r = so_utf8_decode(str, _, &__w);
        so_println("%s %d", "r =", r);
        _ += __w;
    }
    {
        // Compare strings.
        so_String s1 = so_str("hello");
        so_String s2 = so_str("world");
        if (so_string_eq(s1, s2) || so_string_eq(s1, so_str("hello"))) {
            so_println("%s", "ok");
        }
    }
    {
        // String conversion.
        so_String s = so_str("1世3");
        so_Slice bs = so_string_bytes(s);
        if (so_at(uint8_t, bs, 0) != '1') {
            so_panic("unexpected byte");
        }
        so_Slice rs = so_string_runes(s, so_len(s));
        if (so_at(int32_t, rs, 1) != U'世') {
            so_panic("unexpected rune");
        }
    }
}
