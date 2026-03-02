#include "main.h"

// -- Forward declarations (functions and methods) --
static main_Point offset(main_Point p, so_int dx, so_int dy);

// -- Implementation --

// NewPoint creates a new Point.
main_Point main_NewPoint(so_int x, so_int y) {
    return (main_Point){.x = x, .y = y};
}

// Scale multiplies both coordinates by a factor.
void main_Point_Scale(void* self, so_int factor) {
    main_Point* p = (main_Point*)self;
    p->x = p->x * factor;
    p->y = p->y * factor;
}

// offset is unexported.
static main_Point offset(main_Point p, so_int dx, so_int dy) {
    return (main_Point){.x = p.x + dx, .y = p.y + dy};
}

// MaxCoord is the maximum coordinate value.
const so_int main_MaxCoord = 1000;

int main(void) {
    main_Point p = main_NewPoint(1, 2);
    main_Point_Scale(&p, main_MaxCoord);
    p = offset(p, 1, 1);
    (void)p;
}
