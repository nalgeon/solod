#include "main.h"

// -- Forward declarations (functions and methods) --
static so_Result lenInt64(so_Slice buf);
static so_Result lenInt64Impl(so_Slice buf);

// -- Implementation --

static so_Result lenInt64(so_Slice buf) {
    so_Result _res1 = lenInt64Impl(buf);
    int64_t n = _res1.val.as_int;
    return (so_Result){.val.as_int = n, .err = NULL};
}

static so_Result lenInt64Impl(so_Slice buf) {
    return (so_Result){.val.as_int = (int64_t)(so_len(buf)), .err = NULL};
}

int main(void) {
    {
        // Slicing an array.
        so_int nums[5] = {1, 2, 3, 4, 5};
        so_Slice s1 = so_array_slice(so_int, nums, 0, 5, 5);
        so_at(so_int, s1, 1) = 200;
        (void)s1;
        so_Slice s2 = so_array_slice(so_int, nums, 2, 5, 5);
        (void)s2;
        so_Slice s3 = so_array_slice(so_int, nums, 0, 3, 5);
        (void)s3;
        so_Slice s4 = so_array_slice(so_int, nums, 1, 4, 5);
        (void)s4;
        // n == 3
        so_int n = so_copy(so_int, s4, s1);
        (void)n;
    }
    {
        // Slicing a string.
        so_String str = so_str("hello");
        so_String s1 = so_string_slice(str, 0, str.len);
        if (so_string_ne(s1, so_str("hello"))) {
            so_panic("want s1 == hello");
        }
        so_String s2 = so_string_slice(str, 2, str.len);
        if (so_string_ne(s2, so_str("llo"))) {
            so_panic("want s2 == llo");
        }
        so_String s3 = so_string_slice(str, 0, 3);
        if (so_string_ne(s3, so_str("hel"))) {
            so_panic("want s3 == hel");
        }
        so_String s4 = so_string_slice(str, 1, 4);
        if (so_string_ne(s4, so_str("ell"))) {
            so_panic("want s4 == ell");
        }
    }
    {
        // Slicing a slice.
        so_Slice nums = (so_Slice){(so_int[5]){1, 2, 3, 4, 5}, 5, 5};
        so_Slice s1 = so_slice(so_int, nums, 0, nums.len);
        if (so_at(so_int, s1, 0) != 1 || so_at(so_int, s1, 4) != 5) {
            so_panic("want s1[0] == 1 && s1[4] == 5");
        }
        so_Slice s2 = so_slice(so_int, nums, 2, nums.len);
        if (so_at(so_int, s2, 0) != 3 || so_at(so_int, s2, 2) != 5) {
            so_panic("want s2[0] == 3 && s2[2] == 5");
        }
        so_Slice s3 = so_slice(so_int, nums, 0, 3);
        if (so_at(so_int, s3, 0) != 1 || so_at(so_int, s3, 2) != 3) {
            so_panic("want s3[0] == 1 && s3[2] == 3");
        }
        so_Slice s4 = so_slice(so_int, nums, 1, 4);
        if (so_at(so_int, s4, 0) != 2 || so_at(so_int, s4, 2) != 4) {
            so_panic("want s4[0] == 2 && s4[2] == 4");
        }
    }
    {
        // Slice literals.
        so_Slice empty = (so_Slice){0};
        (void)empty;
        so_Slice strSlice = (so_Slice){(so_String[3]){so_str("a"), so_str("b"), so_str("c")}, 3, 3};
        // sLen == 3
        so_int sLen = so_len(strSlice);
        (void)sLen;
        so_Slice twoD = (so_Slice){(so_Slice[2]){(so_Slice){(so_int[3]){1, 2, 3}, 3, 3}, (so_Slice){(so_int[3]){4, 5, 6}, 3, 3}}, 2, 2};
        // x == 2
        so_int x = so_at(so_int, so_at(so_Slice, twoD, 0), 1);
        (void)x;
    }
    {
        // Make a slice.
        so_Slice s = so_make_slice(so_int, 4, 4);
        so_at(so_int, s, 0) = 1;
        so_at(so_int, s, 1) = 2;
        so_at(so_int, s, 2) = 3;
        so_at(so_int, s, 3) = 4;
        (void)s;
    }
    {
        // Pass and return slices.
        so_byte buf[4] = {0};
        so_Result _res1 = lenInt64(so_array_slice(so_byte, buf, 0, 4, 4));
        int64_t n = _res1.val.as_int;
        if (n != 4) {
            so_panic("want 4");
        }
        so_Result _res2 = lenInt64((so_Slice){(so_byte[3]){1, 2, 3}, 3, 3});
        n = _res2.val.as_int;
        if (n != 3) {
            so_panic("want 3");
        }
    }
    {
        // Number operations on slice elements.
        so_Slice s = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        so_at(so_int, s, 1) += 10;
        so_at(so_int, s, 1) -= 10;
        so_at(so_int, s, 1) *= 10;
        so_at(so_int, s, 1) /= 2;
        so_at(so_int, s, 1) %= 6;
        so_at(so_int, s, 1)++;
        so_at(so_int, s, 1)--;
        if (so_at(so_int, s, 1) != 4) {
            so_panic("want 4");
        }
    }
    {
        // Bitwise operations on slice elements.
        so_Slice s = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        so_at(so_int, s, 1) <<= 2;
        so_at(so_int, s, 1) >>= 1;
        so_at(so_int, s, 1) |= 0b1100;
        so_at(so_int, s, 1) &= 0b1111;
        so_at(so_int, s, 1) ^= 0b0101;
        // s[1] &^= 0b1010  // not supported
        if (so_at(so_int, s, 1) != 9) {
            so_panic("want 9");
        }
    }
    {
        // Copying a slice.
        so_Slice s = so_make_slice(so_String, 3, 6);
        so_at(so_String, s, 0) = so_str("a");
        so_at(so_String, s, 1) = so_str("b");
        so_at(so_String, s, 2) = so_str("c");
        so_Slice c = so_make_slice(so_String, so_len(s), so_len(s));
        so_copy(so_String, c, s);
        if (so_string_ne(so_at(so_String, c, 0), so_str("a")) || so_string_ne(so_at(so_String, c, 2), so_str("c"))) {
            so_panic("want c[0] == 'a' && c[2] == 'c'");
        }
    }
    {
        // For-range over slices.
        so_Slice s = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        so_int sum = 0;
        for (so_int i = 0; i < so_len(s); i++) {
            sum += so_at(so_int, s, i);
        }
        if (sum != 6) {
            so_panic("want sum == 6");
        }
        sum = 0;
        for (so_int _ = 0; _ < so_len(s); _++) {
            so_int num = so_at(so_int, s, _);
            sum += num;
        }
        if (sum != 6) {
            so_panic("want sum == 6");
        }
        sum = 0;
        for (so_int i = 0; i < so_len(s); i++) {
            so_int num = so_at(so_int, s, i);
            (void)i;
            sum += num;
        }
        if (sum != 6) {
            so_panic("want sum == 6");
        }
        for (so_int _i = 0; _i < so_len(s); _i++) {
        }
    }
}
