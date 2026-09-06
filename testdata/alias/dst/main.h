#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_Rect main_Rect;
typedef struct main_Frame main_Frame;

typedef struct main_Rect {
    so_int width;
    so_int height;
} main_Rect;

typedef struct main_Shape {
    void* self;
    so_int (*Area)(void* self);
} main_Shape;

static inline so_int main_Shape_Area(main_Shape self) {
    return self.Area(self.self);
}

// RectPtr aliases a pointer to a named type.
typedef main_Rect* main_RectPtr;

// Canvas aliases a named interface.
typedef main_Shape main_Canvas;

typedef struct main_Frame {
    main_Rect* rect;
} main_Frame;

// -- Functions and methods --
so_int main_Rect_Area(void* self);
