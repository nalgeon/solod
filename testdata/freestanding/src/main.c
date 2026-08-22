//go:build ignore

#if !defined(so_build_hosted)

// This case stands in for the board. A freestanding environment has no entropy
// source and no clock, so it defines the target hooks that crypto/crand and
// time need. A real target reads its hardware here.

// so_crand_read counts up, so the test can check the bytes it gets back.
so_int so_crand_read(uint8_t* buf, so_int size) {
    for (so_int i = 0; i < size; i++) {
        buf[i] = (uint8_t)(i + 1);
    }
    return size;
}

// fakeMono counts the nanoseconds that so_time_mono and so_time_sleep report.
static int64_t fakeMono = 1000000;

// so_time_wall reports a fixed 2009-11-10T23:00:00Z, the Go playground date.
so_R_i64_i32 so_time_wall(void) {
    return (so_R_i64_i32){.val = 1257894000, .val2 = 0};
}

// so_time_mono adds a millisecond on every call, so elapsed time always grows.
int64_t so_time_mono(void) {
    fakeMono += 1000000;
    return fakeMono;
}

// so_time_sleep adds the full duration to the monotonic count.
void so_time_sleep(int64_t ns) {
    fakeMono += ns;
}

#endif  // !so_build_hosted
