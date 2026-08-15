//go:build ignore

// This file supplies the target hooks for a freestanding test run on wasm32.
// A wasm32-freestanding module can still import the functions of
// wasi_snapshot_preview1, and a runtime like wasmtime supplies them. So the
// test report reaches the terminal, and the clock and the entropy source of the
// host reach the tests.

#include "so/builtin/builtin.h"

// ciovec is the buffer descriptor that fd_write reads. The WASI ABI is 32-bit,
// so both fields are 32-bit.
typedef struct {
    const uint8_t* buf;
    uint32_t len;
} ciovec;

// subscription is the event descriptor that poll_oneoff reads. This definition
// covers the clock event alone, so it names the fields of the union member
// subscription_u.clock directly. The padding keeps each field at the offset the
// WASI ABI gives it.
typedef struct {
    uint64_t userdata;
    uint8_t tag;  // 0 selects a clock event
    uint8_t pad1[7];
    uint32_t clock_id;
    uint8_t pad2[4];
    uint64_t timeout;
    uint64_t precision;
    uint16_t flags;  // 0 selects a relative timeout
    uint8_t pad3[6];
} subscription;

// event is the result descriptor that poll_oneoff writes. The test run reads no
// field of it, so this definition reserves the 32 bytes of the WASI ABI alone.
typedef struct {
    uint8_t reserved[32];
} event;

// Clock identifiers of the WASI ABI.
#define wasi_clock_realtime 0
#define wasi_clock_monotonic 1

// wasi_fd_write writes the buffers to the file descriptor and stores the number
// of bytes written in nwritten. It returns 0 on success.
__attribute__((import_module("wasi_snapshot_preview1"), import_name("fd_write"))) extern uint32_t
wasi_fd_write(uint32_t fd, const ciovec* iovs, uint32_t iovs_len, uint32_t* nwritten);

// wasi_random_get fills the buffer with len random bytes.
// It returns 0 on success.
__attribute__((import_module("wasi_snapshot_preview1"), import_name("random_get"))) extern uint32_t
wasi_random_get(uint8_t* buf, uint32_t len);

// wasi_clock_time_get stores the reading of the clock in time, in nanoseconds.
// It returns 0 on success.
__attribute__((import_module("wasi_snapshot_preview1"),
               import_name("clock_time_get"))) extern uint32_t
wasi_clock_time_get(uint32_t clock_id, uint64_t precision, uint64_t* time);

// wasi_poll_oneoff waits for one of the subscriptions and stores the events it
// receives in out. It returns 0 on success.
__attribute__((import_module("wasi_snapshot_preview1"), import_name("poll_oneoff"))) extern uint32_t
wasi_poll_oneoff(const subscription* in, event* out, uint32_t nsubscriptions, uint32_t* nevents);

// so_write_out writes size bytes to the standard output of the WASI host.
so_int so_write_out(const uint8_t* buf, so_int size) {
    ciovec iov = {.buf = buf, .len = (uint32_t)size};
    uint32_t written = 0;
    if (wasi_fd_write(1, &iov, 1, &written) != 0) {
        return 0;
    }
    return (so_int)written;
}

// so_crand_read fills buf with size random bytes from the WASI host.
so_int so_crand_read(uint8_t* buf, so_int size) {
    if (wasi_random_get(buf, (uint32_t)size) != 0) {
        return 0;
    }
    return size;
}

// so_time_wall reads the real time clock of the WASI host.
so_R_i64_i32 so_time_wall(void) {
    uint64_t ns = 0;
    if (wasi_clock_time_get(wasi_clock_realtime, 1000, &ns) != 0) {
        return (so_R_i64_i32){.val = 0, .val2 = 0};
    }
    return (so_R_i64_i32){.val = (int64_t)(ns / 1000000000), .val2 = (int32_t)(ns % 1000000000)};
}

// so_time_mono reads the monotonic clock of the WASI host.
int64_t so_time_mono(void) {
    uint64_t ns = 0;
    if (wasi_clock_time_get(wasi_clock_monotonic, 1000, &ns) != 0) {
        return 0;
    }
    // Time.Now reads 0 as the absence of a monotonic clock.
    if (ns == 0) {
        return 1;
    }
    return (int64_t)ns;
}

// so_time_sleep waits for a relative timeout
// on the monotonic clock of the WASI host.
void so_time_sleep(int64_t ns) {
    subscription sub = {0};
    sub.clock_id = wasi_clock_monotonic;
    sub.timeout = (uint64_t)ns;
    sub.precision = 1;

    event ev = {0};
    uint32_t count = 0;
    wasi_poll_oneoff(&sub, &ev, 1, &count);
}
