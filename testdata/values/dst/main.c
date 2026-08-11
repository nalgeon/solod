#include "main.h"

// -- Implementation --

int main(void) {
    {
        // Integer literals.
        const so_unused int64_t d1 = 123;
        (void)d1;
        const so_unused int64_t d2 = 100000;
        (void)d2;
        const so_unused int64_t d3 = 0b1010;
        (void)d3;
        const so_unused int64_t d4 = 0600;
        (void)d4;
        const so_unused int64_t d5 = 0xBadFace;
        (void)d5;
        const so_unused int64_t d6 = 0x677a2fcc40c6;
        (void)d6;
    }
    {
        // Integer literals above MaxInt64 are unsigned in C.
        uint64_t u1 = 18446744073709551615u;
        if (u1 + 1 != 0 || u1 / 2 != 9223372036854775807) {
            so_panic("want MaxUint64");
        }
        // The target decides the width of uint, so the value is derived,
        // not written as a literal.
        so_uint u2 = ~(so_uint)(0) / 2 + 1;
        if (u2 - 1 != ~(so_uint)(0) / 2) {
            so_panic("want MaxInt+1");
        }
        const so_unused uint64_t u3 = 0xFFFFFFFFFFFFFFFFu;
        if (u3 != u1) {
            so_panic("want MaxUint64");
        }
    }
    {
        // Floating-point literals.
        const so_unused double f1 = 3.14;
        (void)f1;
        const so_unused double f2 = 0.25;
        (void)f2;
        const so_unused double f3 = 1e-9;
        (void)f3;
        const so_unused double f4 = 6.022e23;
        (void)f4;
        const so_unused double f5 = 1e6;
        (void)f5;
    }
    // {
    // 	// Imaginary literals - not supported.
    // 	const i1 = 0i
    // 	_ = i1
    // 	const i2 = 0o123i // == 0o123 * 1i == 83i
    // 	_ = i2
    // 	const i3 = 0xabci // == 0xabc * 1i == 2748i
    // 	_ = i3
    // 	const i4 = 2.71828i
    // 	_ = i4
    // 	const i5 = 1.e+0i
    // }
    {
        // Rune literals.
        const so_unused so_rune r1 = 'a';
        (void)r1;
        const so_unused so_rune r2 = 0xe4;
        (void)r2;
        const so_unused so_rune r3 = 0x672c;
        (void)r3;
        const so_unused so_rune r4 = 0xff;
        (void)r4;
        const so_unused so_rune r5 = 0x12e4;
        (void)r5;
    }
    {
        // String literals.
        const so_unused so_String s1 = so_str("abc");
        (void)s1;
        const so_unused so_String s2 = so_str("abc\n\t\tdef");
        (void)s2;
        const so_unused so_String s3 = so_str("\n");
        (void)s3;
        const so_unused so_String s4 = so_str("日本語");
        (void)s4;
        const so_unused so_String s5 = so_str("\377ÿ");
        (void)s5;
    }
    {
        // Escapes that C reads differently than Go.
        so_String s1 = so_str("a\377b");
        if (so_len(s1) != 3 || so_at(so_byte, s1, 1) != 0xff || so_at(so_byte, s1, 2) != 'b') {
            so_panic("want 3 bytes");
        }
        so_String s2 = so_str("\na");
        if (so_len(s2) != 2 || so_at(so_byte, s2, 0) != '\n' || so_at(so_byte, s2, 1) != 'a') {
            so_panic("want newline and a");
        }
        so_String s3 = so_str("AA");
        if (so_len(s3) != 2 || so_at(so_byte, s3, 0) != 'A' || so_at(so_byte, s3, 1) != 'A') {
            so_panic("want AA");
        }
        // trigraph
        so_String s4 = so_str("?\?!");
        if (so_len(s4) != 3 || so_at(so_byte, s4, 0) != '?' || so_at(so_byte, s4, 1) != '?' || so_at(so_byte, s4, 2) != '!') {
            so_panic("want ?\?!");
        }
        // NUL byte
        so_String s5 = so_str("a\000b");
        if (so_len(s5) != 3 || so_at(so_byte, s5, 0) != 'a' || so_at(so_byte, s5, 1) != 0 || so_at(so_byte, s5, 2) != 'b') {
            so_panic("want a, nul, b");
        }
        so_String s6 = so_str("\\x41");
        if (so_len(s6) != 4 || so_at(so_byte, s6, 0) != '\\') {
            so_panic("want 4 bytes");
        }
        so_rune r = 'A';
        if (r != 'A') {
            so_panic("want A");
        }
        so_byte b = 0xe9;
        if (b != 0xe9) {
            so_panic("want 0xe9");
        }
    }
    {
        // Conversions.
        const so_unused so_uint x = 123;
        const so_unused so_int n1 = (so_int)(x);
        (void)n1;
        const so_unused so_int n2 = (so_int)(x & 7);
        (void)n2;
        const so_unused int64_t mask2 = 0b00011111;
        so_byte p0 = 'x';
        so_rune r = (so_rune)(p0 & mask2);
        (void)r;
    }
    return 0;
}
