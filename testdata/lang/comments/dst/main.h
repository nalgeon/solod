#pragma once
#include "so/builtin/builtin.h"

// -- Types --

// Point represents a 2D coordinate.
typedef struct main_Point {
    so_int x;
    so_int y;
} main_Point;

// -- Variables and constants --

// MaxCoord is the maximum coordinate value.
extern const so_int main_MaxCoord;

// -- Functions and methods --

// NewPoint creates a new Point.
main_Point main_NewPoint(so_int x, so_int y);

// Scale multiplies both coordinates by a factor.
void main_Point_Scale(void* self, so_int factor);
