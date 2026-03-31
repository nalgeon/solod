#include <stdarg.h>
#include <stdio.h>
#include "so/builtin/builtin.h"
#include "so/io/io.h"

// BufSize is the size of the internal formatting buffer in bytes.
#define fmt_BufSize 1024

extern so_Error fmt_ErrPrint;  // print failure
extern so_Error fmt_ErrScan;   // scan failure
extern so_Error fmt_ErrSize;   // buffer size exceeded

// Buffer is a fixed-size stack-allocated buffer
// for formatted output and scanning.
typedef struct {
    void* Ptr;
    size_t Len;
} fmt_Buffer;

// NewBuffer creates a new stack-allocated Buffer of the given size.
#define fmt_NewBuffer(size) ({                             \
    so_int _isize = (size);                                \
    if (_isize < 0) {                                      \
        so_panic("fmt: negative buffer size");             \
    }                                                      \
    size_t _bsize = (size_t)_isize;                        \
    (fmt_Buffer){.Ptr = so_alloca(_bsize), .Len = _bsize}; \
})

// Print writes its arguments to standard output, separated by spaces.
// It returns the number of bytes written and any write error encountered.
#define fmt_Print(...) fmt_print(false, __VA_ARGS__, NULL)
// Println is like Print but adds a newline at the end.
#define fmt_Println(...) fmt_print(true, __VA_ARGS__, NULL)
so_R_int_err fmt_print(int newline, ...);

// Printf formats according to a format specifier and writes to standard output.
// It returns the number of bytes written and any write error encountered.
so_R_int_err fmt_Printf(const char* format, ...);

// Sprintf formats according to a format specifier and returns the resulting string.
// If the output size exceeds buf length, it silently truncates the output.
so_String fmt_Sprintf(fmt_Buffer buf, const char* format, ...);

// Fprintf formats according to a format specifier and writes to w.
// It returns the number of bytes written and any write error encountered.
// Returns [ErrSize] if the output size exceeds BufSize.
so_R_int_err fmt_Fprintf(io_Writer w, const char* format, ...);

// Scanf scans text read from standard input, storing successive
// space-separated values into successive arguments as determined by the format.
// It returns the number of items successfully scanned.
so_R_int_err fmt_Scanf(const char* format, ...);

// Sscanf scans the argument string, storing successive space-separated
// values into successive arguments as determined by the format.
// It returns the number of items successfully scanned.
so_R_int_err fmt_Sscanf(const char* str, const char* format, ...);

// Fscanf scans text read from r, storing successive space-separated
// values into successive arguments as determined by the format.
// It returns the number of items successfully scanned.
so_R_int_err fmt_Fscanf(io_Reader r, const char* format, ...);
