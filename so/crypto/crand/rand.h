#include "so/builtin/builtin.h"
#include "so/runtime/runtime.h"

#ifdef so_build_hosted

// read fills buf with size cryptographically secure random bytes.
// Panics if the operating system has no cryptographic random.
static inline void crand_read(uint8_t* buf, so_int size) {
    if (size <= 0) return;
    if (!runtime_crand_read(buf, size)) {
        so_panic("crypto/crand: cryptographic random not available");
    }
}

#else

// so_crand_read fills buf with size cryptographically secure random bytes and
// returns the number of bytes written. A freestanding environment has no
// entropy source of its own, so the target must define this function. Point it
// at the hardware random number generator of the board or at a host import.
//
// The default definition in builtin.c reads no bytes and returns 0.
so_int so_crand_read(uint8_t* buf, so_int size);

// read fills buf with size cryptographically secure random bytes.
// Panics if the target does not define so_crand_read.
static inline void crand_read(uint8_t* buf, so_int size) {
    if (size <= 0) return;
    if (so_crand_read(buf, size) != size) {
        so_panic("crypto/crand: define so_crand_read for this target");
    }
}

#endif  // so_build_hosted
