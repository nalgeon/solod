#include "main.h"

// -- Implementation --

int main(void) {
    double a1 = geom_RectArea(5.0, 10.0);
    (void)a1;
    (void)geom_Pi;
    double a2 = geom_RectArea(5.0, 10.0);
    (void)a2;
    (void)geom_Pi;
    geom_Rect r = {};
    (void)geom_Rect_Area(r);
    return 0;
}
