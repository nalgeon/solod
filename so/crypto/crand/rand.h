#include "so/builtin/builtin.h"

#ifdef so_build_hosted

// read fills buf with size cryptographically secure random bytes.
static inline void crand_read(uint8_t* buf, so_int size) {
    if (size <= 0) return;
#if defined(so_build_darwin) || defined(so_build_netbsd) || defined(so_build_openbsd)
    arc4random_buf(buf, (size_t)size);
#elif defined(so_build_linux) || defined(so_build_freebsd) || defined(so_build_dragonfly)
    while (size > 0) {
        ssize_t n = getrandom(buf, (size_t)size, 0);
        if (n < 0) {
            so_panic("crypto/crand: cryptographic random not available");
        }
        buf += n;
        size -= n;
    }
#elif defined(so_build_wasm)
    while (size > 0) {
        size_t n = size < 256 ? (size_t)size : 256;
        if (getentropy(buf, n) != 0) {
            so_panic("crypto/crand: cryptographic random not available");
        }
        buf += n;
        size -= (so_int)n;
    }
#else
    so_panic("crypto/crand: cryptographic random not available");
#endif
}

#else

// so_crand_read fills buf with size cryptographically secure random bytes and
// returns the number of bytes written. A freestanding environment has no
// entropy source of its own, so the target must define this function. Point it
// at the hardware random number generator of the board or at a host import.
//
// The declaration is weak, so a program that never calls crypto/crand still
// links. A call with no definition panics.
__attribute__((weak)) so_int so_crand_read(uint8_t* buf, so_int size);

// read fills buf with size cryptographically secure random bytes.
static inline void crand_read(uint8_t* buf, so_int size) {
    if (size <= 0) return;
    if (so_crand_read == NULL || so_crand_read(buf, size) != size) {
        so_panic("crypto/crand: define so_crand_read for this target");
    }
}

#endif  // so_build_hosted
