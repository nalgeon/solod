//go:build ignore
#include "time.h"

int64_t time_monoStart = 1;

static void __attribute__((constructor)) init() {
    time_monoStart = time_mono() - 1;
}
