#include "main.h"

// -- Types --

typedef struct box box;
typedef struct arange arange;
typedef so_int array[3];

typedef struct box {
    so_int nums[3];
} box;

typedef struct arange {
    uint8_t lo;
    uint8_t hi;
} arange;

// -- Forward declarations --
static void change(so_int a[3]);
static so_int at(so_int a[3], so_int i);
static so_int* reverse(so_int a[3]);
static box newBox(void);
static so_int box_sum(box b, so_int a[3]);
static so_int sum3(so_int a[3]);

// -- Variables and constants --
static so_unused arange aranges[16] = {[0] = (arange){0x10, 0x20}, [1] = (arange){0x30, 0x40}, [2] = (arange){0x50, 0x60}};

// -- Implementation --

static void change(so_int a[3]) {
    a[0] = 42;
}

static so_int at(so_int a[3], so_int i) {
    return a[i];
}

static so_int* reverse(so_int a[3]) {
    so_int tmp = a[0];
    a[0] = a[2];
    a[2] = tmp;
    return a;
}

static box newBox(void) {
    return (box){.nums = {11, 22, 33}};
}

static so_int box_sum(box b, so_int a[3]) {
    so_int total = 0;
    for (so_int i = 0; i < 3; i++) {
        so_int v = a[i];
        total += b.nums[i] + v;
    }
    return total;
}

