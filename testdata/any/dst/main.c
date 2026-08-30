#include "main.h"

// -- Types --

typedef struct point point;
typedef struct rect rect;
typedef so_int number;

typedef struct point {
    so_int x;
    so_int y;
} point;

typedef struct shape {
    void* self;
    so_int (*area)(void* self);
} shape;

static inline so_unused so_int shape_area(shape self) {
    return self.area(self.self);
}

typedef struct rect {
    so_int width;
    so_int height;
} rect;

// -- Forward declarations --
static so_int rect_area(void* self);
static void acceptAny(void* v);
static void acceptByte(so_byte* v);
static void acceptPoint(point* v);
static void acceptShape(shape v);

// -- Implementation --

static so_int rect_area(void* self) {
    rect* r = self;
    return r->width * r->height;
}

static void acceptAny(void* v) {
    (void)v;
}

static void acceptByte(so_byte* v) {
    (void)v;
}

static void acceptPoint(point* v) {
    (void)v;
}

static void acceptShape(shape v) {
    (void)v;
}

int main(void) {
    {
        // Nil value.
        void* n = NULL;
        acceptAny(n);
        acceptAny(n);
    }
    {
        // Integer value.
        so_int n = 42;
        acceptAny(&n);
        acceptAny(&n);
        acceptByte((so_byte*)&n);
        acceptAny(&(so_int){42});
    }
    {
        // Integer pointer.
        so_int nval = 42;
        so_int* n = &nval;
        acceptAny(n);
        acceptAny(n);
        acceptByte((so_byte*)n);
    }
    {
        // Unsafe pointer.
        so_int nval = 42;
        void* p = (void*)(&nval);
        acceptAny(p);
        acceptAny(p);
        if ((void*)p != p) {
            so_panic("want any(p).(unsafe.Pointer) == p");
        }
    }
    {
        // String value.
        so_String s = so_str("hello");
        acceptAny(&s);
        acceptAny(&s);
        acceptByte((so_byte*)&s);
        acceptAny(&so_str("hello"));
    }
    {
        // String pointer.
        so_String sval = so_str("hello");
        so_String* s = &sval;
        acceptAny(s);
        acceptAny(s);
        acceptByte((so_byte*)s);
    }
    {
        // Slice value.
        so_Slice s = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        acceptAny(&s);
        acceptAny(&s);
        acceptByte((so_byte*)&s);
        acceptAny(&(so_Slice){(so_int[3]){1, 2, 3}, 3, 3});
    }
    {
        // Slice pointer.
        so_Slice sval = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        so_Slice* s = &sval;
        acceptAny(s);
        acceptAny(s);
        acceptByte((so_byte*)s);
    }
    {
        // Struct value.
        point p = (point){1, 2};
        acceptAny(&p);
        acceptAny(&p);
        acceptPoint((point*)&p);
        acceptAny(&(point){1, 2});
    }
    {
        // Struct pointer.
        point pval = (point){1, 2};
        point* p = &pval;
        acceptAny(p);
        acceptAny(p);
        acceptPoint((point*)p);
    }
    {
        // Interface value.
        shape s = (shape){.self = &(rect){.width = 10, .height = 5}, .area = rect_area};
        acceptAny(&s);
        acceptAny(&s);
        acceptShape((*(shape*)&s));
    }
    {
        // Any value casts.
        so_int i = 42;
        void* a = &i;
        if ((*(so_int*)a) != 42) {
            so_panic("want a.(int) == 42");
        }
        number n = 42;
        a = &n;
        if ((*(number*)a) != 42) {
            so_panic("want a.(number) == 42");
        }
        so_String s = so_str("hello");
        a = &s;
        if (so_string_ne((*(so_String*)a), so_str("hello"))) {
            so_panic("want a.(string) == \"hello\"");
        }
        point p = (point){1, 2};
        a = &p;
        point ap = (*(point*)a);
        if (ap.x != 1 || ap.y != 2) {
            so_panic("want a.(point) == point{1, 2}");
        }
    }
    {
        // Any pointer casts.
        so_int i = 42;
        void* a = &i;
        if ((so_int*)a != &i) {
            so_panic("want a.(*int) == &i");
        }
        number n = 42;
        a = &n;
        if ((number*)a != &n) {
            so_panic("want a.(*number) == &n");
        }
        so_String s = so_str("hello");
        a = &s;
        if ((so_String*)a != &s) {
            so_panic("want a.(*string) == &s");
        }
        point p1 = (point){1, 2};
        a = &p1;
        if ((point*)a != &p1) {
            so_panic("want a.(*point) == &p1");
        }
    }
    {
        // Any interface casts.
        void* a = NULL;
        rect r = (rect){.width = 10, .height = 5};
        shape sh = (shape){.self = &r, .area = rect_area};
        a = &sh;
        shape ashape = (*(shape*)a);
        if (shape_area(ashape) != rect_area(&r)) {
            so_panic("want a.(shape) == shape(&r)");
        }
    }
    return 0;
}
