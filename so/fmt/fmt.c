//go:build ignore
#include "fmt.h"

so_Error fmt_ErrPrint = errors_New("print failure");
so_Error fmt_ErrScan = errors_New("scan failure");
so_Error fmt_ErrSize = errors_New("buffer size exceeded");

so_R_int_err fmt_print(int newline, ...) {
    int total = 0;
    so_Error err = {};
    va_list args;

    va_start(args, newline);
    const char* arg = va_arg(args, const char*);
    while (arg != NULL) {
        int n = printf("%s", arg);
        if (n < 0) {
            err = fmt_ErrPrint;
            break;
        }
        total += n;
        arg = va_arg(args, const char*);
        if (arg != NULL) {
            putchar(' ');
            total++;
        } else if (newline) {
            putchar('\n');
            total++;
        }
    }
    va_end(args);

    return (so_R_int_err){.val = total, .err = err};
}

so_R_int_err fmt_Printf(const char* format, ...) {
    va_list args;
    va_start(args, format);
    int n = vprintf(format, args);
    va_end(args);
    so_Error err = n < 0 ? fmt_ErrPrint : (so_Error){};
    return (so_R_int_err){.val = n, .err = err};
}

so_String fmt_vsprintf(fmt_Buffer buf, const char* format, va_list args) {
    int n = vsnprintf(buf.Ptr, (size_t)buf.Len, format, args);
    if (n < 0) {
        n = 0;  // treat encoding errors as empty output
    } else if (n >= buf.Len) {
        n = buf.Len - 1;  // truncate output to fit buffer
    }
    return (so_String){.ptr = buf.Ptr, .len = n};
}

so_String fmt_Sprintf(fmt_Buffer buf, const char* format, ...) {
    va_list args;
    va_start(args, format);
    so_String s = fmt_vsprintf(buf, format, args);
    va_end(args);
    return s;
}

so_R_int_err fmt_Fprintf(io_Writer w, const char* format, ...) {
    char buf[fmt_BufSize];

    va_list args;
    va_start(args, format);
    int n = vsnprintf(buf, sizeof(buf), format, args);
    va_end(args);
    if (n < 0) {
        return (so_R_int_err){.err = fmt_ErrPrint};
    }

    if ((size_t)n >= sizeof(buf)) {
        return (so_R_int_err){.val = n, .err = fmt_ErrSize};
    }
    so_Slice slice = {.ptr = buf, .len = n, .cap = n};
    return w.Write(w.self, slice);
}

so_R_int_err fmt_Scanf(const char* format, ...) {
    va_list args;
    va_start(args, format);
    int n = vscanf(format, args);
    va_end(args);
    so_Error err = n < 0 ? fmt_ErrScan : (so_Error){};
    return (so_R_int_err){.val = n, .err = err};
}

so_R_int_err fmt_Sscanf(const char* str, const char* format, ...) {
    va_list args;
    va_start(args, format);
    int n = vsscanf(str, format, args);
    va_end(args);
    so_Error err = n < 0 ? fmt_ErrScan : (so_Error){};
    return (so_R_int_err){.val = n, .err = err};
}

so_R_int_err fmt_Fscanf(io_Reader r, const char* format, ...) {
    char buf[fmt_BufSize];
    so_int len = sizeof(buf) - 1;  // leave space for null terminator
    so_Slice slice = {.ptr = buf, .len = len, .cap = len};
    so_R_int_err res = r.Read(r.self, slice);
    if (res.err.self) {
        return (so_R_int_err){.err = res.err};
    }
    buf[res.val] = '\0';

    va_list args;
    va_start(args, format);
    int n = vsscanf(buf, format, args);
    va_end(args);

    so_Error err = n < 0 ? fmt_ErrScan : (so_Error){};
    return (so_R_int_err){.val = n, .err = err};
}
