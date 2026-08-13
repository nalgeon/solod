// begin include
// CRAND_HOOKED reports whether crypto/crand draws from so_crand_read in main.c.
// A hosted build draws from the CSPRNG of the operating system instead.
#ifdef so_build_hosted
#define CRAND_HOOKED 0
#else
#define CRAND_HOOKED 1
#endif
// end include
