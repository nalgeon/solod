#include "so/builtin/builtin.h"
#include <time.h>

#define time_tm struct tm

// strptime may not be declared without _XOPEN_SOURCE before system headers.
// Provide an explicit declaration for portability (e.g. glibc with gcc).
char* strptime(const char*, const char*, struct tm*);

// wall returns the current wall clock time.
static inline so_R_i64_i32 time_wall() {
    struct timespec ts;
    clock_gettime(CLOCK_REALTIME, &ts);
    return (so_R_i64_i32){.val = ts.tv_sec, .val2 = (int32_t)ts.tv_nsec};
}

// mono returns the current monotonic time in nanoseconds.
static inline int64_t time_mono() {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (int64_t)ts.tv_sec * 1000000000LL + ts.tv_nsec;
}
