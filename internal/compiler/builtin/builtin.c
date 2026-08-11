#include <stdarg.h>
#include "builtin.h"

#ifdef so_build_hosted
#include <stdio.h>
#endif

#if defined(so_build_hosted) && SO_PANIC_MODE == SO_PANIC_TRACE
// backtrace lives in execinfo.h, a glibc/BSD extension. musl (Alpine) omits it.
#if defined(__has_include) && __has_include(<execinfo.h>)
#include <execinfo.h>
#include <unistd.h>
// print_trace writes a symbolized backtrace of the current call stack to
// stderr. The top frame (this function itself) is dropped so the panic site
// appears first. Needs -rdynamic for symbol names; frames-only otherwise.
void so_print_trace(void) {
    void* frames[64];
    int n = backtrace(frames, (int)(sizeof(frames) / sizeof(frames[0])));
    if (n > 1) {
        backtrace_symbols_fd(frames + 1, n - 1, STDERR_FILENO);
    } else {
        backtrace_symbols_fd(frames, n, STDERR_FILENO);
    }
}
#else
void so_print_trace(void) {}
#endif  // __has_include(<execinfo.h>)
#endif  // so_build_hosted && SO_PANIC_MODE == SO_PANIC_TRACE

// A memory-access fault becomes a panic in every hosted mode except abort,
// which leaves the fault alone so it can dump a core. The handler is POSIX
// only (sigaction/SIGBUS); Windows has no handler and faults like C.
#if defined(so_build_hosted) && !defined(so_build_windows) && SO_PANIC_MODE != SO_PANIC_ABORT
#include <signal.h>
#include <unistd.h>

// so_fault_handler turns a SIGSEGV or SIGBUS into a panic. A fault within the
// first 64KB is reported as a nil pointer dereference (a nil base plus a small
// field or index offset); any other address as an invalid access. In trace
// mode it prints a backtrace, led by this handler and the signal-trampoline
// frame before the faulting frame.
static void so_fault_handler(int sig, siginfo_t* info, void* ctx) {
    // Everything here is async-signal-safe: constant messages via write(), the
    // _fd backtrace variant, and _exit(). No fprintf, so a fault raised while the
    // stdio or malloc lock is held reports instead of deadlocking.
    (void)sig;
    (void)ctx;
    static const char nil_msg[] = "panic: nil pointer dereference\n";
    static const char bad_msg[] = "panic: invalid memory address\n";
    const char* msg = nil_msg;
    size_t len = sizeof(nil_msg) - 1;
    if ((uintptr_t)info->si_addr >= 0x10000) {
        msg = bad_msg;
        len = sizeof(bad_msg) - 1;
    }
    ssize_t written = write(STDERR_FILENO, msg, len);
    (void)written;
#if SO_PANIC_MODE == SO_PANIC_TRACE
    so_print_trace();
#endif
    _exit(1);
}

// so_install_fault_handler registers so_fault_handler for SIGSEGV and SIGBUS
// before main runs. The handler runs on the normal stack, not a signal stack:
// backtrace() cannot unwind from an alternate stack on macOS. This means a
// stack overflow, which exhausts that stack, cannot be reported.
__attribute__((constructor)) static void so_install_fault_handler(void) {
    struct sigaction sa = {};
    sa.sa_sigaction = so_fault_handler;
    sa.sa_flags = SA_SIGINFO;
    sigemptyset(&sa.sa_mask);
    sigaction(SIGSEGV, &sa, NULL);
    sigaction(SIGBUS, &sa, NULL);
}
#endif

// Command-line arguments, populated by main().
so_Slice os_Args = {};

// utf8_decode decodes one UTF-8 rune from string s at byte offset i.
// Stores the byte width in *w.
// Returns the decoded rune, or 0xFFFD for invalid UTF-8.
so_rune so_utf8_decode(so_String s, so_int i, so_int* w) {
    const uint8_t* p = (const uint8_t*)s.ptr + i;
    so_int remaining = s.len - i;
    if (remaining <= 0) {
        *w = 0;
        return 0xFFFD;
    }

    uint8_t b = p[0];
    if (b < 0x80) {
        *w = 1;
        return (so_rune)b;
    }
    if ((b & 0xE0) == 0xC0 && remaining >= 2 &&
        (p[1] & 0xC0) == 0x80) {
        *w = 2;
        return ((so_rune)(b & 0x1F) << 6) |
               ((so_rune)(p[1] & 0x3F));
    }
    if ((b & 0xF0) == 0xE0 && remaining >= 3 &&
        (p[1] & 0xC0) == 0x80 && (p[2] & 0xC0) == 0x80) {
        *w = 3;
        return ((so_rune)(b & 0x0F) << 12) |
               ((so_rune)(p[1] & 0x3F) << 6) |
               ((so_rune)(p[2] & 0x3F));
    }
    if ((b & 0xF8) == 0xF0 && remaining >= 4 &&
        (p[1] & 0xC0) == 0x80 &&
        (p[2] & 0xC0) == 0x80 &&
        (p[3] & 0xC0) == 0x80) {
        *w = 4;
        return ((so_rune)(b & 0x07) << 18) |
               ((so_rune)(p[1] & 0x3F) << 12) |
               ((so_rune)(p[2] & 0x3F) << 6) |
               ((so_rune)(p[3] & 0x3F));
    }

    *w = 1;
    return 0xFFFD;
}

