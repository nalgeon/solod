//go:build ignore

#if CRAND_HOOKED
// so_crand_read is the entropy source that crypto/crand needs in a freestanding
// environment. A real target reads the hardware random number generator here.
// This one counts up, so the test can check the bytes it gets back.
so_int so_crand_read(uint8_t* buf, so_int size) {
    for (so_int i = 0; i < size; i++) {
        buf[i] = (uint8_t)(i + 1);
    }
    return size;
}
#endif  // CRAND_HOOKED
