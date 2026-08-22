#include "so/builtin/builtin.h"

#if !defined(so_build_hosted)
#error "sync: hosted environment required"
#endif

#include <time.h>

// sync_condInitMono initializes cond so that its timed waits measure elapsed
// time on the monotonic clock, immune to wall-clock changes. Returns 0 on
// success or a pthread error code.
static inline int sync_condInitMono(pthread_cond_t* cond) {
#if defined(so_build_darwin)
    // Darwin lacks pthread_condattr_setclock; its timed waits use the
    // relative-np variant (see sync_condWaitRel), which already measures
    // the monotonic clock, so a default cond is correct.
    return pthread_cond_init(cond, NULL);
#else
    pthread_condattr_t attr;
    int rc = pthread_condattr_init(&attr);
    if (rc != 0) return rc;
    rc = pthread_condattr_setclock(&attr, CLOCK_MONOTONIC);
    if (rc == 0) rc = pthread_cond_init(cond, &attr);
    pthread_condattr_destroy(&attr);
    return rc;
#endif
}

// sync_condWaitRel atomically unlocks mu and blocks on cond until it is
// signaled or nsec nanoseconds elapse on the monotonic clock, then re-locks
// mu. A non-positive nsec polls without blocking. Returns 0 if signaled or
// ETIMEDOUT on timeout.
static inline int sync_condWaitRel(pthread_cond_t* cond, pthread_mutex_t* mu,
                                   int64_t nsec) {
    if (nsec < 0) nsec = 0;
    struct timespec dur = {
        .tv_sec = (time_t)(nsec / 1000000000LL),
        .tv_nsec = (long)(nsec % 1000000000LL),
    };
#if defined(so_build_darwin)
    return pthread_cond_timedwait_relative_np(cond, mu, &dur);
#else
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    ts.tv_sec += dur.tv_sec;
    ts.tv_nsec += dur.tv_nsec;
    if (ts.tv_nsec >= 1000000000LL) {
        ts.tv_sec += 1;
        ts.tv_nsec -= 1000000000LL;
    }
    return pthread_cond_timedwait(cond, mu, &ts);
#endif
}
