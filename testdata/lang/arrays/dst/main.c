#include "main.h"

// -- Types --
typedef so_int array[3];

typedef struct box {
    so_int nums[3];
} box;

typedef struct arange {
    uint8_t lo;
    uint8_t hi;
} arange;

// -- Variables and constants --
static arange aranges[16] = {[0] = (arange){0x10, 0x20}, [1] = (arange){0x30, 0x40}, [2] = (arange){0x50, 0x60}};

// -- Forward declarations --
static void change(so_int a[3]);
static box newBox(void);

// -- Implementation --

static void change(so_int a[3]) {
    a[0] = 42;
}

static box newBox(void) {
    return (box){.nums = {11, 22, 33}};
}

int main(void) {
    {
        // Array literals.
        so_int a[5] = {0};
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
        // Array length is fixed and part of the type.
        so_int a[3] = {1, 2, 3};
        if (3 != 3) {
            so_panic("want len(a) == 3");
        }
        (void)a;
        so_int b[3] = {1, 2, 3};
        if (so_array_ne(b, a, 3 * sizeof(so_int))) {
            so_panic("want b == a");
        }
        so_int c[3] = {3, 2, 1};
        if (so_array_eq(c, a, 3 * sizeof(so_int))) {
            so_panic("want c != a");
        }
        if (so_array_ne(c, ((so_int[3]){3, 2, 1}), 3 * sizeof(so_int))) {
            so_panic("want c == {3, 2, 1}");
        }
    }
    {
        // Arrays decay to pointers when passed to functions.
        so_int a[3] = {1, 2, 3};
        change(a);
        if (a[0] != 42) {
            so_panic("want a[0] == 42");
        }
    }
    {
        // Arrays can be struct fields.
        box b = newBox();
        if (b.nums[1] != 22) {
            so_panic("want b.nums[1] == 22");
        }
    }
    {
        // Array-to-array assignment.
        so_int a[3] = {1, 2, 3};
        so_int b[3] = {0, 0, 0};
        memcpy(b, a, sizeof(b));
        if (b[0] != 1 || b[2] != 3) {
            so_panic("want b == {1, 2, 3}");
        }
        so_int c[3] = {0};
        memcpy(c, (so_int[3]){1, 2, 3}, sizeof(c));
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
        array a = {0};
        a[1] = 42;
        if (a[1] != 42) {
            so_panic("want a[1] == 42");
        }
    }
    {
        // Array pointers.
        so_int a[3] = {1, 2, 3};
        so_int (*p)[3] = &a;
        if (so_array_ne((*p), a, 3 * sizeof(so_int))) {
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
        const so_int n = 3;
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
        int32_t twoD[2][3] = {0};
        for (so_int i = 0; i < 2; i++) {
            for (so_int j = 0; j < 3; j++) {
                twoD[i][j] = (int32_t)(i * 10 + j + 1);
            }
        }
        if (twoD[0][0] != 1 || twoD[1][2] != 13) {
            so_panic("want twoD == {{1, 2, 3}, {11, 12, 13}}");
        }
        memcpy(twoD, (int32_t[2][3]){{1, 2, 3}, {11, 12, 13}}, sizeof(twoD));
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
        so_int b[3] = {0};
        b[0] = 1;
        b[1] = 2;
        b[2] = 3;
        if (so_array_ne(a, b, 3 * sizeof(so_int))) {
            so_panic("want a == b");
        }
        so_int c[3] = {3, 2, 1};
        if (so_array_eq(a, c, 3 * sizeof(so_int))) {
            so_panic("want a != c");
        }
    }
    (void)aranges;
}
