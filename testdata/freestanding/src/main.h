// HOOKED reports whether crypto/crand and time draw from the target hooks in
// main.c. A hosted build draws from the operating system instead, so the values
// it gets back are the real ones and the test cannot predict them.
#ifdef so_build_hosted
#define HOOKED 0
#else
#define HOOKED 1
#endif
