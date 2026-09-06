#include "main.h"

// -- Forward declarations --
static so_int unnamed1(so_int _0 so_unused);
static so_int unnamed2(so_int _0 so_unused, float _1 so_unused);
static so_int blank1(so_int _0 so_unused);
static so_int blank2(so_int n, so_int _1 so_unused);
static so_int blank3(so_int _0 so_unused, float _1 so_unused, so_int n);

// -- Variables and constants --

// Blank package-level constants.

// Iota with blank identifiers.
static const so_unused int64_t iotaD = 3;
static const so_unused int64_t iotaE = 4;

// Blank package-level variables.

// Blank package-level variables of different types.
static so_unused main_Value value = (main_Value){42};

// -- Implementation --

// Unnamed receiver.
so_int main_Value_One(main_Value self) {
    (void)self;
    return 1;
}

// Blank receiver.
so_int main_Value_Two(main_Value self) {
    (void)self;
    return 2;
}

// Unnamed method parameters.
so_int main_Value_Decr1(main_Value v, so_int _1 so_unused) {
    return v.x - 1;
}

so_int main_Value_Decr2(void* self, so_int _1 so_unused, float _2 so_unused) {
    main_Value* v = self;
    return v->x - 2;
}

// Blank method parameters.
so_int main_Value_Incr1(main_Value v, so_int _1 so_unused) {
    return v.x + 1;
}

so_int main_Value_Incr2(main_Value v, so_int n, float _2 so_unused) {
    return v.x + n;
}

so_int main_Value_Incr3(void* self, so_int _1 so_unused, float _2 so_unused, so_int n) {
    main_Value* v = self;
    return v->x + n;
}

// Unnamed function parameters.
static so_int unnamed1(so_int _0 so_unused) {
    return 1;
}

static so_int unnamed2(so_int _0 so_unused, float _1 so_unused) {
    return 2;
}

// Blank function parameters.
static so_int blank1(so_int _0 so_unused) {
    return 1;
}

static so_int blank2(so_int n, so_int _1 so_unused) {
    return n;
}

static so_int blank3(so_int _0 so_unused, float _1 so_unused, so_int n) {
    return n;
}

int main(void) {
    {
        if (iotaD != 3 || iotaE != 4) {
            so_panic("unexpected iota values");
        }
    }
    {
        // Keyed literal, field assignment, and field access.
        main_Point p = (main_Point){.x = 1, .y = 2};
        p.x = 3;
        if (p.x != 3 || p.y != 2) {
            so_panic("unexpected Point value");
        }
    }
    {
        // Positional literal also fills the blank fields.
        main_Point q = (main_Point){4, 0, 5, 0};
        if (q.x != 4 || q.y != 5) {
            so_panic("unexpected Point value");
        }
    }
    {
        // Inner struct field.
        main_Wrapper w = {};
        w.inner.n = 6;
        if (w.inner.n != 6) {
            so_panic("unexpected Wrapper value");
        }
    }
    {
        // Local anonymous struct.
        struct { double f; so_String _1; so_String _2; } st = {};
        st.f = 7.0;
        if (st.f != 7.0) {
            so_panic("unexpected st value");
        }
    }
    {
        // Anonymous struct literal.
        so_auto lit = (struct {
            so_int n;
            so_String _1;
        }){
            .n = 8,
            ._1 = so_str("skip"),
        };
        if (lit.n != 8) {
            so_panic("unexpected lit value");
        }
    }
    {
        // Unnamed or blank method receiver.
        main_Value v = (main_Value){42};
        if (main_Value_One(v) != 1) {
            so_panic("unexpected Value.One result");
        }
        if (main_Value_Two(v) != 2) {
            so_panic("unexpected Value.Two result");
        }
    }
    {
        // Unnamed or blank method parameter.
        main_Value v = (main_Value){5};
        if (main_Value_Decr1(v, 1) != 4) {
            so_panic("unexpected Value.Decr1 result");
        }
        if (main_Value_Decr2(&v, 1, 2.0f) != 3) {
            so_panic("unexpected Value.Decr2 result");
        }
        if (main_Value_Incr1(v, 1) != 6) {
            so_panic("unexpected Value.Incr1 result");
        }
        if (main_Value_Incr2(v, 5, 6.0f) != 10) {
            so_panic("unexpected Value.Incr2 result");
        }
        if (main_Value_Incr3(&v, 5, 6.0f, 7) != 12) {
            so_panic("unexpected Value.Incr3 result");
        }
    }
    {
        // Interface methods with unnamed and blank parameters.
        main_Valuer v = (main_Valuer){.self = &(main_Value){5}, .Decr2 = main_Value_Decr2, .Incr3 = main_Value_Incr3};
        if (main_Valuer_Decr2(v, 1, 2.0f) != 3) {
            so_panic("unexpected Valuer.Decr2 result");
        }
        if (main_Valuer_Incr3(v, 5, 6.0f, 7) != 12) {
            so_panic("unexpected Valuer.Incr3 result");
        }
    }
    {
        // Unnamed or blank function parameters.
        if (unnamed1(5) != 1) {
            so_panic("unexpected unnamed1 result");
        }
        if (unnamed2(5, 6.0f) != 2) {
            so_panic("unexpected unnamed2 result");
        }
        if (blank1(5) != 1) {
            so_panic("unexpected blank1 result");
        }
        if (blank2(5, 6) != 5) {
            so_panic("unexpected blank2 result");
        }
        if (blank3(5, 6.0f, 7) != 7) {
            so_panic("unexpected blank3 result");
        }
    }
    {
        // Unnamed or blank generic function parameters.
        if (unnamedGen1(so_int, (5)) != 1) {
            so_panic("unexpected unnamedGen1 result");
        }
        if (unnamedGen2(so_int, (5), (6.0f)) != 2) {
            so_panic("unexpected unnamedGen2 result");
        }
        if (blankGen1(so_int, (5)) != 1) {
            so_panic("unexpected blankGen1 result");
        }
        if (blankGen2(so_int, (5), (6)) != 5) {
            so_panic("unexpected blankGen2 result");
        }
        if (blankGen3(so_int, (5), (6.0f), (7)) != 7) {
            so_panic("unexpected blankGen3 result");
        }
    }
    {
        // Discarding values with blank identifier.
        so_int v1 = 11;
        so_int v2 = 22;
        so_int v3 = 51;
        (void)52;
        (void)61;
        so_int v4 = 62;
        (void)71;
        (void)72;
        (void)81;
        (void)v1;
        (void)v2;
        (void)v3;
        (void)v4;
    }
    {
        // Discarding an array literal.
        (void)(so_int[3]){1, 2, 3};
        (void)(so_int[3]){1, 2, 3};
        (void)(so_Slice[2]){(so_Slice){}, (so_Slice){}};
        (void)(so_int[3]){1, 2, 3};
        so_int n1 = 1;
        (void)(so_int[2]){2, 3};
        (void)n1;
    }
    return 0;
}
