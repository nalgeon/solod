#include "main.h"

// -- Forward declarations (types) --
typedef struct point point;

// -- Forward declarations (functions and methods) --
static void acceptAny(void* v);
static void acceptByte(so_byte* v);
static void acceptPoint(point* v);

// -- Implementation --

typedef struct point {
    so_int x;
    so_int y;
} point;

static void acceptAny(void* v) {
    (void)v;
}

static void acceptByte(so_byte* v) {
    (void)v;
}

static void acceptPoint(point* v) {
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
        // String value.
        so_String s = so_str("hello");
        acceptAny(&s);
        acceptAny(&s);
        acceptByte((so_byte*)&s);
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
        // Any casts.
        so_int n = 42;
        void* a = &n;
        so_byte* b = (so_byte*)a;
        if (*b != 42) {
            so_panic("want *b == 42");
        }
        so_String s1 = so_str("hello");
        a = &s1;
        so_String* s2 = (so_String*)a;
        if (s2 != &s1) {
            so_panic("want s2 == s1");
        }
    }
}
