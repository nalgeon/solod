#include "main.h"

// -- Implementation --

int main(void) {
    {
        // Integer arithmetics.
        so_int a = 11, b = 22, c = 33;
        so_int d = b / a + (a - c) * a + c % b;
        d += 10;
        d -= 10;
        d *= 10;
        d /= 2;
        d %= 5;
        d++;
        d--;
        (void)d;
    }
    {
        // Floating-point arithmetics.
        double x = 1.1, y = 2.2, z = 3.3;
        double f = x / y + (y - z) * x;
        f += 1.0;
        f -= 1.0;
        f *= 2.0;
        f /= 2.0;
        f++;
        f--;
        (void)f;
    }
    // {
    // 	// String concatenation is not supported.
    // 	var s1, s2, s3 string = "hello", " ", "world"
    // 	s4 := s1 + s2 + s3
    // 	_ = s4
    // }
    {
        // Bitwise operations.
        so_int b1 = 0b1010, b2 = 0b1100;
        so_int b3 = ((b1 | b2) & (b1 & b2)) | (b1 ^ b2);
        b3 = b3 << 2;
        b3 = b3 >> 1;
        b3 <<= 2;
        b3 >>= 1;
        b3 = b3 & ~b1;
        (void)b3;
        so_int b4 = 0b1010;
        b4 |= 0b1100;
        b4 &= 0b1100;
        b4 ^= 0b1100;
        // b4 &^= 0b1010 // not supported
        so_int b5 = ~b4;
        (void)b5;
    }
    {
        // Logical operations.
        bool a = true, b = false, c = true;
        bool d = ((a && b) || (b || c)) && !a;
        (void)d;
    }
    {
        // Number comparison.
        so_int x = 10, y = 20, z = 30;
        bool e1 = ((x < y) && (y > z)) || (x == z);
        (void)e1;
        bool e2 = ((x <= y) && (y >= z)) || (x != z);
        (void)e2;
    }
    {
        // Byte comparison.
        so_byte b1 = 'a', b2 = 'b', b3 = 'c';
        bool e1 = ((b1 < b2) && (b2 > b3)) || (b1 == b3);
        (void)e1;
        bool e2 = ((b1 <= b2) && (b2 >= b3)) || (b1 != b3);
        (void)e2;
    }
    {
        // Rune comparison.
        so_rune r1 = U'a', r2 = U'b', r3 = U'本';
        bool e1 = ((r1 < r2) && (r2 > r3)) || (r1 == r3);
        (void)e1;
        bool e2 = ((r1 <= r2) && (r2 >= r3)) || (r1 != r3);
        (void)e2;
    }
    {
        // String comparison.
        so_String s1 = so_str("hello"), s2 = so_str("world"), s3 = so_str("hello");
        bool e1 = ((so_string_lt(s1, s2)) || (so_string_gt(s1, s3))) && ((so_string_eq(s1, s3)) || (so_string_ne(s2, s3)));
        (void)e1;
        bool e2 = ((so_string_lte(s1, s2)) && (so_string_gte(s1, s3))) || (so_string_ne(s1, s3));
        (void)e2;
    }
}
