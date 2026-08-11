#include "main.h"

// -- Types --

typedef struct value value;

typedef struct value {
    so_int x;
} value;

// -- Forward declarations --
static void value_init(void* self, so_int x);

// -- Variables and constants --
static so_int state = 0;

// -- Implementation --

static void value_init(void* self, so_int x) {
    value* v = self;
    v->x = x;
}

int main(void) {
    {
        // Init function.
        if (state != 42) {
            so_panic("init() did not run");
        }
        so_println("%s", "ok");
    }
    {
        // Method named init (just a regular method).
        value v = {};
        value_init(&v, 123);
        if (v.x != 123) {
            so_panic("v.x != 123");
        }
    }
    return 0;
}

static void __attribute__((constructor)) main_init() {
    state = 42;
}
