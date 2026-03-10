#include "main.h"

// -- Forward declarations (types) --
typedef struct circle circle;

// -- Forward declarations (functions and methods) --
static so_int main_Rect_perim(void* self, so_int n);
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

typedef struct circle {
    so_int radius;
} circle;

static so_int circle_area(void* self) {
    circle* c = (circle*)self;
    return 3 * c->radius * c->radius;
}

// Methods on primitive types are not supported.
// type HttpStatus int
// func (s HttpStatus) String() string {
// 	switch s {
// 	case 200:
// 		return "OK"
// 	case 404:
// 		return "Not Found"
// 	case 500:
// 		return "Error"
// 	default:
// 		return "Other"
// 	}
// }
int main(void) {
    main_Rect r = (main_Rect){.width = 10, .height = 5};
    so_int rArea = main_Rect_Area(&r);
    (void)rArea;
    so_int rPerim = main_Rect_perim(&r, 2);
    (void)rPerim;
    main_Rect* rp = &r;
    so_int rpArea = main_Rect_Area(rp);
    (void)rpArea;
    so_int rpPerim = main_Rect_perim(rp, 2);
    (void)rpPerim;
    circle c = (circle){.radius = 7};
    so_int cArea = circle_area(&c);
    (void)cArea;
}
