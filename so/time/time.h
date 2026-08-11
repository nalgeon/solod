#include "so/builtin/builtin.h"

#ifdef so_build_hosted
#include <errno.h>
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

// time_sleep pauses the calling thread for at least ns nanoseconds.
// It restarts if interrupted by a signal so the full duration elapses.
static inline void time_sleep(int64_t ns) {
    struct timespec req = {
        .tv_sec = (time_t)(ns / 1000000000LL),
        .tv_nsec = (long)(ns % 1000000000LL),
    };
    struct timespec rem;
    while (nanosleep(&req, &rem) != 0 && errno == EINTR) {
        req = rem;
    }
}

#else

typedef struct {
    int tm_sec;    // seconds after the minute [0-60]
    int tm_min;    // minutes after the hour [0-59]
    int tm_hour;   // hours since midnight [0-23]
    int tm_mday;   // day of the month [1-31]
    int tm_mon;    // months since January [0-11]
    int tm_year;   // years since 1900
    int tm_wday;   // days since Sunday [0-6]
    int tm_yday;   // days since January 1 [0-365]
    int tm_isdst;  // Daylight Saving Time flag
} time_tm;

static inline size_t strftime(char* str, size_t count, const char* format, time_tm* tm) {
    (void)str;
    (void)count;
    (void)format;
    (void)tm;
    so_panic("time: formatting requires a hosted environment");
    return 0;
}

static inline char* strptime(const char* str, const char* format, time_tm* tm) {
    (void)str;
    (void)format;
    (void)tm;
    so_panic("time: parsing requires a hosted environment");
    return NULL;
}

static inline so_R_i64_i32 time_wall() {
    so_panic("time: wall clock time requires a hosted environment");
    return (so_R_i64_i32){};
}

static inline int64_t time_mono() {
    return 0;
}

static inline void time_sleep(int64_t ns) {
    (void)ns;
    so_panic("time: sleep requires a hosted environment");
}

#endif  // so_build_hosted
