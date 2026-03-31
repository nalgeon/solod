//go:build ignore
#include "so/builtin/builtin.h"
#include <fcntl.h>
#include <sys/stat.h>
#include <unistd.h>

// Stat result - flat struct filled by C helpers.
typedef struct {
    int64_t size;
    uint32_t mode;
    int64_t modSec;
    int64_t modNsec;
    uint64_t dev;
    uint64_t ino;
    bool ok;
} os_statResult;

// os_gethostname wraps gethostname with a null check to avoid
// glibc fortify-source nonnull warning when buf comes from a slice.
static int os_gethostname(so_byte* buf, so_int len) {
    if (buf == NULL) return -1;
    return gethostname((char*)buf, (size_t)len);
}

// os_getcwd wraps getcwd with a null check to avoid
// glibc fortify-source nonnull warning when buf comes from a slice.
static so_byte* os_getcwd(so_byte* buf, so_int len) {
    if (buf == NULL) return NULL;
    return (so_byte*)getcwd((char*)buf, (size_t)len);
}

// os_lstat fills result from lstat().
static os_statResult os_lstat(const char* path) {
    struct stat st;
    if (lstat(path, &st) != 0) return (os_statResult){.ok = false};
    return (os_statResult){
        .size = st.st_size,
        .mode = st.st_mode,
        .modSec = st.st_mtime,
        .modNsec = 0,  // fields differ on macos and linux, set to 0 for now
        .dev = st.st_dev,
        .ino = st.st_ino,
        .ok = true,
    };
}

// os_readlink wraps readlink with a null check to avoid
// glibc fortify-source nonnull warning when buf comes from a slice.
static so_int os_readlink(const char* path, so_byte* buf, so_int bufsiz) {
    if (buf == NULL) return -1;
    ssize_t n = readlink(path, (char*)buf, (size_t)bufsiz);
    return (so_int)n;
}

// os_stat fills result from stat().
static os_statResult os_stat(const char* path) {
    struct stat st;
    if (stat(path, &st) != 0) return (os_statResult){.ok = false};
    return (os_statResult){
        .size = st.st_size,
        .mode = st.st_mode,
        .modSec = st.st_mtime,
        .modNsec = 0,  // fields differ on macos and linux, set to 0 for now
        .dev = st.st_dev,
        .ino = st.st_ino,
        .ok = true,
    };
}

// os_utimens sets access and modification times using utimensat.
// A tv_nsec of UTIME_OMIT leaves the corresponding time unchanged.
static int os_utimens(const char* path, int64_t asec, int64_t ansec, int64_t msec, int64_t mnsec) {
    struct timespec times[2] = {
        {.tv_sec = asec, .tv_nsec = ansec},
        {.tv_sec = msec, .tv_nsec = mnsec},
    };
    return utimensat(AT_FDCWD, path, times, 0);
}
