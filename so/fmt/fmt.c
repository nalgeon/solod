//go:build ignore
#include "fmt.h"

so_Error fmt_ErrPrint = errors_New("print failure");
so_Error fmt_ErrScan = errors_New("scan failure");

// --- Output ---

#ifdef so_build_hosted

// writeOut writes p to the standard output.
static so_int fmt_writeOut(so_Slice p) {
    if (p.len == 0) {
        return 0;
    }
    return (so_int)fwrite(p.ptr, 1, (size_t)p.len, stdout);
}

#else

// A freestanding host has no standard output, so writeOut drops the bytes. The
// full count is what io.Discard reports as well.
static so_int fmt_writeOut(so_Slice p) {
    return p.len;
}

#endif  // so_build_hosted

// --- Argument collection ---

// argCount returns the number of arguments that format needs.
static so_int fmt_argCount(so_String format) {
    return fmt_argKinds(format, (so_Slice){});
}

// fill pulls the arguments that format needs from ap into args, which holds
// fmt_argCount(format) values. The So side gives the kind of each argument, so
// the C type of every va_arg follows the verb table in print.go.
//
// A va_list argument is used up here. The caller must pass it to va_end and
// must not read it again.
static void fmt_fill(so_String format, so_Slice args, va_list ap) {
    so_Slice kinds = so_make_slice(so_int, args.len, args.len);
    fmt_argKinds(format, kinds);

    fmt_arg* dst = args.ptr;
    for (so_int i = 0; i < args.len; i++) {
        so_int kind = ((so_int*)kinds.ptr)[i];
        fmt_arg a = {.kind = kind};
        if (kind == fmt_kindInt) {
            a.i = va_arg(ap, so_int);
        } else if (kind == fmt_kindUint) {
            a.i = (so_int)va_arg(ap, so_uint);
        } else if (kind == fmt_kindFloat) {
            a.f = va_arg(ap, double);
        } else if (kind == fmt_kindRune) {
            a.i = va_arg(ap, so_int);
        } else if (kind == fmt_kindString) {
            a.s = va_arg(ap, so_String);
        } else if (kind == fmt_kindBool) {
            a.i = va_arg(ap, so_int) != 0;
        } else if (kind == fmt_kindPtr) {
            a.i = (so_int)(uintptr_t)va_arg(ap, void*);
        }
        dst[i] = a;
    }
}

// --- Print ---

so_R_int_err fmt_print(int newline, ...) {
    // Print and Println have no format string. The arguments are strings up to
    // the marker that the macro adds, so one pass over a copy counts them. The
    // marker is the string with a negative length, see fmt.h.
    va_list ap, cp;
    va_start(ap, newline);
    va_copy(cp, ap);
    so_int count = 0;
    while (va_arg(cp, so_String).len >= 0) {
        count++;
    }
    va_end(cp);

    so_Slice args = so_make_slice(fmt_arg, count, count);
    fmt_arg* dst = args.ptr;
    for (so_int i = 0; i < count; i++) {
        so_String s = va_arg(ap, so_String);
        dst[i] = (fmt_arg){.kind = fmt_kindString, .s = s};
    }
    va_end(ap);

    return fmt_vjoin(fmt_Output, args, newline != 0);
}

so_R_int_err fmt_Printf(so_String format, ...) {
    so_int count = fmt_argCount(format);
    so_Slice args = so_make_slice(fmt_arg, count, count);

    va_list ap;
    va_start(ap, format);
    fmt_fill(format, args, ap);
    va_end(ap);

    return fmt_vfprint(fmt_Output, format, args);
}

so_String fmt_vsprintf(so_Slice buf, so_String format, va_list ap) {
    so_int count = fmt_argCount(format);
    so_Slice args = so_make_slice(fmt_arg, count, count);
    fmt_fill(format, args, ap);
    return fmt_vsprint(buf, format, args);
}

so_String fmt_Sprintf(so_Slice buf, so_String format, ...) {
    va_list ap;
    va_start(ap, format);
    so_String s = fmt_vsprintf(buf, format, ap);
    va_end(ap);
    return s;
}

so_R_int_err fmt_Fprintf(io_Writer w, so_String format, ...) {
    so_int count = fmt_argCount(format);
    so_Slice args = so_make_slice(fmt_arg, count, count);

    va_list ap;
    va_start(ap, format);
    fmt_fill(format, args, ap);
    va_end(ap);

    return fmt_vfprint(w, format, args);
}

// --- Scan ---

#ifdef so_build_hosted

// scanSize is the size of the buffer that Fscanf reads a line into. A longer
// line is truncated.
#define fmt_scanSize 1024

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
    char buf[fmt_scanSize];
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

#else

// A freestanding host has no stdio, so the scan family panics there.

so_R_int_err fmt_Scanf(const char* format, ...) {
    (void)format;
    so_panic("fmt: scanning requires a hosted environment");
    return (so_R_int_err){};
}

so_R_int_err fmt_Sscanf(const char* str, const char* format, ...) {
    (void)str;
    (void)format;
    so_panic("fmt: scanning requires a hosted environment");
    return (so_R_int_err){};
}

so_R_int_err fmt_Fscanf(io_Reader r, const char* format, ...) {
    (void)r;
    (void)format;
    so_panic("fmt: scanning requires a hosted environment");
    return (so_R_int_err){};
}

#endif  // so_build_hosted
