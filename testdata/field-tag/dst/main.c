#include "main.h"

// -- Implementation --

int main(void) {
    // Keyed literal, field assignment, and field access.
    main_Event e = (main_Event){.type = 7, .data = 42};
    e.type = 8;
    if (e.type != 8) {
        so_panic("unexpected Event.etype value");
    }
    // Positional literal.
    main_Event p = (main_Event){9, 10};
    if (p.type != 9) {
        so_panic("unexpected Event.etype value");
    }
    // Instantiated generic field resolves the same override.
    main_Box b = (main_Box){.ident = 5};
    if (b.ident != 5) {
        so_panic("unexpected Box.id value");
    }
    // Override declared in an imported package still applies here.
    sdl_CommonEvent ce = (sdl_CommonEvent){.type = 3};
    ce.type = 4;
    if (ce.type != 4) {
        so_panic("unexpected CommonEvent.Type value");
    }
    return 0;
}
