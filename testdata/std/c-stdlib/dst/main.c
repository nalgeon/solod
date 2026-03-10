#include "main.h"

// -- Implementation --

int main(void) {
    {
        // Constants.
        so_int status = stdlib_ExitSuccess;
        if (status == stdlib_ExitFailure) {
            so_panic("unexpected failure");
        }
    }
    {
        // String-to-number conversion.
        so_int n = stdlib_Atoi("42");
        if (n != 42) {
            so_panic("want n == 42");
        }
        double f = stdlib_Atof("3.14");
        if (f < 3.0) {
            so_panic("want f >= 3.0");
        }
    }
    {
        // Memory management.
        void* ptr = stdlib_Malloc(unsafe_Sizeof((so_int)(0)));
        if (ptr == NULL) {
            so_panic("malloc failed");
        }
        stdlib_Free(ptr);
        ptr = stdlib_Calloc(10, unsafe_Sizeof((so_int)(0)));
        if (ptr == NULL) {
            so_panic("calloc failed");
        }
        ptr = stdlib_Realloc(ptr, 20 * unsafe_Sizeof((so_int)(0)));
        if (ptr == NULL) {
            so_panic("realloc failed");
        }
        stdlib_Free(ptr);
    }
    {
        // Environment.
        so_byte* env = stdlib_Getenv("PATH");
        if (env == NULL) {
            so_panic("PATH not set");
        }
    }
    {
        // Exit (must be last).
        stdlib_Exit(0);
    }
}
