#include "so/builtin/builtin.h"

#ifdef so_build_hosted
#include <errno.h>
#include <time.h>

#define time_tm struct tm

#ifdef so_build_windows

static inline char* strptime(const char* str, const char* format, struct tm* tm) {
    (void)str;
    (void)format;
    (void)tm;
    so_panic("time: parsing a custom layout requires a POSIX environment");
    return NULL;
}

#else

// strptime may not be declared without _XOPEN_SOURCE before system headers.
// Provide an explicit declaration for portability (e.g. glibc with gcc).
char* strptime(const char*, const char*, struct tm*);

#endif  // so_build_windows

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

// so_time_wall returns the current wall clock time as seconds and nanoseconds
// since the Unix epoch. A freestanding environment has no clock of its own, so
// the target must define this function. Point it at the real time clock of the
// board or at a host import. A target that counts elapsed time only returns 0
// seconds, which dates every Time at the epoch and keeps Since and Until exact.
//
// The default definition in builtin.c panics.
so_R_i64_i32 so_time_wall(void);

// so_time_mono returns a monotonic count of nanoseconds from an arbitrary
// origin. The target must define this function to get a monotonic clock. Point
// it at a timer peripheral, at the tick counter of the operating system, or at
// a host import.
//
// The count must never decrease and must never be 0, because Now reads 0 as the
// absence of a monotonic clock. Convert the tick of the board to nanoseconds
// here, and widen a counter that wraps: a 32-bit counter at 1 kHz wraps after
// 49 days.
//
// The default definition in builtin.c returns 0. A target with no definition
// of its own has no monotonic clock, so Time holds a wall clock reading alone.
int64_t so_time_mono(void);

// so_time_sleep pauses for at least ns nanoseconds. The target must define this
// function. Point it at a wait instruction, at the delay of the operating
// system, or at a host import.
//
// The default definition in builtin.c panics.
void so_time_sleep(int64_t ns);

static inline so_R_i64_i32 time_wall() {
    return so_time_wall();
}

static inline int64_t time_mono() {
    return so_time_mono();
}

static inline void time_sleep(int64_t ns) {
    so_time_sleep(ns);
}

#endif  // so_build_hosted
