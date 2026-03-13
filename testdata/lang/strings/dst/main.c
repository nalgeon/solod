#include "main.h"

// -- Implementation --

int main(void) {
    {
        // String literals.
        so_String s = so_str("Hello, 世界!");
        if (so_len(s) != 7 + 3 + 3 + 1) {
            so_panic("want len(s) == 14");
        }
    }
    {
        // Loop over string bytes.
        so_String str = so_str("Hi 世界!");
        for (so_int i = 0; i < so_len(str); i++) {
            so_byte chr = so_at(so_byte, str, i);
            so_println("%s %" PRId64 " %s %u", "i =", i, "chr =", chr);
        }
    }
    {
        // Loop over string runes.
        so_String str = so_str("Hi 世界!");
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
        so_rune r = 0;
        for (so_int _ = 0; _ < so_len(str);) {
            int __w = 0;
            r = so_utf8_decode(str, _, &__w);
            (void)r;
            _ += __w;
        }
        for (so_int i = 0; i < so_len(so_str("go"));) {
            int _iw = 0;
            so_rune r = so_utf8_decode(so_str("go"), i, &_iw);
            so_println("%s %" PRId64 " %s %d", "i =", i, "r =", r);
            i += _iw;
        }
        for (so_int _i = 0; _i < so_len(str);) {
            int _iw = 0;
            so_utf8_decode(str, _i, &_iw);
            _i += _iw;
        }
    }
    {
        // Compare strings.
        so_String s1 = so_str("hello");
        so_String s2 = so_str("world");
        if (so_string_eq(s1, s2) || so_string_eq(s1, so_str("hello"))) {
            so_println("%s", "ok");
        }
    }
    // {
    // 	// String addition is not supported.
    // 	s1 := "Hello, "
    // 	s2 := "世界!"
    // 	s3 := s1 + s2
    // 	if s3 != "Hello, 世界!" {
    // 		panic("want s3 == Hello, 世界!")
    // 	}
    // }
    {
        // String conversion to byte and rune slices, and vice versa.
        so_String s1 = so_str("1世3");
        so_Slice bs = so_string_bytes(s1);
        if (so_at(so_byte, bs, 0) != '1') {
            so_panic("unexpected byte");
        }
        so_Slice rs = so_string_runes(s1, so_len(s1));
        if (so_at(so_rune, rs, 1) != U'世') {
            so_panic("unexpected rune");
        }
        so_String s2 = so_bytes_string(bs);
        if (so_string_ne(s2, s1)) {
            so_panic("want s2 == s1");
        }
        so_String s3 = so_runes_string(rs);
        if (so_string_ne(s3, s1)) {
            so_panic("want s3 == s1");
        }
        so_byte b = 'A';
        if (so_string_ne(so_byte_string(b), so_str("A"))) {
            so_panic("want string(b) == A");
        }
        so_rune r = U'世';
        if (so_string_ne(so_rune_string(r), so_str("世"))) {
            so_panic("want string(r) == 世");
        }
    }
}
