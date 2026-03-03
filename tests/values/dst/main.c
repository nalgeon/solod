#include "main.h"

// -- Implementation --

int main(void) {
    {
        // Integer literals.
        const so_int d1 = 123;
        (void)d1;
        const so_int d2 = 100000;
        (void)d2;
        const so_int d3 = 0b1010;
        (void)d3;
        const so_int d4 = 0600;
        (void)d4;
        const so_int d5 = 0xBadFace;
        (void)d5;
        const so_int d6 = 0x677a2fcc40c6;
        (void)d6;
    }
    {
        // Floating-point literals.
        const double f1 = 3.14;
        (void)f1;
        const double f2 = 0.25;
        (void)f2;
        const double f3 = 1e-9;
        (void)f3;
        const double f4 = 6.022e23;
        (void)f4;
        const double f5 = 1e6;
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
        const int32_t r1 = U'a';
        (void)r1;
        const int32_t r2 = U'ä';
        (void)r2;
        const int32_t r3 = U'本';
        (void)r3;
        const int32_t r4 = U'\xff';
        (void)r4;
        const int32_t r5 = U'\u12e4';
        (void)r5;
    }
    {
        // String literals.
        const so_String s1 = so_strlit("abc");
        (void)s1;
        const so_String s2 = so_strlit("abc\n\t\tdef");
        (void)s2;
        const so_String s3 = so_strlit("\n");
        (void)s3;
        const so_String s4 = so_strlit("日本語");
        (void)s4;
        const so_String s5 = so_strlit("\xff\u00FF");
        (void)s5;
    }
}
