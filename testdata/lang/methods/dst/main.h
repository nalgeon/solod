#pragma once
#include "so/builtin/builtin.h"

// -- Types --

// Methods on struct types.
typedef struct main_Rect {
    so_int width;
    so_int height;
} main_Rect;

// Methods on primitive types are also supported.
typedef so_int main_HttpStatus;

// -- Functions and methods --
so_int main_Rect_Area(void* self);
so_String main_HttpStatus_String(main_HttpStatus s);