// string_runes_impl decodes UTF-8 string bytes into a rune buffer.
so_Slice so_string_runes_impl(so_String s, so_rune* buf) {
    so_int n = 0;
    for (so_int i = 0; i < s.len;) {
        so_int w = 0;
        buf[n++] = so_utf8_decode(s, i, &w);
        i += w;
    }
    return (so_Slice){buf, n, n};
}

// utf8_encode encodes a single rune into buf (up to 4 bytes).
// Returns the number of bytes written.
so_int so_utf8_encode(so_rune r, char* buf) {
    if (r < 0x80) {
        buf[0] = (char)r;
        return 1;
    }
    if (r < 0x800) {
        buf[0] = (char)(0xC0 | (r >> 6));
        buf[1] = (char)(0x80 | (r & 0x3F));
        return 2;
    }
    if (r < 0x10000) {
        buf[0] = (char)(0xE0 | (r >> 12));
        buf[1] = (char)(0x80 | ((r >> 6) & 0x3F));
        buf[2] = (char)(0x80 | (r & 0x3F));
        return 3;
    }
    buf[0] = (char)(0xF0 | (r >> 18));
    buf[1] = (char)(0x80 | ((r >> 12) & 0x3F));
    buf[2] = (char)(0x80 | ((r >> 6) & 0x3F));
    buf[3] = (char)(0x80 | (r & 0x3F));
    return 4;
}

// runes_string_impl encodes runes into a UTF-8 buffer and returns a string.
so_String so_runes_string_impl(so_Slice rs, char* buf) {
    so_int pos = 0;
    so_rune* runes = (so_rune*)rs.ptr;
    for (so_int i = 0; i < rs.len; i++) {
        pos += so_utf8_encode(runes[i], buf + pos);
    }
    return (so_String){buf, pos};
}

// map_nextpow2 rounds up to the next power of 2.
so_int so_map_nextpow2(so_int n) {
    if (n == 0) return 1;
    n--;
    n |= n >> 1;
    n |= n >> 2;
    n |= n >> 4;
    n |= n >> 8;
    n |= n >> 16;
#if so_int_bits == 64
    n |= n >> 32;
#endif
    return n + 1;
}

// map_find looks up a key in the map.
// If found, copies the value to out_val (when non-NULL) and sets *found = true.
// If not found, sets *found = false and leaves out_val unchanged.
void so_map_find(const so_Map* m, const void* key, size_t key_size,
                 void* out_val, size_t val_size,
                 uint64_t hash, bool* found,
                 bool (*eq)(const void*, const void*, size_t)) {
    if (m->cap == 0) {
        *found = false;
        return;
    }
    size_t mask = m->cap - 1;
    size_t step = (size_t)(hash >> 32) | 1;
    size_t idx = (size_t)hash & mask;
    for (so_int p = 0; p < m->cap; p++) {
        if (!m->used[idx]) {
            *found = false;
            return;
        }
        if (eq((const char*)m->keys + idx * key_size, key, key_size)) {
            if (out_val) {
                memcpy(out_val, (const char*)m->vals + idx * val_size, val_size);
            }
            *found = true;
            return;
        }
        idx = (idx + step) & mask;
    }
    *found = false;
}

// map_set_impl inserts or updates a key-value pair in the map.
// Panics if the map is full and the key is not found.
void so_map_set_impl(so_Map* m, const void* key, size_t key_size,
                     const void* val, size_t val_size,
                     uint64_t hash,
                     bool (*eq)(const void*, const void*, size_t)) {
    size_t mask = m->cap - 1;
    size_t step = (size_t)(hash >> 32) | 1;
    size_t idx = (size_t)hash & mask;
    for (so_int p = 0;; p++) {
        if (p >= m->cap)
            so_panic("map: out of capacity");
        if (!m->used[idx]) {
            memcpy((char*)m->keys + idx * key_size, key, key_size);
            memcpy((char*)m->vals + idx * val_size, val, val_size);
            m->used[idx] = 1;
            m->len++;
            return;
        }
        if (eq((const char*)m->keys + idx * key_size, key, key_size)) {
            memcpy((char*)m->vals + idx * val_size, val, val_size);
            return;
        }
        idx = (idx + step) & mask;
    }
}

#ifdef so_build_hosted

// print writes the formatted string to stdout.
// Returns the number of bytes written.
int so_print(const char* format, ...) {
    va_list args;
    va_start(args, format);
    int n = vprintf(format, args);
    va_end(args);
    return n;
}

// println writes the formatted string to stdout with a newline.
// Returns the number of bytes written.
int so_println(const char* format, ...) {
    va_list args;
    va_start(args, format);
    int n = vprintf(format, args);
    va_end(args);
    putchar('\n');
    return n + 1;
}

#else

int so_print(const char* format, ...) {
    (void)format;
    return 0;
}
int so_println(const char* format, ...) {
    (void)format;
    return 0;
}

#endif  // so_build_hosted
