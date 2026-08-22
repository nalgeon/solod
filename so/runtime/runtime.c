//go:build ignore
#include "runtime.h"

#if defined(so_build_hosted)

#if defined(so_build_darwin) || defined(so_build_netbsd) || defined(so_build_openbsd)
#include <stdlib.h>
#elif defined(so_build_linux) || defined(so_build_freebsd) || defined(so_build_dragonfly)
#include <sys/random.h>
#include <sys/types.h>  // ssize_t
#elif defined(so_build_wasm)
#include <unistd.h>
#elif defined(so_build_windows)
// BCryptGenRandom requires bcrypt. Link the program with -lbcrypt,
// because a so:link directive cannot depend on the target.
#include <windows.h>

#include <bcrypt.h>
#endif

// runtime_crand_read fills buf with size bytes of the
// cryptographic random of the operating system.
bool runtime_crand_read(uint8_t* buf, so_int size) {
    if (size <= 0) return true;
#if defined(so_build_darwin) || defined(so_build_netbsd) || defined(so_build_openbsd)
    arc4random_buf(buf, (size_t)size);
    return true;
#elif defined(so_build_linux) || defined(so_build_freebsd) || defined(so_build_dragonfly)
    while (size > 0) {
        ssize_t n = getrandom(buf, (size_t)size, 0);
        if (n < 0) return false;
        buf += n;
        size -= (so_int)n;
    }
    return true;
#elif defined(so_build_wasm)
    // getentropy reads 256 bytes at most.
    while (size > 0) {
        size_t n = size < 256 ? (size_t)size : 256;
        if (getentropy(buf, n) != 0) return false;
        buf += n;
        size -= (so_int)n;
    }
    return true;
#elif defined(so_build_windows)
    // BCryptGenRandom counts the bytes in 32 bits, so a larger buffer
    // needs more calls. A null algorithm handle draws from the preferred
    // generator of the system.
    while (size > 0) {
        ULONG n = size > 0x40000000 ? 0x40000000 : (ULONG)size;
        if (BCryptGenRandom(NULL, buf, n, BCRYPT_USE_SYSTEM_PREFERRED_RNG) != 0) {
            return false;
        }
        buf += n;
        size -= (so_int)n;
    }
    return true;
#else
    (void)buf;
    return false;
#endif
}

// Seed returns a random 64-bit seed.
uint64_t runtime_Seed(void) {
    uint64_t seed = 0;
    if (!runtime_crand_read((uint8_t*)&seed, 8)) {
        so_panic("runtime: cryptographic random not available");
    }
    return seed;
}

#else  // !so_build_hosted

// seed_ticket numbers the calls to the fallback sequence of Seed.
static uint32_t seed_ticket = 0;

// Seed returns a random 64-bit seed.
uint64_t runtime_Seed(void) {
    uint64_t seed = 0;
    if (so_crand_read((uint8_t*)&seed, 8) == 8 && seed != 0) {
        return seed;
    }
#if __GCC_ATOMIC_INT_LOCK_FREE == 2
    uint32_t n = __atomic_fetch_add(&seed_ticket, 1u, __ATOMIC_RELAXED);
#else
    // The target has no lock-free 32-bit atomic. A plain add is not thread-safe.
    uint32_t n = seed_ticket++;
#endif
    // SplitMix64 mixes the ticket into a well distributed 64-bit seed.
    // The mixer maps 0 to 0, so the added constant makes the input n + 1 and
    // the result is never 0. Seed reads a 0 from the hook as a broken hook,
    // so the fallback must not return 0 either.
    uint64_t x = (uint64_t)n * 0x9e3779b97f4a7c15ULL + 0x9e3779b97f4a7c15ULL;
    x ^= x >> 30;
    x *= 0xbf58476d1ce4e5b9ULL;
    x ^= x >> 27;
    x *= 0x94d049bb133111ebULL;
    x ^= x >> 31;
    return x;
}

#endif  // so_build_hosted
