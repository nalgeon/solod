#include "so/builtin/builtin.h"

// NumCPU returns the number of online logical CPUs, always >= 1.
//
// A POSIX target reads _SC_NPROCESSORS_ONLN. This macro is not POSIX, but is
// widely available. Other hosted and freestanding targets always return 1.
static inline so_int runtime_NumCPU(void);

#if defined(so_build_hosted)

#if defined(so_build_posix)

#include <unistd.h>

static inline so_int runtime_NumCPU(void) {
    long n = sysconf(_SC_NPROCESSORS_ONLN);
    return n > 0 ? (so_int)n : 1;
}

#else  // !so_build_posix

static inline so_int runtime_NumCPU(void) {
    return 1;
}

#endif  // so_build_posix

// runtime_crand_read fills buf with size bytes of the cryptographic random of
// the operating system. It reports whether it filled the whole buffer. A
// caller that needs an error message writes the message of its own package.
bool runtime_crand_read(uint8_t* buf, so_int size);

#else  // !so_build_hosted

static inline so_int runtime_NumCPU(void) {
    return 1;
}

// so_crand_read is the entropy hook of the target. crypto/crand declares the
// same function and documents it. Seed reads the hook, so one definition gives
// the program both cryptographic random and an unpredictable seed.
so_int so_crand_read(uint8_t* buf, so_int size);

#endif  // so_build_hosted

// Seed returns a random 64-bit seed.
uint64_t runtime_Seed(void);

#if defined(so_build_hosted)
#define runtime_Hosted true
#else
#define runtime_Hosted false
#endif

#define runtime_buildVersion so_str(so_version)

#if defined(so_build_darwin)
#define runtime_GOOS so_str("darwin")
#elif defined(so_build_linux)
#define runtime_GOOS so_str("linux")
#elif defined(so_build_freebsd)
#define runtime_GOOS so_str("freebsd")
#elif defined(so_build_netbsd)
#define runtime_GOOS so_str("netbsd")
#elif defined(so_build_openbsd)
#define runtime_GOOS so_str("openbsd")
#elif defined(so_build_dragonfly)
#define runtime_GOOS so_str("dragonfly")
#elif defined(so_build_wasm) && defined(so_build_hosted)
#define runtime_GOOS so_str("wasip1")
#elif defined(so_build_windows)
#define runtime_GOOS so_str("windows")
#else
#define runtime_GOOS so_str("bare")
#endif

#if defined(so_build_amd64)
#define runtime_GOARCH so_str("amd64")
#elif defined(so_build_arm64)
#define runtime_GOARCH so_str("arm64")
#elif defined(so_build_riscv64)
#define runtime_GOARCH so_str("riscv64")
#elif defined(so_build_i386)
#define runtime_GOARCH so_str("386")
#elif defined(so_build_wasm32)
#define runtime_GOARCH so_str("wasm")
#else
#define runtime_GOARCH so_str("unknown")
#endif
