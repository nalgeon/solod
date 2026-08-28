#include "main.h"

// -- Types --

typedef struct point point;

typedef struct point {
    so_int x;
    so_int y;
} point;
typedef so_int buf[4];

// -- Implementation --

int main(void) {
    {
        // new with type
        so_int* n = &(so_int){};
        if (n == NULL || *n != 0) {
            so_panic("expected n == 0");
        }
        point* p = &(point){};
        if (p == NULL || p->x != 0 || p->y != 0) {
            so_panic("expected p.x == 0 && p.y == 0");
        }
    }
    {
        // new with value
        so_int* n = &(so_int){42};
        if (n == NULL || *n != 42) {
            so_panic("expected n == 42");
        }
        point* p1 = &(point){1, 2};
        if (p1 == NULL || p1->x != 1 || p1->y != 2) {
            so_panic("expected p1.x == 1 && p1.y == 2");
        }
        point pval = (point){3, 4};
        (void)pval;
        point* p2 = &pval;
        if (p2 == NULL || p2->x != 3 || p2->y != 4) {
            so_panic("expected p2.x == 3 && p2.y == 4");
        }
    }
    {
        // new with an array type
        so_int (*a)[3] = &(so_int[3]){};
        if (3 != 3 || (*a)[0] != 0 || (*a)[2] != 0) {
            so_panic("expected a == [0 0 0]");
        }
        (*a)[2] = 42;
        if ((*a)[2] != 42) {
            so_panic("expected a[2] == 42");
        }
        buf* b = &(buf){};
        if (4 != 4 || (*b)[3] != 0) {
            so_panic("expected b == [0 0 0 0]");
        }
        so_int (*m)[2][3] = &(so_int[2][3]){};
        if (2 != 2 || 3 != 3 || (*m)[1][2] != 0) {
            so_panic("expected m == [[0 0 0] [0 0 0]]");
        }
    }
    {
        // new with an array value
        so_int (*a)[3] = &(so_int[3]){1, 2, 3};
        if ((*a)[0] != 1 || (*a)[2] != 3) {
            so_panic("expected a == [1 2 3]");
        }
        buf* b = &(so_int[4]){1, 2, 3, 4};
        if ((*b)[3] != 4) {
            so_panic("expected b[3] == 4");
        }
        so_int aval[3] = {5, 6, 7};
        so_int (*c)[3] = &aval;
        if ((*c)[1] != 6) {
            so_panic("expected c[1] == 6");
        }
        (*c)[1] = 8;
        if (aval[1] != 8) {
            so_panic("expected aval[1] == 8");
        }
    }
    return 0;
}
