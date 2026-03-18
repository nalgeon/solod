#include "so/builtin/builtin.h"
#include <time.h>

// Monotonic times are reported as offsets from monoStart.
extern int64_t time_monoStart;

// wall returns the current wall clock time.
static inline so_Result time_wall() {
    struct timespec ts;
    clock_gettime(CLOCK_REALTIME, &ts);
    so_Value sec = {.as_i64 = ts.tv_sec};
    so_Value nsec = {.as_i32 = (int32_t)ts.tv_nsec};
    return (so_Result){.val = sec, .val2 = nsec, .err = NULL};
}

// mono returns the current monotonic time in nanoseconds.
static inline int64_t time_mono() {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (int64_t)ts.tv_sec * 1000000000LL + ts.tv_nsec;
}
