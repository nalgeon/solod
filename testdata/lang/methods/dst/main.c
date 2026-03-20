#include "main.h"

// -- Types --

typedef struct circle {
    so_int radius;
} circle;

// -- Forward declarations --
static so_int main_Rect_perim(void* self, so_int n);
static main_Rect main_Rect_resize(main_Rect r, so_int x);
static so_int circle_area(void* self);

// -- Implementation --

so_int main_Rect_Area(void* self) {
    main_Rect* r = (main_Rect*)self;
    return r->width * r->height;
}

static so_int main_Rect_perim(void* self, so_int n) {
    main_Rect* r = (main_Rect*)self;
    return n * (2 * r->width + 2 * r->height);
}

static main_Rect main_Rect_resize(main_Rect r, so_int x) {
    r.height *= x;
    r.width *= x;
    return r;
}

static so_int circle_area(void* self) {
    circle* c = (circle*)self;
    return 3 * c->radius * c->radius;
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
}
