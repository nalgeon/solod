#include "main.h"

// -- Types --

typedef struct counter counter;

typedef struct counter {
    so_int* val;
} counter;

// -- Implementation --

int main(void) {
    {
        // Integer arithmetics.
        so_int a = 11, b = 22, c = 33;
        so_int d = so_div(b, a) + (a - c) * a + so_mod(c, b);
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
        // Division by -1 overflows for the minimum value.
        int32_t i32 = -2147483648;
        int32_t neg = -1;
        if (so_div(i32, neg) != -2147483648 || so_mod(i32, neg) != 0) {
            so_panic("expected i32 == -2147483648 && rem == 0");
        }
        if (so_div(i32, -1) != -2147483648 || so_mod(i32, -1) != 0) {
            so_panic("expected the same for a constant divisor");
        }
        int64_t i64 = INT64_MIN;
        if (so_div(i64, -1) != INT64_MIN || so_mod(i64, -1) != 0) {
            so_panic("expected i64 == -9223372036854775808 && rem == 0");
        }
        int8_t i8 = -128;
        if ((int8_t)(so_div(i8, -1)) != -128 || so_mod(i8, -1) != 0) {
            so_panic("expected i8 == -128 && rem == 0");
        }
        // An unsigned divisor never reaches the guard, because -1
        // converts to the maximum value of the type.
        uint32_t u1 = 7, u2 = 4294967295;
        if (so_div(u1, u2) != 0 || so_mod(u1, u2) != 7) {
            so_panic("expected u1/u2 == 0 && u1%u2 == 7");
        }
        int32_t q = (int32_t)(-2147483648);
        q = so_div(q, -1);
        if (q != -2147483648) {
            so_panic("expected q == -2147483648");
        }
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
    {
        // String addition is supported for string literals (but not for variables).
        so_String s = so_str("hello" " " "world");
        (void)s;
    }
    {
        // Bitwise operations.
        so_int b1 = 0b1010, b2 = 0b1100;
        so_int b3 = (((b1 | b2) & (b1 & b2)) | (b1 ^ b2));
        b3 = (b3 << 2);
        b3 = (b3 >> 1);
        b3 <<= 2;
        b3 >>= 1;
        b3 = (b3 & ~b1);
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
        // Arithmetic on a type narrower than int. C promotes the operands to
        // int and computes at that width, so every result needs a conversion
        // back to the narrow type.
        so_byte n1 = 3, n2 = 10;
        so_byte n3 = (so_byte)(n1 - n2);
        (void)n3;
        so_int n4 = (so_int)((so_byte)(n1 - n2));
        (void)n4;
        so_int n5 = (so_int)((so_byte)(n1 * n2));
        (void)n5;
        so_int n6 = (so_int)(so_byte)((n1 << 6));
        (void)n6;
        so_int n7 = (so_int)((so_byte)(~n1));
        (void)n7;
        so_int n8 = (so_int)((so_byte)(-n1));
        (void)n8;
        int16_t s1 = 30000, s2 = 30000;
        so_int n9 = (so_int)((int16_t)(s1 + s2));
        (void)n9;
    }
    {
        // Increment/decrement through a pointer.
        so_int n = 1;
        so_int* p = &n;
        (*p)++;
        (*p)--;
        counter c = (counter){.val = &n};
        (*c.val)++;
        (*c.val)--;
        (void)n;
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
        so_rune r1 = 'a', r2 = 'b', r3 = 0x672c;
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
    return 0;
}
