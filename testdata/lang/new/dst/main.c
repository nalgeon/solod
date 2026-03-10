#include "main.h"

// -- Forward declarations (types) --
typedef struct point point;

// -- Implementation --

typedef struct point {
    so_int x;
    so_int y;
} point;

int main(void) {
    {
        // new with type
        so_int* n = &(so_int){0};
        if (n == NULL || *n != 0) {
            so_panic("expected n == 0");
        }
        point* p = &(point){0};
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
}
