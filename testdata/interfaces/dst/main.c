#include "main.h"

// -- Forward declarations --
static so_int calcShape(main_Shape s);
static bool shapeIsRect(main_Shape s);
static main_Rect* shapeAsRect(main_Shape s);
static main_Shape rectAsShape(main_Rect* r);
static bool shapeCheckAssign(main_Shape s);
static main_Shape nilShape(void);
static main_Shape countShape(main_Rect* r);

// -- Variables and constants --
static so_unused so_int shapeCalls = 0;

// -- Implementation --

so_int main_Rect_Area(void* self) {
    main_Rect* r = self;
    return r->width * r->height;
}

so_int main_Rect_Perim(void* self, so_int n) {
    main_Rect* r = self;
    return n * (2 * r->width + 2 * r->height);
}

so_int main_Rect_Paint(void* self, so_int n, so_String _2 so_unused) {
    main_Rect* r = self;
    return n * r->width;
}

so_int main_Rect_Fill(void* self, so_int n) {
    main_Rect* r = self;
    return n + r->height;
}

static so_int calcShape(main_Shape s) {
    return main_Shape_Perim(s, 2) + main_Shape_Area(s);
}

static bool shapeIsRect(main_Shape s) {
    bool ok = (s.Area == main_Rect_Area);
    return ok;
}

static main_Rect* shapeAsRect(main_Shape s) {
    {
        bool ok = (s.Area == main_Rect_Area);
        if (!ok) {
            return NULL;
        }
    }
    main_Rect* r = (main_Rect*)s.self;
    return r;
}

static main_Shape rectAsShape(main_Rect* r) {
    return (main_Shape){.self = r, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
}

static bool shapeCheckAssign(main_Shape s) {
    bool ok = false;
    ok = (s.Area == main_Rect_Area);
    return ok;
}

static main_Shape nilShape(void) {
    return (main_Shape){};
}

static main_Shape countShape(main_Rect* r) {
    shapeCalls++;
    return (main_Shape){.self = r, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
}

int main(void) {
    main_Rect r = (main_Rect){.width = 10, .height = 5};
    {
        // Shape interface is implemented by *Rect pointer.
        main_Shape s = (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
        // also works
        main_Shape s2 = (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
        (void)s2;
        calcShape(s);
        // also works
        calcShape((main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim});
        // also works
        calcShape((main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim});
        (void)shapeIsRect(s);
        (void)shapeCheckAssign(s);
        main_Rect* rval = shapeAsRect(s);
        (void)rval;
    }
    {
        // Wrap Rect value into Shape via function.
        main_Shape s = rectAsShape(&r);
        (void)s;
    }
    {
        // Nil interface. Nil works on either side.
        main_Shape s1 = {};
        if (s1.self != NULL) {
            so_panic("want nil interface");
        }
        if (s1.self != NULL) {
            so_panic("want nil interface");
        }
        main_Shape s2 = (main_Shape){};
        if (s2.self != NULL) {
            so_panic("want nil interface");
        }
        main_Shape s3 = nilShape();
        if (s3.self != NULL) {
            so_panic("want nil interface");
        }
        bool isRect = shapeIsRect((main_Shape){});
        if (isRect) {
            so_panic("want isRect == false");
        }
        main_Rect r = {};
        main_Shape s4 = (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
        if (s4.self == NULL) {
            so_panic("want non-nil interface");
        }
        if (s4.self == NULL) {
            so_panic("want non-nil interface");
        }
    }
    {
        // Interface field in struct.
        main_Canvas c1 = (main_Canvas){.name = so_str("c1"), .shape = (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim}};
        if (main_Shape_Area(c1.shape) != 50) {
            so_panic("c1.shape.Area() != 50");
        }
        main_Canvas c2 = (main_Canvas){.name = so_str("c2"), .shape = (main_Shape){.self = &(main_Rect){5, 4}, .Area = main_Rect_Area, .Perim = main_Rect_Perim}};
        if (main_Shape_Area(c2.shape) != 20) {
            so_panic("c2.shape.Area() != 20");
        }
        main_Canvas c3 = (main_Canvas){.name = so_str("c3"), .shape = (main_Shape){}};
        if (c3.shape.self != NULL) {
            so_panic("c3.shape != nil");
        }
    }
    {
        // Interface field assignment.
        main_Canvas c = {};
        c.shape = (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
        if (main_Shape_Area(c.shape) != 50) {
            so_panic("c.shape.Area() != 50");
        }
    }
    {
        // Existing interface in struct literal.
        main_Shape s = (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
        main_Canvas c = (main_Canvas){.name = so_str("wrap"), .shape = s};
        if (main_Shape_Area(c.shape) != 50) {
            so_panic("c.shape.Area() != 50");
        }
    }
    {
        // Multi-var interface declaration.
        main_Rect r2 = (main_Rect){.width = 3, .height = 4};
        main_Shape s1 = (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim}, s2 = (main_Shape){.self = &r2, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
        if (main_Shape_Area(s1) != 50) {
            so_panic("s1.Area() != 50");
        }
        if (main_Shape_Area(s2) != 12) {
            so_panic("s2.Area() != 12");
        }
    }
    {
        // Redeclared interface variable.
        main_Shape s = {};
        s = (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
        so_int n = 42;
        (void)n;
        if (main_Shape_Area(s) != 50) {
            so_panic("s.Area() != 50");
        }
    }
    {
        // Call interface method as function.
        main_Shape s = (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
        so_int area = main_Shape_Area(s);
        if (area != 50) {
            so_panic("Shape.Area(s) != 50");
        }
    }
    {
        // A method call evaluates the interface value once.
        shapeCalls = 0;
        if (main_Shape_Area(countShape(&r)) != 50) {
            so_panic("countShape(&r).Area() != 50");
        }
        if (shapeCalls != 1) {
            so_panic("shapeCalls != 1");
        }
    }
    {
        // Interface methods with unnamed and blank parameters.
        main_Painter p = (main_Painter){.self = &r, .Fill = main_Rect_Fill, .Paint = main_Rect_Paint};
        if (main_Painter_Paint(p, 2, so_str("x")) != 20) {
            so_panic("p.Paint(2, \"x\") != 20");
        }
        if (main_Painter_Fill(p, 1) != 6) {
            so_panic("p.Fill(1) != 6");
        }
    }
    {
        // Method call through a pointer to an interface.
        main_Shape s = (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim};
        main_Shape* p = &s;
        if (main_Shape_Area((*p)) != 50) {
            so_panic("(*p).Area() != 50");
        }
    }
    return 0;
}
