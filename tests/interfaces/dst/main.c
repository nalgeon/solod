#include "main.h"

// -- Forward declarations (functions and methods) --
static so_int calcShape(main_Shape s);
static so_int calcLine(main_Line l);
static bool shapeIsRect(main_Shape s);
static main_Rect shapeAsRect(main_Shape s);
static main_Shape rectAsShape(main_Rect* r);
static bool lineIsRect(main_Line l);
static main_Rect* lineAsRect(main_Line l);

// -- Implementation --

so_int main_Rect_Area(void* self) {
    main_Rect* r = (main_Rect*)self;
    return r->width * r->height;
}

so_int main_Rect_Perim(void* self, so_int n) {
    main_Rect* r = (main_Rect*)self;
    return n * (2 * r->width + 2 * r->height);
}

so_int main_Rect_Length(void* self) {
    main_Rect* r = (main_Rect*)self;
    return 2 * r->width + 2 * r->height;
}

static so_int calcShape(main_Shape s) {
    return s.Perim(s.self, 2) + s.Area(s.self);
}

static so_int calcLine(main_Line l) {
    return l.Length(l.self);
}

static bool shapeIsRect(main_Shape s) {
    bool ok = (s.Area == main_Rect_Area);
    return ok;
}

static main_Rect shapeAsRect(main_Shape s) {
    bool ok = (s.Area == main_Rect_Area);
    if (!ok) {
        return (main_Rect){};
    }
    main_Rect r = *((main_Rect*)s.self);
    return r;
}

static main_Shape rectAsShape(main_Rect* r) {
    return (main_Shape){.self = r, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
}

static bool lineIsRect(main_Line l) {
    bool ok = (l.Length == main_Rect_Length);
    return ok;
}

static main_Rect* lineAsRect(main_Line l) {
    bool ok = (l.Length == main_Rect_Length);
    if (!ok) {
        return NULL;
    }
    main_Rect* r = (main_Rect*)l.self;
    return r;
}

int main(void) {
    main_Rect r = (main_Rect){.width = 10, .height = 5};
    {
        // Shape interface is implemented by Rect value.
        main_Shape s = (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
        main_Shape s2 = (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
        (void)s2;
        main_Shape s3 = (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
        (void)s3;
        calcShape(s);
        // also works
        calcShape((main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim});
        // also works
        calcShape((main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim});
        (void)shapeIsRect(s);
        main_Rect rval = shapeAsRect(s);
        (void)rval;
    }
    {
        // Line interface is implemented by *Rect pointer.
        main_Line l = (main_Line){.self = &r, .Length = main_Rect_Length};
        main_Line l2 = (main_Line){.self = &r, .Length = main_Rect_Length};
        (void)l2;
        calcLine(l);
        (void)lineIsRect(l);
        main_Rect* rptr = lineAsRect(l);
        (void)rptr;
    }
    {
        // Wrap Rect value into Shape via function.
        main_Shape s = rectAsShape(&r);
        (void)s;
    }
    // Converting between interfaces (Shape to Line) is not supported.
    // s := Shape(r)
    // _, ok := s.(Line)
    // l := s.(Line)
    {
    }
}
