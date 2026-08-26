#include "main.h"

// -- Types --

typedef struct circle circle;

typedef struct circle {
    so_int radius;
} circle;

typedef so_int (*circleValFunc)(circle);

typedef so_int (*circlePtrFunc)(circle*);

// -- Forward declarations --
static so_int main_Rect_perim(void* self, so_int n);
static so_int main_Rect_sides(main_Rect self);
static so_int circle_shape(void* self);
static main_Rect main_Rect_resize(main_Rect r, so_int x);
static so_int circle_area(void* self);
static so_int circle_perim(circle c);
static void circle_scale(void* self, so_int x);

// -- Implementation --

so_int main_Rect_Area(void* self) {
    main_Rect* r = self;
    return r->width * r->height;
}

static so_int main_Rect_perim(void* self, so_int n) {
    main_Rect* r = self;
    return n * (2 * r->width + 2 * r->height);
}

// Blank receiver.
static so_int main_Rect_sides(main_Rect self) {
    (void)self;
    return 4;
}

static so_int circle_shape(void* self) {
    (void)self;
    return 0;
}

static main_Rect main_Rect_resize(main_Rect r, so_int x) {
    r.height *= x;
    r.width *= x;
    return r;
}

static so_int circle_area(void* self) {
    circle* c = self;
    return 3 * c->radius * c->radius;
}

static so_int circle_perim(circle c) {
    return 2 * 3 * c.radius;
}

static void circle_scale(void* self, so_int x) {
    circle* c = self;
    c->radius *= x;
}

so_String main_HttpStatus_String(main_HttpStatus s) {
    if (s == 200) {
        return so_str("OK");
    } else if (s == 404) {
        return so_str("Not Found");
    } else if (s == 500) {
        return so_str("Error");
    } else {
        return so_str("Other");
    }
}

int main(void) {
    main_Rect r = (main_Rect){.width = 10, .height = 5};
    {
        // Value + pointer receiver.
        so_int rArea = main_Rect_Area(&r);
        if (rArea != 50) {
            so_panic("unexpected area");
        }
        so_int rPerim = main_Rect_perim(&r, 2);
        if (rPerim != 60) {
            so_panic("unexpected perimeter");
        }
    }
    {
        // Pointer + pointer receiver.
        main_Rect* rp = &r;
        so_int rpArea = main_Rect_Area(rp);
        if (rpArea != 50) {
            so_panic("unexpected area");
        }
        so_int rpPerim = main_Rect_perim(rp, 2);
        if (rpPerim != 60) {
            so_panic("unexpected perimeter");
        }
    }
    {
        // Value + value receiver.
        main_Rect rResized = main_Rect_resize(r, 2);
        if (r.width != 10 || r.height != 5) {
            so_panic("unexpected original rect");
        }
        if (rResized.width != 20 || rResized.height != 10) {
            so_panic("unexpected resized rect");
        }
    }
    {
        // Pointer + value receiver.
        main_Rect* rp = &r;
        main_Rect rResized = main_Rect_resize(*rp, 2);
        if (r.width != 10 || r.height != 5) {
            so_panic("unexpected original rect");
        }
        if (rResized.width != 20 || rResized.height != 10) {
            so_panic("unexpected resized rect");
        }
    }
    {
        // Unexported type and method.
        circle c = (circle){.radius = 7};
        so_int cArea = circle_area(&c);
        if (cArea != 147) {
            so_panic("unexpected area");
        }
    }
    {
        // Blank receiver.
        if (main_Rect_sides(r) != 4) {
            so_panic("unexpected sides");
        }
        circle c = (circle){.radius = 7};
        if (circle_shape(&c) != 0) {
            so_panic("unexpected shape");
        }
    }
    {
        // Method on primitive type.
        main_HttpStatus s = 200;
        if (so_string_ne(main_HttpStatus_String(s), so_str("OK"))) {
            so_panic("unexpected string");
        }
        s = 404;
        if (so_string_ne(main_HttpStatus_String(s), so_str("Not Found"))) {
            so_panic("unexpected string");
        }
    }
    {
        // Method expression.
        circle c = (circle){.radius = 7};
        so_int (*areaFn)(circle*) = (so_int (*)(circle*))circle_area;
        so_int area = areaFn(&c);
        if (area != 147) {
            so_panic("unexpected area");
        }
        so_int (*perimFn)(circle) = circle_perim;
        so_int perim = perimFn(c);
        if (perim != 42) {
            so_panic("unexpected perimeter");
        }
        // No named func type matches this signature.
        void (*scaleFn)(circle*, so_int) = (void (*)(circle*, so_int))circle_scale;
        scaleFn(&c, 2);
        if (c.radius != 14) {
            so_panic("unexpected radius");
        }
    }
    return 0;
}
