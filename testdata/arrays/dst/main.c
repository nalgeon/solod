#include "main.h"

// -- Forward declarations (types) --
typedef so_int array[3];
typedef struct box box;

// -- Forward declarations (functions and methods) --
static void change(so_int a[3]);
static box newBox(void);

// -- Implementation --
typedef so_int array[3];

typedef struct box {
    so_int nums[3];
} box;

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
    }
}
