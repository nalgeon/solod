#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_Canvas main_Canvas;
typedef struct main_Rect main_Rect;

typedef struct main_Shape {
    void* self;
    so_int (*Area)(void* self);
    so_int (*Perim)(void* self, so_int);
} main_Shape;

static inline so_int main_Shape_Area(main_Shape self) {
    return self.Area(self.self);
}

static inline so_int main_Shape_Perim(main_Shape self, so_int n) {
    return self.Perim(self.self, n);
}

// Painter declares its methods with unnamed and blank parameters.
typedef struct main_Painter {
    void* self;
    so_int (*Fill)(void* self, so_int);
    so_int (*Paint)(void* self, so_int, so_String);
} main_Painter;

static inline so_int main_Painter_Fill(main_Painter self, so_int _0) {
    return self.Fill(self.self, _0);
}

static inline so_int main_Painter_Paint(main_Painter self, so_int _0, so_String _1) {
    return self.Paint(self.self, _0, _1);
}

typedef struct main_Canvas {
    so_String name;
    main_Shape shape;
} main_Canvas;

typedef struct main_Rect {
    so_int width;
    so_int height;
} main_Rect;

// -- Functions and methods --
so_int main_Rect_Area(void* self);
so_int main_Rect_Perim(void* self, so_int n);
so_int main_Rect_Paint(void* self, so_int n, so_String _2 so_unused);
so_int main_Rect_Fill(void* self, so_int n);
