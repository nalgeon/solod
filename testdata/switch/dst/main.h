#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_Square main_Square;

typedef struct main_Shape {
    void* self;
    so_int (*Area)(void* self);
} main_Shape;

typedef struct main_Square {
    so_int side;
} main_Square;

// -- Functions and methods --
so_int main_Square_Area(void* self);
