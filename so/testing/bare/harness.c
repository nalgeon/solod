//go:build ignore

// This file supplies the output hook for a freestanding test run on wasm32.
// A wasm32-freestanding module can still import the fd_write function of
// wasi_snapshot_preview1, and a runtime like wasmtime supplies it. So the
// test report and the panic messages reach the terminal.

#include "so/builtin/builtin.h"

// ciovec is the buffer descriptor that fd_write reads. The WASI ABI is 32-bit,
// so both fields are 32-bit.
typedef struct {
    const uint8_t* buf;
    uint32_t len;
} ciovec;

// wasi_fd_write writes the buffers to the file descriptor and stores the number
// of bytes written in nwritten. It returns 0 on success.
__attribute__((import_module("wasi_snapshot_preview1"), import_name("fd_write"))) extern uint32_t
wasi_fd_write(uint32_t fd, const ciovec* iovs, uint32_t iovs_len, uint32_t* nwritten);

// so_write_out writes size bytes to the standard output of the WASI host.
so_int so_write_out(const uint8_t* buf, so_int size) {
    ciovec iov = {.buf = buf, .len = (uint32_t)size};
    uint32_t written = 0;
    if (wasi_fd_write(1, &iov, 1, &written) != 0) {
        return 0;
    }
    return (so_int)written;
}
