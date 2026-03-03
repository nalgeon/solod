#include "main.h"

// -- Forward declarations (functions and methods) --
static so_Result copyBuf(so_Slice buf);
static so_Result copyImpl(so_Slice buf);

// -- Implementation --

static so_Result copyBuf(so_Slice buf) {
    so_Result _res1 = copyImpl(buf);
    int64_t n1 = _res1.val.as_int;
    so_Result _res2 = copyImpl((so_Slice){(uint8_t[0]){}, 0, 0});
    int64_t n2 = _res2.val.as_int;
    return (so_Result){.val.as_int = n1 + n2, .err = NULL};
}

static so_Result copyImpl(so_Slice buf) {
    return (so_Result){.val.as_int = (int64_t)10 + so_len(buf), .err = NULL};
}

int main(void) {
    {
        // Slicing an array.
        so_Slice nums = (so_Slice){(so_int[5]){1, 2, 3, 4, 5}, 5, 5};
        so_Slice s1 = so_slice(so_int, nums, 0, nums.len);
        so_index(so_int, s1, 1) = 200;
        (void)s1;
        so_Slice s2 = so_slice(so_int, nums, 2, nums.len);
        (void)s2;
        so_Slice s3 = so_slice(so_int, nums, 0, 3);
        (void)s3;
        so_Slice s4 = so_slice(so_int, nums, 1, 4);
        (void)s4;
        // n == 3
        so_int n = so_copy(so_int, s4, s1);
        (void)n;
    }
    {
        // Slice literals.
        so_Slice strSlice = (so_Slice){(so_String[3]){so_strlit("a"), so_strlit("b"), so_strlit("c")}, 3, 3};
        // sLen == 3
        so_int sLen = so_len(strSlice);
        (void)sLen;
        so_Slice twoD = (so_Slice){(so_Slice[2]){(so_Slice){(so_int[3]){1, 2, 3}, 3, 3}, (so_Slice){(so_int[3]){4, 5, 6}, 3, 3}}, 2, 2};
        // x == 2
        so_int x = so_index(so_int, so_index(so_Slice, twoD, 0), 1);
        (void)x;
    }
    {
        // Make a slice.
        so_Slice s = so_make_slice(so_int, 4, 4);
        so_index(so_int, s, 0) = 1;
        so_index(so_int, s, 1) = 2;
        so_index(so_int, s, 2) = 3;
        so_index(so_int, s, 3) = 4;
        (void)s;
    }
    {
        // Pass and return slices.
        so_Slice buf = (so_Slice){(uint8_t[4]){0}, 4, 4};
        so_Result _res1 = copyBuf(so_slice(uint8_t, buf, 0, buf.len));
        int64_t n = _res1.val.as_int;
        if (n != 24) {
            so_panic("want 24");
        }
    }
}
