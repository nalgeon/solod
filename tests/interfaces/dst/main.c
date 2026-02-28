#include "main.h"

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

static so_int calc(main_Shape s) {
    return s.Perim(s.self, 2) + s.Area(s.self);
}

static bool shapeIsRect(main_Shape s) {
    bool ok = (s.Area == main_Rect_Area);
    return ok;
}

static so_int shapeAsRect(main_Shape s) {
    bool ok = (s.Area == main_Rect_Area);
    if (!ok) {
        return 0;
    }
    main_Rect r = *((main_Rect*)s.self);
    return main_Rect_Area(&r);
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

typedef struct {
    void* self;
    so_int (*Read)(void* self, so_Slice p, so_Error* err);
} reader;

int main(void) {
    main_Rect r = {.width = 10, .height = 5};
    {
        main_Shape s = (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
        calc(s);
        calc((main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim});
        calc((main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim});
        (void)shapeIsRect(s);
        so_int a = shapeAsRect(s);
        (void)a;
    }
    {
        main_Line l = (main_Line){.self = &r, .Length = main_Rect_Length};
        (void)lineIsRect(l);
        main_Rect* rptr = lineAsRect(l);
        (void)rptr;
    }
    {
        reader rdr = {0};
        (void)rdr;
    }
}
