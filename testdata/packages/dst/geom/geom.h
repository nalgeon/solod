#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct geom_Rect geom_Rect;

typedef struct geom_Rect {
    double W;
    double H;
} geom_Rect;

// -- Variables and constants --
static const double geom_Pi = 3.14159;

// -- Functions and methods --
double geom_Rect_Area(geom_Rect r);
double geom_RectArea(double width, double height);