int main(void) {
    {
        // Array literals.
        so_int a[5] = {};
        (void)a;
        a[4] = 100;
        so_int x = a[4];
        (void)x;
        so_int l = 5;
        (void)l;
        so_int b[5] = {1, 2, 3, 4, 5};
        (void)b;
        so_int c[5] = {1, 2, 3, 4, 5};
        (void)c;
        so_int d[5] = {100, [3] = 400, 500};
        (void)d;
    }
    {
        // Multi-variable array declaration.
        so_byte a1[2] = {};
        so_byte a2[2] = {};
        (void)a1;
        (void)a2;
        so_byte b1[2] = {'1', '2'};
        so_byte b2[2] = {'3', '4'};
        (void)b1;
        (void)b2;
        so_byte c1[2] = {'1', '2'};
        so_byte c2[2] = {'3', '4'};
        (void)c1;
        (void)c2;
        so_byte d1[2] = {'1', '2'};
        so_byte d2[2] = {'3', '4'};
        (void)d1;
        (void)d2;
    }
    {
        // Array length is fixed and part of the type.
        so_int a[3] = {1, 2, 3};
        if (3 != 3) {
            so_panic("want len(a) == 3");
        }
        (void)a;
        so_int b[3] = {1, 2, 3};
        if (so_mem_ne(b, a, 3 * sizeof(so_int))) {
            so_panic("want b == a");
        }
        so_int c[3] = {3, 2, 1};
        if (so_mem_eq(c, a, 3 * sizeof(so_int))) {
            so_panic("want c != a");
        }
        if (so_mem_ne(c, ((so_int[3]){3, 2, 1}), 3 * sizeof(so_int))) {
            so_panic("want c == {3, 2, 1}");
        }
    }
    {
        // Passing arrays to functions.
        so_int a[3] = {1, 2, 3};
        change(a);
        if (a[0] != 42) {
            so_panic("want a[0] == 42");
        }
        so_int v1 = at((so_int[3]){11, 22, 33}, 1);
        if (v1 != 22) {
            so_panic("want at([11, 22, 33], 1) == 22");
        }
    }
    {
        // Passing array literals to methods.
        box b = newBox();
        so_int total = box_sum(b, (so_int[3]){11, 22, 33});
        if (total != 66 * 2) {
            so_panic("want b.sum([11, 22, 33]) == 66*2");
        }
    }
    {
        // Returning arrays from functions.
        so_int a[3] = {1, 2, 3};
        memcpy(a, reverse(a), sizeof(so_int[3]));
        if (a[0] != 3 || a[1] != 2 || a[2] != 1) {
            so_panic("want reverse({1, 2, 3}) == {3, 2, 1}");
        }
    }
    {
        // Arrays can be struct fields.
        box b1 = newBox();
        if (b1.nums[1] != 22) {
            so_panic("want b1.nums[1] == 22");
        }
        box b2 = {};
        memcpy(b2.nums, (so_int[3]){1, 2, 3}, sizeof(so_int[3]));
        if (b2.nums[1] != 2) {
            so_panic("want b2.nums[1] == 2");
        }
        box b3 = {};
        so_int arr[3] = {1, 2, 3};
        memcpy(b3.nums, arr, sizeof(so_int[3]));
        if (b3.nums[1] != 2) {
            so_panic("want b3.nums[1] == 2");
        }
    }
    {
        // Array-to-array assignment.
        so_int a[3] = {1, 2, 3};
        so_int b[3] = {0, 0, 0};
        memcpy(b, a, sizeof(so_int[3]));
        if (b[0] != 1 || b[2] != 3) {
            so_panic("want b == {1, 2, 3}");
        }
        so_int c[3] = {};
        memcpy(c, (so_int[3]){1, 2, 3}, sizeof(so_int[3]));
        if (c[0] != 1 || c[2] != 3) {
            so_panic("want c == {1, 2, 3}");
        }
        so_int d[3];
        memcpy(d, c, sizeof(d));
        if (d[0] != 1 || d[2] != 3) {
            so_panic("want d == {1, 2, 3}");
        }
    }
    {
        // Arrays can be named types.
        array a = {};
        a[1] = 42;
        if (a[1] != 42) {
            so_panic("want a[1] == 42");
        }
    }
    {
        // Array pointers.
        so_int a[3] = {1, 2, 3};
        so_int (*p)[3] = &a;
        if (so_mem_ne((*p), a, 3 * sizeof(so_int))) {
            so_panic("want p == a");
        }
        if ((*p)[1] != 2) {
            so_panic("want p[1] == 2");
        }
    }
    {
        // Array pointer slicing.
        so_int a[5] = {1, 2, 3, 4, 5};
        so_int (*p)[5] = &a;
        so_Slice s = so_array_slice(so_int, (*p), 1, 4, 5);
        if (so_len(s) != 3 || so_at(so_int, s, 0) != 2 || so_at(so_int, s, 2) != 4) {
            so_panic("want p[1:4] == {2, 3, 4}");
        }
    }
    {
        // Array pointer len, range.
        so_int a[3] = {10, 20, 30};
        so_int (*p)[3] = &a;
        if (3 != 3) {
            so_panic("want len(p) == 3");
        }
        so_int sum = 0;
        for (so_int _ = 0; _ < 3; _++) {
            so_int v = (*p)[_];
            sum += v;
        }
        if (sum != 60) {
            so_panic("want sum == 60");
        }
    }
    {
        // Variable-length arrays are not possible, because
        // Go's type checker resolves n to a constant.
        const so_unused int64_t n = 3;
        (void)n;
        so_int a[3] = {};
        if (a[0] != 0 || a[1] != 0 || a[2] != 0) {
            so_panic("want a == {0, 0, 0}");
        }
        a[0] = 42;
        if (a[0] != 42) {
            so_panic("want a[0] == 42");
        }
    }
    {
        // Multi-dimensional arrays.
        int32_t twoD[2][3] = {};
        for (so_int i = 0; i < 2; i++) {
            for (so_int j = 0; j < 3; j++) {
                twoD[i][j] = (int32_t)(i * 10 + j + 1);
            }
        }
        if (twoD[0][0] != 1 || twoD[1][2] != 13) {
            so_panic("want twoD == {{1, 2, 3}, {11, 12, 13}}");
        }
        memcpy(twoD, (int32_t[2][3]){{1, 2, 3}, {11, 12, 13}}, sizeof(int32_t[2][3]));
        if (twoD[0][0] != 1 || twoD[1][2] != 13) {
            so_panic("want twoD == {{1, 2, 3}, {11, 12, 13}}");
        }
    }
    {
        // For-range over arrays.
        so_int a[3] = {1, 2, 3};
        so_int sum = 0;
        for (so_int i = 0; i < 3; i++) {
            sum += a[i];
        }
        if (sum != 6) {
            so_panic("want sum == 6");
        }
        sum = 0;
        for (so_int _ = 0; _ < 3; _++) {
            so_int num = a[_];
            sum += num;
        }
        if (sum != 6) {
            so_panic("want sum == 6");
        }
        // The loop assigns to i and num, so both
        // keep the values of the last iteration.
        so_int i = 0;
        so_int num = 0;
        for (so_int _ii = 0; _ii < 3; _ii++) {
            i = _ii;
            num = a[_ii];
        }
        if (i != 2 || num != 3) {
            so_panic("want i == 2 && num == 3");
        }
        sum = 0;
        for (so_int i = 0; i < 3; i++) {
            so_int num = a[i];
            (void)i;
            sum += num;
        }
        if (sum != 6) {
            so_panic("want sum == 6");
        }
        for (so_int _i = 0; _i < 3; _i++) {
        }
    }
    {
        // Array comparisons.
        so_int a[3] = {1, 2, 3};
        so_int b[3] = {};
        b[0] = 1;
        b[1] = 2;
        b[2] = 3;
        if (so_mem_ne(a, b, 3 * sizeof(so_int))) {
            so_panic("want a == b");
        }
        so_int c[3] = {3, 2, 1};
        if (so_mem_eq(a, c, 3 * sizeof(so_int))) {
            so_panic("want a != c");
        }
    }
    {
        // Slice-to-array conversion.
        so_Slice s = (so_Slice){(so_int[3]){11, 22, 33}, 3, 3};
        so_int a[3];
        memcpy(a, so_slice_array(s, 3), sizeof(a));
        if (a[0] != 11 || a[1] != 22 || a[2] != 33) {
            so_panic("want a == {11, 22, 33}");
        }
        so_int v1 = at(so_slice_array(s, 3), 1);
        if (v1 != 22) {
            so_panic("want at([11, 22, 33], 1) == 22");
        }
    }
    {
        // Slice-to-array-pointer conversion. The pointer aliases
        // the slice data, so a write through it changes the slice.
        so_Slice s = (so_Slice){(so_int[3]){11, 22, 33}, 3, 3};
        so_int (*p)[3] = (so_int(*)[3])so_slice_array(s, 3);
        if ((*p)[0] != 11 || (*p)[2] != 33) {
            so_panic("want p == &{11, 22, 33}");
        }
        memcpy(*p, (so_int[3]){1, 2, 3}, sizeof(so_int[3]));
        if (so_at(so_int, s, 0) != 1 || so_at(so_int, s, 2) != 3) {
            so_panic("want s == {1, 2, 3}");
        }
        // A sub-slice converts to a shorter array pointer.
        so_int (*q)[2] = (so_int(*)[2])so_slice_array(so_slice(so_int, s, 1, s.len), 2);
        (*q)[0] = 42;
        if (so_at(so_int, s, 1) != 42) {
            so_panic("want s[1] == 42");
        }
        // Assignment through an unnamed array pointer.
        memcpy(*(so_int(*)[1])so_slice_array(so_slice(so_int, s, 2, s.len), 1), (so_int[1]){7}, sizeof(so_int[1]));
        if (so_at(so_int, s, 2) != 7) {
            so_panic("want s[2] == 7");
        }
    }
    {
        // Array copy in a var declaration.
        so_int src[3] = {1, 2, 3};
        so_int d1[3];
        memcpy(d1, src, sizeof(d1));
        so_int d2[3];
        memcpy(d2, src, sizeof(d2));
        so_int d3[3];
        memcpy(d3, src, sizeof(d3));
        d1[0] = 99;
        d2[0] = 98;
        if (src[0] != 1 || d1[0] != 99 || d2[0] != 98 || d3[0] != 1) {
            so_panic("want an independent copy");
        }
        box b = {};
        memcpy(b.nums, src, sizeof(so_int[3]));
        so_int field[3];
        memcpy(field, b.nums, sizeof(field));
        if (field[2] != 3) {
            so_panic("want field == {1, 2, 3}");
        }
    }
    {
        // Parentheses around an array literal.
        so_int a[3] = {1, 2, 3};
        so_int b[3] = {};
        memcpy(b, (so_int[3]){4, 5, 6}, sizeof(so_int[3]));
        if (a[0] != 1 || b[2] != 6) {
            so_panic("want a == {1, 2, 3} and b == {4, 5, 6}");
        }
        if (!(so_mem_eq(a, ((so_int[3]){1, 2, 3}), 3 * sizeof(so_int)))) {
            so_panic("want a == {1, 2, 3}");
        }
        if (sum3((so_int[3]){1, 2, 3}) != 6) {
            so_panic("want sum3 == 6");
        }
        box c = (box){.nums = {7, 8, 9}};
        if (c.nums[1] != 8) {
            so_panic("want c.nums == {7, 8, 9}");
        }
    }
    (void)aranges;
    return 0;
}

// sum3 adds the elements of a three-element array.
static so_int sum3(so_int a[3]) {
    return a[0] + a[1] + a[2];
}
