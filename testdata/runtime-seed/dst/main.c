#include "main.h"

// -- Forward declarations --
static void check(bool ok, so_String msg);

// -- Variables and constants --

// seedCount is the number of seeds that the case collects.
static const so_unused int64_t seedCount = 16;

// -- Implementation --

static void check(bool ok, so_String msg) {
    if (!ok) {
        so_panic(so_cstr(msg));
    }
}

int main(void) {
    uint64_t seeds[16] = {};
    for (so_int i = 0; i < 16; i++) {
        seeds[i] = runtime_Seed();
    }
    // The fallback Seed implementation must never return 0.
    for (so_int _ = 0; _ < 16; _++) {
        uint64_t seed = seeds[_];
        check(seed != 0, so_str("runtime: zero seed"));
    }
    // The whole program draws from one counter, so no value repeats.
    for (so_int i = 0; i < 16; i++) {
        for (so_int j = i + 1; j < seedCount; j++) {
            check(seeds[i] != seeds[j], so_str("runtime: repeated seed"));
        }
    }
    so_println("%s", "ok");
    return 0;
}
