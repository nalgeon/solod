#include <stdarg.h>
#include "so/builtin/builtin.h"
#include "so/io/io.h"

extern so_Error fmt_ErrPrint;  // print failure
extern so_Error fmt_ErrScan;   // scan failure

// The print family is nodecay, so every argument arrives as a So type.

// Print writes its arguments to standard output, separated by spaces.
// It returns the number of bytes written and any write error encountered.
//
// The last argument marks the end of the list. No real So string has a
// negative length, so an argument cannot collide with the marker. The ##
// deletes the comma when the argument list is empty.
#define fmt_Print(...) fmt_print(false, ##__VA_ARGS__, (so_String){NULL, -1})
// Println is like Print but adds a newline at the end.
#define fmt_Println(...) fmt_print(true, ##__VA_ARGS__, (so_String){NULL, -1})
so_R_int_err fmt_print(int newline, ...);

// Printf formats according to a format specifier and writes to standard output.
// It returns the number of bytes written and any write error encountered.
so_R_int_err fmt_Printf(so_String format, ...);

// Sprintf formats according to a format specifier and returns the resulting string.
// If the output size exceeds buf length, it silently truncates the output.
so_String fmt_Sprintf(so_Slice buf, so_String format, ...);

// vsprintf is like Sprintf but takes a va_list instead of a variable number of arguments.
// It uses up the va_list, so the caller must pass it to va_end
// and must not read it again.
so_String fmt_vsprintf(so_Slice buf, so_String format, va_list ap);

// Fprintf formats according to a format specifier and writes to w.
// It returns the number of bytes written and any write error encountered.
so_R_int_err fmt_Fprintf(io_Writer w, so_String format, ...);

// The scan family reads through the stdio of the host, so it needs a hosted
// environment. A freestanding call panics.

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
