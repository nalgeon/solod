#include "main.h"

// -- Forward declarations --
static so_R_i64_err lenInt64(so_Slice buf);
static so_R_i64_err lenInt64Impl(so_Slice buf);
static so_int sumSlice(so_Slice s);
static void modifySlice(so_Slice s);
static so_int sumVariadic(so_Slice nums);
static so_int main_SliceHolder_sum(main_SliceHolder h);
static so_int main_SliceHolder_get(main_SliceHolder h, so_int i);

// -- Implementation --

static so_R_i64_err lenInt64(so_Slice buf) {
    so_R_i64_err _res1 = lenInt64Impl(buf);
    int64_t n = _res1.val;
    return (so_R_i64_err){.val = n, .err = NULL};
}

static so_R_i64_err lenInt64Impl(so_Slice buf) {
    return (so_R_i64_err){.val = (int64_t)(so_len(buf)), .err = NULL};
}

static so_int sumSlice(so_Slice s) {
    so_int total = 0;
    for (so_int _ = 0; _ < so_len(s); _++) {
        so_int v = so_at(so_int, s, _);
        total += v;
    }
    return total;
}

static void modifySlice(so_Slice s) {
    so_at(so_int, s, 0) = 99;
    so_at(so_int, s, 1) = 88;
}

static so_int sumVariadic(so_Slice nums) {
    so_int total = 0;
    for (so_int _ = 0; _ < so_len(nums); _++) {
        so_int n = so_at(so_int, nums, _);
        total += n;
    }
    return total;
}

static so_int main_SliceHolder_sum(main_SliceHolder h) {
    so_int s = 0;
    for (so_int _ = 0; _ < so_len(h.nums); _++) {
        so_int v = so_at(so_int, h.nums, _);
        s += v;
    }
    return s;
}

static so_int main_SliceHolder_get(main_SliceHolder h, so_int i) {
    return so_at(so_int, h.nums, i);
}

