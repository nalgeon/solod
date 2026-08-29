#include "geom.h"

// -- Forward declarations --
static double rectArea(double width, double height);

// -- Implementation --

double geom_Rect_Area(geom_Rect r) {
    return rectArea(r.W, r.H);
}

static double rectArea(double width, double height) {
    return width * height;
}

double geom_RectArea(double width, double height) {
    return rectArea(width, height);
}
