//go:build ignore
#include "testing.h"

#ifdef so_build_hosted

// exitFail ends the program with a failure status.
static void testing_exitFail(void) {
    exit(1);
}

#else

// A freestanding host has no exit status, so exitFail traps.
static void testing_exitFail(void) {
    __builtin_trap();
}

#endif  // so_build_hosted

void testing_T_Errorf(void* self, so_String format, ...) {
    char buf[testing_msgSize];
    va_list args;
    va_start(args, format);
    so_String msg = fmt_vsprintf((so_Slice){buf, testing_msgSize, testing_msgSize}, format, args);
    va_end(args);
    testing_T_Error(self, msg);
}

void testing_T_Fatalf(void* self, so_String format, ...) {
    char buf[testing_msgSize];
    va_list args;
    va_start(args, format);
    so_String msg = fmt_vsprintf((so_Slice){buf, testing_msgSize, testing_msgSize}, format, args);
    va_end(args);
    testing_T_Fatal(self, msg);
}
