#include "main.h"

// -- Types --

typedef struct point point;

typedef struct point {
    so_int x;
    so_int y;
} point;

// -- Implementation --

int main(void) {
    // append a composite-literal value.
    so_Slice pts = so_make_slice(point, 0, 2);
    pts = so_append(point, pts, 1, ((point){1, 2}));
    if (so_at(point, pts, 0).y != 2) {
        so_panic("append value");
    }
    // map with a composite-literal value.
    so_Map* mv = so_make_map(so_int, point, 1);
    so_map_set(so_int, point, mv, 0, ((point){3, 4}));
    if (so_map_get(so_int, point, mv, 0).x != 3) {
        so_panic("map value");
    }
    // map with a composite-literal key.
    so_Map* mk = so_make_map(point, so_int, 1);
    so_map_set(point, so_int, mk, ((point){1, 2}), 42);
    if (so_map_get(point, so_int, mk, ((point){1, 2})) != 42) {
        so_panic("map key");
    }
    so_int v = so_map_get(point, so_int, mk, ((point){1, 2}));
    bool ok = so_map_has(point, mk, ((point){1, 2}));
    if (!ok || v != 42) {
        so_panic("map key comma-ok");
    }
    return 0;
}
