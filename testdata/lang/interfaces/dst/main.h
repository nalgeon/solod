#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_Shape {
    void* self;
    so_int (*Area)(void* self);
    so_int (*Perim)(void* self, so_int n);
} main_Shape;

typedef struct main_Line {
    void* self;
    so_int (*Length)(void* self);
} main_Line;

typedef struct main_Rect {
    so_int width;
    so_int height;
} main_Rect;

// -- Functions and methods --
so_int main_Rect_Area(void* self);
so_int main_Rect_Perim(void* self, so_int n);
so_int main_Rect_Length(void* self);
