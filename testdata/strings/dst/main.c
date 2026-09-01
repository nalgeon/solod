#include "main.h"

// -- Variables and constants --

// Emitted as adjacent string literals.
static const so_unused so_String quoted = so_str("\"" "ok" "\"");
static const so_unused bool equal = true;
static so_unused bool greater = true;
static so_unused bool flags[2] = {true, false};

// -- Implementation --

int main(void) {
    {
        // Empty string.
        so_String s1 = so_str("");
        if (so_len(s1) != 0 || so_string_ne(s1, so_str(""))) {
            so_panic("want empty string");
        }
        so_String s2 = so_str("");
        if (so_len(s2) != 0 || so_string_ne(s2, so_str(""))) {
            so_panic("want empty string");
        }
    }
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
            so_println("%s %" PRIdINT " %s %u", "i =", i, "chr =", chr);
        }
    }
    {
        // Loop over string runes.
        so_String str = so_str("Hi 世界!");
        for (so_int i = 0, _iw = 0; i < so_len(str); i += _iw) {
            _iw = 0;
            so_rune r = so_utf8_decode(str, i, &_iw);
            so_println("%s %" PRIdINT " %s %d", "i =", i, "r =", r);
        }
        for (so_int i = 0, _iw = 0; i < so_len(str); i += _iw) {
            _iw = 0;
            so_utf8_decode(str, i, &_iw);
            so_println("%s %" PRIdINT, "i =", i);
        }
        for (so_int _ = 0, __w = 0; _ < so_len(str); _ += __w) {
            __w = 0;
            so_rune r = so_utf8_decode(str, _, &__w);
            so_println("%s %d", "r =", r);
        }
        so_rune r = 0;
        for (so_int _ = 0, __w = 0; _ < so_len(str); _ += __w) {
            __w = 0;
            r = so_utf8_decode(str, _, &__w);
            (void)r;
        }
        for (so_int i = 0, _iw = 0; i < so_len(so_str("go")); i += _iw) {
            _iw = 0;
            so_rune r = so_utf8_decode(so_str("go"), i, &_iw);
            so_println("%s %" PRIdINT " %s %d", "i =", i, "r =", r);
        }
        for (so_int _i = 0, _iw = 0; _i < so_len(str); _i += _iw) {
            _iw = 0;
            so_utf8_decode(str, _i, &_iw);
        }
    }
    {
        // Loop over string runes with a declared key. The loop assigns to the
        // key, so the key keeps the index of the last rune.
        so_String s = so_str("go世");
        so_int i = 0;
        for (so_int _ii = 0, _iw = 0; _ii < so_len(s); _ii += _iw) {
            i = _ii;
            _iw = 0;
            so_utf8_decode(s, _ii, &_iw);
            (void)i;
        }
        if (i != 2) {
            so_panic("want i == 2");
        }
        so_rune r = 0;
        for (so_int _ii = 0, _iw = 0; _ii < so_len(s); _ii += _iw) {
            i = _ii;
            _iw = 0;
            r = so_utf8_decode(s, _ii, &_iw);
            (void)r;
        }
        if (i != 2 || r != 0x4e16) {
            so_panic("want i == 2 && r == 世");
        }
    }
    {
        // Continue in range-over-string loop.
        so_String s = so_str("hello");
        so_int n = 0;
        for (so_int _ = 0, __w = 0; _ < so_len(s); _ += __w) {
            __w = 0;
            so_rune c = so_utf8_decode(s, _, &__w);
            if (c == 'l') {
                continue;
            }
            n++;
        }
        if (n != 3) {
            so_panic("want n == 3");
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
    {
        // Compare constant strings.
        const so_unused bool localLess = true;
        if (!main_Less || !equal || !greater || !localLess) {
            so_panic("want true");
        }
        if (!flags[0] || flags[1]) {
            so_panic("want flags == [true, false]");
        }
    }
    {
        // String addition.
        so_String s1 = so_str("Hello, ");
        so_String s2 = so_str("世界!");
        so_String s3 = so_string_add(s1, s2);
        if (so_string_ne(s3, so_str("Hello, 世界!"))) {
            so_panic("want s3 == Hello, 世界!");
        }
    }
    {
        // String conversion to byte and rune slices, and vice versa.
        so_String s1 = so_str("1世3");
        so_Slice bs = so_string_bytes(s1);
        if (so_at(so_byte, bs, 0) != '1') {
            so_panic("unexpected byte");
        }
        so_Slice rs = so_string_runes(s1);
        if (so_at(so_rune, rs, 1) != 0x4e16) {
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
        if (so_string_ne(so_int_string(b), so_str("A"))) {
            so_panic("want string(b) == A");
        }
        so_rune r = 0x4e16;
        if (so_string_ne(so_int_string(r), so_str("世"))) {
            so_panic("want string(r) == 世");
        }
        so_byte b2 = 200;
        if (so_string_ne(so_int_string(b2), so_str("È"))) {
            so_panic("want string(b2) == È");
        }
        so_int n = 0x4e16;
        if (so_string_ne(so_int_string(n), so_str("世"))) {
            so_panic("want string(n) == 世");
        }
        uint64_t u = 0x10001f4a9;
        if (so_string_ne(so_int_string(u), so_str("�"))) {
            so_panic("want string(u) == replacement char");
        }
        const so_unused so_String c = so_str("�");
        if (false) {
            so_panic("want c == replacement char");
        }
        if (false) {
            so_panic("want quoted == \"ok\"");
        }
    }
    {
        // String conversion to slices of named byte and rune types.
        so_String s1 = so_str("1世3");
        so_Slice bs = so_string_bytes(s1);
        if (so_at(main_Byte, bs, 0) != '1') {
            so_panic("unexpected byte");
        }
        so_Slice rs = so_string_runes(s1);
        if (so_at(main_Rune, rs, 1) != 0x4e16) {
            so_panic("unexpected rune");
        }
        if (so_string_ne(so_bytes_string(bs), s1)) {
            so_panic("want string(bs) == s1");
        }
        if (so_string_ne(so_runes_string(rs), s1)) {
            so_panic("want string(rs) == s1");
        }
    }
    {
        // Conversion between string types.
        so_String s = so_str("1世3");
        if (so_string_ne(s, s)) {
            so_panic("want string(s) == s");
        }
        if (so_string_ne(so_string_slice(s, 1, s.len), so_str("世3"))) {
            so_panic("want string(s)[1:] == 世3");
        }
        so_println("%.*s", so_string_slice(s, 0, 1).len, so_string_slice(s, 0, 1).ptr);
        main_Name n = s;
        if (so_string_ne(n, s)) {
            so_panic("want string(n) == s");
        }
        so_println("%.*s", so_string_slice(n, 0, 1).len, so_string_slice(n, 0, 1).ptr);
    }
    return 0;
}