int main(void) {
    {
        // Slicing an array: all forms.
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
        // Slicing a slice: all forms.
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
        // Three-index slice expression.
        so_Slice nums = (so_Slice){(so_int[5]){1, 2, 3, 4, 5}, 5, 5};
        so_Slice s = so_slice3(so_int, nums, 1, 3, 4);
        if (so_len(s) != 2 || so_cap(s) != 3) {
            so_panic("want len 2, cap 3");
        }
        if (so_at(so_int, s, 0) != 2 || so_at(so_int, s, 1) != 3) {
            so_panic("want s[0] == 2 && s[1] == 3");
        }
    }
    {
        // Slice literals.
        so_Slice nils = (so_Slice){0};
        if (nils.ptr != NULL) {
            so_panic("want nils == nil");
        }
        if (so_len(nils) != 0) {
            so_panic("want len(nils) == 0");
        }
        so_Slice empty = (so_Slice){0};
        if (so_len(empty) != 0) {
            so_panic("want len(empty) == 0");
        }
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
        // Make a slice: len only.
        so_Slice s = so_make_slice(so_int, 4, 4);
        so_at(so_int, s, 0) = 1;
        so_at(so_int, s, 1) = 2;
        so_at(so_int, s, 2) = 3;
        so_at(so_int, s, 3) = 4;
        if (so_at(so_int, s, 0) != 1 || so_at(so_int, s, 3) != 4) {
            so_panic("want s[0]==1, s[3]==4");
        }
        if (so_len(s) != 4) {
            so_panic("want len==4");
        }
    }
    {
        // Make a slice: len and cap.
        so_Slice s = so_make_slice(so_int, 0, 8);
        if (so_len(s) != 0 || so_cap(s) != 8) {
            so_panic("want len==0, cap==8");
        }
        s = so_append(so_int, s, 10);
        if (so_len(s) != 1 || so_at(so_int, s, 0) != 10) {
            so_panic("want len==1, s[0]==10");
        }
        if (so_cap(s) != 8) {
            so_panic("want cap still 8");
        }
    }
    {
        // Make with string element type.
        so_Slice s = so_make_slice(so_String, 3, 3);
        so_at(so_String, s, 0) = so_str("hello");
        so_at(so_String, s, 1) = so_str("world");
        so_at(so_String, s, 2) = so_str("!");
        if (so_string_ne(so_at(so_String, s, 0), so_str("hello")) || so_string_ne(so_at(so_String, s, 2), so_str("!"))) {
            so_panic("want make string slice");
        }
    }
    {
        // Append: single value.
        so_Slice s = so_make_slice(so_int, 0, 4);
        s = so_append(so_int, s, 1);
        s = so_append(so_int, s, 2);
        if (so_len(s) != 2 || so_at(so_int, s, 0) != 1 || so_at(so_int, s, 1) != 2) {
            so_panic("want append single");
        }
    }
    {
        // Append: multiple values.
        so_Slice s = so_make_slice(so_int, 0, 8);
        s = so_append(so_int, s, 1, 2, 3);
        if (so_len(s) != 3 || so_at(so_int, s, 0) != 1 || so_at(so_int, s, 2) != 3) {
            so_panic("want append multi");
        }
    }
    {
        // Append: spread another slice.
        so_Slice s = so_make_slice(so_int, 0, 8);
        s = so_append(so_int, s, 1, 2);
        so_Slice other = (so_Slice){(so_int[3]){3, 4, 5}, 3, 3};
        s = so_extend(so_int, s, (other));
        if (so_len(s) != 5 || so_at(so_int, s, 2) != 3 || so_at(so_int, s, 4) != 5) {
            so_panic("want append spread");
        }
    }
    {
        // Append: string to byte slice.
        so_Slice b = so_make_slice(so_byte, 0, 8);
        so_String s = so_str("hello");
        b = so_extend(so_byte, b, so_string_bytes(s));
        if (so_len(b) != 5) {
            so_panic("len(b) != 5");
        }
        if (so_string_ne(so_bytes_string(b), so_str("hello"))) {
            so_panic("string(b) != hello");
        }
    }
    {
        // Append: strings.
        so_Slice s = so_make_slice(so_String, 0, 4);
        s = so_append(so_String, s, so_str("hello"));
        s = so_append(so_String, s, so_str("world"));
        if (so_len(s) != 2 || so_string_ne(so_at(so_String, s, 0), so_str("hello")) || so_string_ne(so_at(so_String, s, 1), so_str("world"))) {
            so_panic("want append strings");
        }
    }
    {
        // Cap: literal slice.
        so_Slice s = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        if (so_cap(s) != 3) {
            so_panic("want cap(literal)==3");
        }
    }
    {
        // Cap: make with len only.
        so_Slice s = so_make_slice(so_int, 5, 5);
        if (so_cap(s) != 5) {
            so_panic("want cap(make)==5");
        }
    }
    {
        // Cap: make with len and cap.
        so_Slice s = so_make_slice(so_int, 2, 10);
        if (so_len(s) != 2 || so_cap(s) != 10) {
            so_panic("want len==2, cap==10");
        }
    }
    {
        // Cap: sub-slice shares capacity.
        so_Slice s = so_make_slice(so_int, 5, 10);
        so_Slice s2 = so_slice(so_int, s, 2, s.len);
        if (so_len(s2) != 3 || so_cap(s2) != 8) {
            so_panic("want sub-slice cap");
        }
    }
    {
        // Len: after append.
        so_Slice s = so_make_slice(so_int, 0, 8);
        if (so_len(s) != 0) {
            so_panic("want len==0 before append");
        }
        s = so_append(so_int, s, 1);
        if (so_len(s) != 1) {
            so_panic("want len==1 after append");
        }
        s = so_append(so_int, s, 2, 3);
        if (so_len(s) != 3) {
            so_panic("want len==3 after multi append");
        }
    }
    {
        // Len: in expression.
        so_Slice s = (so_Slice){(so_int[4]){1, 2, 3, 4}, 4, 4};
        so_int n = so_len(s) + 1;
        if (n != 5) {
            so_panic("want len in expr");
        }
    }
    {
        // Pass and return slices.
        so_byte buf[4] = {0};
        so_R_i64_err _res1 = lenInt64(so_array_slice(so_byte, buf, 0, 4, 4));
        int64_t n = _res1.val;
        if (n != 4) {
            so_panic("want 4");
        }
        so_R_i64_err _res2 = lenInt64((so_Slice){(so_byte[3]){1, 2, 3}, 3, 3});
        n = _res2.val;
        if (n != 3) {
            so_panic("want 3");
        }
    }
    {
        // Pass slice to function: reads.
        so_Slice s = (so_Slice){(so_int[3]){10, 20, 30}, 3, 3};
        if (sumSlice(s) != 60) {
            so_panic("want sumSlice==60");
        }
    }
    {
        // Pass slice to function: modification (reference semantics).
        so_Slice s = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        modifySlice(s);
        if (so_at(so_int, s, 0) != 99 || so_at(so_int, s, 1) != 88) {
            so_panic("want modified slice");
        }
    }
    {
        // Variadic function: individual args.
        if (sumVariadic((so_Slice){(so_int[3]){1, 2, 3}, 3, 3}) != 6) {
            so_panic("want variadic sum==6");
        }
    }
    {
        // Variadic function: spread slice.
        so_Slice nums = (so_Slice){(so_int[3]){10, 20, 30}, 3, 3};
        if (sumVariadic(nums) != 60) {
            so_panic("want variadic spread==60");
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
        if (so_at(so_int, s, 1) != 9) {
            so_panic("want 9");
        }
    }
    {
        // Slice element in comparison.
        so_Slice s = (so_Slice){(so_int[3]){10, 20, 30}, 3, 3};
        if (so_at(so_int, s, 0) > so_at(so_int, s, 1)) {
            so_panic("want s[0] <= s[1]");
        }
        if (so_at(so_int, s, 2) < so_at(so_int, s, 1)) {
            so_panic("want s[2] >= s[1]");
        }
        if (so_at(so_int, s, 0) == so_at(so_int, s, 1)) {
            so_panic("want s[0] != s[1]");
        }
    }
    {
        // Slice element in arithmetic expression.
        so_Slice s = (so_Slice){(so_int[3]){2, 3, 5}, 3, 3};
        so_int result = so_at(so_int, s, 0) * so_at(so_int, s, 1) + so_at(so_int, s, 2);
        if (result != 11) {
            so_panic("want 2*3+5==11");
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
        // Copy: return value (partial copy when dst is smaller).
        so_Slice src = (so_Slice){(so_int[5]){1, 2, 3, 4, 5}, 5, 5};
        so_Slice dst = so_make_slice(so_int, 3, 3);
        so_int n = so_copy(so_int, dst, src);
        if (n != 3) {
            so_panic("want copy returned 3");
        }
        if (so_at(so_int, dst, 0) != 1 || so_at(so_int, dst, 2) != 3) {
            so_panic("want partial copy values");
        }
    }
    {
        // Copy: return value (partial copy when src is smaller).
        so_Slice src = (so_Slice){(so_int[2]){1, 2}, 2, 2};
        so_Slice dst = so_make_slice(so_int, 5, 5);
        so_int n = so_copy(so_int, dst, src);
        if (n != 2) {
            so_panic("want copy returned 2");
        }
        if (so_at(so_int, dst, 0) != 1 || so_at(so_int, dst, 1) != 2 || so_at(so_int, dst, 2) != 0) {
            so_panic("want partial copy src smaller");
        }
    }
    {
        // Copying a string to a byte slice.
        so_String str = so_str("hello");
        so_Slice b = so_make_slice(so_byte, so_len(str), so_len(str));
        so_copy_string(b, str);
        if (so_string_ne(so_bytes_string(b), so_str("hello"))) {
            so_panic("want string(b) == 'hello'");
        }
        // Copying a string literal to a byte slice.
        so_Slice b2 = so_make_slice(so_byte, 2, 2);
        so_copy_string(b2, so_str("ab"));
        if (so_string_ne(so_bytes_string(b2), so_str("ab"))) {
            so_panic("want string(b2) == 'ab'");
        }
    }
    {
        // Element types: byte.
        so_Slice s = (so_Slice){(so_byte[3]){0x41, 0x42, 0x43}, 3, 3};
        if (so_at(so_byte, s, 0) != 0x41 || so_at(so_byte, s, 2) != 0x43) {
            so_panic("want byte slice");
        }
    }
    {
        // Element types: bool.
        so_Slice s = (so_Slice){(bool[3]){true, false, true}, 3, 3};
        if (!so_at(bool, s, 0) || so_at(bool, s, 1) || !so_at(bool, s, 2)) {
            so_panic("want bool slice");
        }
    }
    {
        // Element types: float64.
        so_Slice s = (so_Slice){(double[3]){1.5, 2.5, 3.5}, 3, 3};
        double sum = so_at(double, s, 0) + so_at(double, s, 1) + so_at(double, s, 2);
        if (sum != 7.5) {
            so_panic("want float64 sum==7.5");
        }
    }
    {
        // Element types: rune.
        so_Slice s = (so_Slice){(so_rune[3]){U'a', U'b', U'c'}, 3, 3};
        if (so_at(so_rune, s, 0) != U'a' || so_at(so_rune, s, 2) != U'c') {
            so_panic("want rune slice");
        }
    }
    {
        // Element types: struct.
        so_Slice s = (so_Slice){(main_Pair[2]){(main_Pair){1, 2}, (main_Pair){3, 4}}, 2, 2};
        if (so_at(main_Pair, s, 0).x != 1 || so_at(main_Pair, s, 1).y != 4) {
            so_panic("want struct slice");
        }
        so_at(main_Pair, s, 0).x = 10;
        if (so_at(main_Pair, s, 0).x != 10) {
            so_panic("want modified struct field");
        }
    }
    {
        // Element types: pointer.
        so_int a = 42;
        so_int b = 99;
        so_Slice s = (so_Slice){(so_int*[2]){&a, &b}, 2, 2};
        if (*so_at(so_int*, s, 0) != 42 || *so_at(so_int*, s, 1) != 99) {
            so_panic("want pointer slice");
        }
        *so_at(so_int*, s, 0) = 100;
        if (a != 100) {
            so_panic("want modified through pointer");
        }
    }
    {
        // Element types: string.
        so_Slice s = (so_Slice){(so_String[3]){so_str("hello"), so_str("world"), so_str("!")}, 3, 3};
        if (so_string_ne(so_at(so_String, s, 0), so_str("hello")) || so_string_ne(so_at(so_String, s, 2), so_str("!"))) {
            so_panic("want string slice values");
        }
        if (so_len(s) != 3) {
            so_panic("want string slice len==3");
        }
    }
    {
        // 2D slice: access and modify.
        so_Slice twoD = (so_Slice){(so_Slice[2]){(so_Slice){(so_int[3]){1, 2, 3}, 3, 3}, (so_Slice){(so_int[3]){4, 5, 6}, 3, 3}}, 2, 2};
        if (so_at(so_int, so_at(so_Slice, twoD, 0), 0) != 1 || so_at(so_int, so_at(so_Slice, twoD, 1), 2) != 6) {
            so_panic("want 2D values");
        }
        so_at(so_int, so_at(so_Slice, twoD, 0), 1) = 20;
        if (so_at(so_int, so_at(so_Slice, twoD, 0), 1) != 20) {
            so_panic("want 2D modified");
        }
    }
    {
        // Nil slice: comparison.
        so_Slice s = {0};
        if (s.ptr != NULL) {
            so_panic("want nil slice");
        }
        s = (so_Slice){(so_int[1]){1}, 1, 1};
        if (s.ptr == NULL) {
            so_panic("want non-nil slice");
        }
    }
    {
        // Nil slice: len and cap.
        so_Slice s = {0};
        if (so_len(s) != 0) {
            so_panic("want nil len==0");
        }
        if (so_cap(s) != 0) {
            so_panic("want nil cap==0");
        }
    }
    {
        // Slice assigned to another variable (shared backing).
        so_Slice s1 = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        so_Slice s2 = s1;
        so_at(so_int, s2, 0) = 99;
        if (so_at(so_int, s1, 0) != 99) {
            so_panic("want shared backing");
        }
    }
    {
        // Struct with slice field.
        main_SliceHolder h = (main_SliceHolder){.nums = (so_Slice){(so_int[3]){10, 20, 30}, 3, 3}};
        if (main_SliceHolder_get(h, 0) != 10 || main_SliceHolder_get(h, 2) != 30) {
            so_panic("want struct slice field get");
        }
        if (main_SliceHolder_sum(h) != 60) {
            so_panic("want struct slice field sum");
        }
    }
    {
        // Named slice type: literal.
        main_IntSlice s = (so_Slice){(so_int[3]){10, 20, 30}, 3, 3};
        if (so_at(so_int, s, 0) != 10 || so_at(so_int, s, 2) != 30) {
            so_panic("want named type literal");
        }
        if (so_len(s) != 3) {
            so_panic("want named type len");
        }
    }
    {
        // Named slice type: make.
        main_IntSlice s = so_make_slice(so_int, 0, 4);
        s = so_append(so_int, s, 1, 2);
        if (so_len(s) != 2 || so_at(so_int, s, 0) != 1 || so_at(so_int, s, 1) != 2) {
            so_panic("want named type make+append");
        }
    }
    {
        // Named slice type: range.
        main_IntSlice s = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        so_int sum = 0;
        for (so_int _ = 0; _ < so_len(s); _++) {
            so_int v = so_at(so_int, s, _);
            sum += v;
        }
        if (sum != 6) {
            so_panic("want named type range");
        }
    }
    {
        // For-range over slices: index only.
        so_Slice s = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        so_int sum = 0;
        for (so_int i = 0; i < so_len(s); i++) {
            sum += so_at(so_int, s, i);
        }
        if (sum != 6) {
            so_panic("want sum == 6");
        }
    }
    {
        // For-range over slices: value only.
        so_Slice s = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        so_int sum = 0;
        for (so_int _ = 0; _ < so_len(s); _++) {
            so_int num = so_at(so_int, s, _);
            sum += num;
        }
        if (sum != 6) {
            so_panic("want sum == 6");
        }
    }
    {
        // For-range over slices: index and value.
        so_Slice s = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        so_int sum = 0;
        for (so_int i = 0; i < so_len(s); i++) {
            so_int num = so_at(so_int, s, i);
            (void)i;
            sum += num;
        }
        if (sum != 6) {
            so_panic("want sum == 6");
        }
    }
    {
        // For-range: empty body.
        so_Slice s = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        for (so_int _i = 0; _i < so_len(s); _i++) {
        }
    }
    {
        // For-range: assign (not define).
        so_Slice s = (so_Slice){(so_int[3]){10, 20, 30}, 3, 3};
        so_int i = 0;
        so_int v = 0;
        so_int sum = 0;
        for (i = 0; i < so_len(s); i++) {
            v = so_at(so_int, s, i);
            sum += i + v;
        }
        // (0+10) + (1+20) + (2+30) = 63
        if (sum != 63) {
            so_panic("want range assign sum==63");
        }
    }
    {
        // For-range over string slice.
        so_Slice s = (so_Slice){(so_String[3]){so_str("a"), so_str("b"), so_str("c")}, 3, 3};
        so_String result = so_str("");
        for (so_int _ = 0; _ < so_len(s); _++) {
            so_String v = so_at(so_String, s, _);
            result = so_string_add(result, v);
        }
        if (so_string_ne(result, so_str("abc"))) {
            so_panic("want range string concat");
        }
    }
    {
        // For-range over struct slice.
        so_Slice s = (so_Slice){(main_Pair[3]){(main_Pair){1, 2}, (main_Pair){3, 4}, (main_Pair){5, 6}}, 3, 3};
        so_int sum = 0;
        for (so_int _ = 0; _ < so_len(s); _++) {
            main_Pair p = so_at(main_Pair, s, _);
            sum += p.x + p.y;
        }
        if (sum != 21) {
            so_panic("want range struct sum==21");
        }
    }
    {
        // Slice from array: modification affects the array.
        so_int arr[3] = {1, 2, 3};
        so_Slice s = so_array_slice(so_int, arr, 0, 3, 3);
        so_at(so_int, s, 0) = 99;
        if (arr[0] != 99) {
            so_panic("want array modified via slice");
        }
    }
    {
        // Sub-slice: modification affects the original.
        so_Slice s = (so_Slice){(so_int[5]){1, 2, 3, 4, 5}, 5, 5};
        so_Slice sub = so_slice(so_int, s, 1, 4);
        so_at(so_int, sub, 0) = 99;
        if (so_at(so_int, s, 1) != 99) {
            so_panic("want original modified via sub-slice");
        }
    }
    {
        // Append after sub-slice.
        so_Slice s = so_make_slice(so_int, 3, 6);
        so_at(so_int, s, 0) = 1;
        so_at(so_int, s, 1) = 2;
        so_at(so_int, s, 2) = 3;
        s = so_append(so_int, s, 4);
        if (so_len(s) != 4 || so_at(so_int, s, 3) != 4) {
            so_panic("want append after make");
        }
    }
    {
        // Make with cap, fill with append.
        so_Slice s = so_make_slice(so_int, 0, 5);
        s = so_append(so_int, s, 10);
        s = so_append(so_int, s, 20, 30);
        s = so_append(so_int, s, 40, 50);
        if (so_len(s) != 5 || so_cap(s) != 5) {
            so_panic("want filled to cap");
        }
        if (so_at(so_int, s, 0) != 10 || so_at(so_int, s, 4) != 50) {
            so_panic("want filled values");
        }
    }
    {
        // Make with byte slice: zero-initialized.
        so_Slice s = so_make_slice(so_byte, 4, 4);
        if (so_at(so_byte, s, 0) != 0 || so_at(so_byte, s, 3) != 0) {
            so_panic("want byte zero init");
        }
        so_at(so_byte, s, 0) = 0xFF;
        if (so_at(so_byte, s, 0) != 0xFF) {
            so_panic("want byte set");
        }
    }
    {
        // Slice in if-init statement.
        so_Slice s = (so_Slice){(so_int[3]){10, 20, 30}, 3, 3};
        {
            so_int v = so_at(so_int, s, 1);
            if (v == 20) {
                (void)v;
            } else {
                so_panic("want if-init slice");
            }
        }
    }
    {
        // Index with variable.
        so_Slice s = (so_Slice){(so_int[3]){100, 200, 300}, 3, 3};
        so_int idx = 2;
        if (so_at(so_int, s, idx) != 300) {
            so_panic("want variable index");
        }
        so_at(so_int, s, idx) = 999;
        if (so_at(so_int, s, 2) != 999) {
            so_panic("want variable index set");
        }
    }
    {
        // Index with expression.
        so_Slice s = (so_Slice){(so_int[3]){100, 200, 300}, 3, 3};
        if (so_at(so_int, s, so_len(s) - 1) != 300) {
            so_panic("want expr index");
        }
    }
    {
        // Len and cap in comparison.
        so_Slice s = so_make_slice(so_int, 3, 8);
        if (so_len(s) >= so_cap(s)) {
            so_panic("want len < cap");
        }
        if (so_cap(s) != 8) {
            so_panic("want cap==8");
        }
    }
    {
        // Clear slice.
        so_Slice s = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        so_clear(so_int, s);
        if (so_at(so_int, s, 0) != 0 || so_at(so_int, s, 1) != 0 || so_at(so_int, s, 2) != 0) {
            so_panic("want zeroed after clear");
        }
        if (so_len(s) != 3) {
            so_panic("want len preserved after clear");
        }
    }
}
