#include "main.h"

// -- Forward declarations --
static so_int areaOf(main_Shape s);
static main_Shape asShape(main_Rect* r);

// -- Implementation --

so_int main_Rect_Area(void* self) {
    main_Rect* r = self;
    return r->width * r->height;
}

static so_int areaOf(main_Shape s) {
    return main_Shape_Area(s);
}

static main_Shape asShape(main_Rect* r) {
    return (main_Shape){.self = r, .Area = main_Rect_Area};
}

int main(void) {
    main_Rect r = (main_Rect){.width = 10, .height = 5};
    {
        // An alias to a pointer converts to an interface.
        main_Shape s = (main_Shape){.self = NULL, .Area = main_Rect_Area};
        main_Rect* p = &r;
        s = (main_Shape){.self = p, .Area = main_Rect_Area};
        if (main_Shape_Area(s) != 50) {
            so_panic("s.Area() != 50");
        }
    }
    {
        // An alias to a pointer passes as an interface argument and result.
        main_Rect* p = &r;
        if (areaOf((main_Shape){.self = p, .Area = main_Rect_Area}) != 50) {
            so_panic("areaOf(p) != 50");
        }
        if (main_Shape_Area(asShape(p)) != 50) {
            so_panic("asShape(p).Area() != 50");
        }
    }
    {
        // A struct field of an alias type converts to an interface.
        main_Frame f = (main_Frame){.rect = &r};
        main_Shape s = (main_Shape){.self = f.rect, .Area = main_Rect_Area};
        if (main_Shape_Area(s) != 50) {
            so_panic("s.Area() != 50");
        }
    }
    {
        // A type assertion accepts an alias.
        main_Shape s = (main_Shape){.self = &r, .Area = main_Rect_Area};
        main_Rect* p = (main_Rect*)s.self;
        if (main_Rect_Area(p) != 50) {
            so_panic("p.Area() != 50");
        }
        {
            bool ok = (s.Area == main_Rect_Area);
            if (!ok) {
                so_panic("want ok == true");
            }
        }
    }
    {
        // An alias to a named interface holds a concrete value.
        main_Shape c = (main_Shape){.self = &r, .Area = main_Rect_Area};
        if (main_Shape_Area(c) != 50) {
            so_panic("c.Area() != 50");
        }
    }
    {
        // A method expression accepts an alias receiver.
        so_int (*area)(main_Rect*) = (so_int (*)(main_Rect*))main_Rect_Area;
        if (area(&r) != 50) {
            so_panic("area(&r) != 50");
        }
    }
    {
        // Two variables of an alias pointer type declare separately.
        main_Rect r2 = (main_Rect){.width = 3, .height = 4};
        main_Rect* p1 = &r;
        main_Rect* p2 = &r2;
        if (main_Rect_Area(p1) != 50) {
            so_panic("p1.Area() != 50");
        }
        if (main_Rect_Area(p2) != 12) {
            so_panic("p2.Area() != 12");
        }
    }
    {
        // print accepts an alias to a pointer.
        main_Rect* p = NULL;
        so_println("%p", p);
    }
    return 0;
}
