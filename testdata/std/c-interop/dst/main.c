#include "main.h"
#include <string.h>

// -- Implementation --

int main(void) {
    {
        // c.String: convert C string to So string.
        so_byte* ptr = stdlib_Getenv("PATH");
        so_String path = c_String(ptr);
        if (so_len(path) == 0) {
            so_panic("want non-empty PATH");
        }
    }
    {
        // c.String: nil pointer returns empty string.
        so_byte* ptr = stdlib_Getenv("SOLOD_NONEXISTENT_VAR");
        so_String s = c_String(ptr);
        if (so_len(s) != 0) {
            so_panic("want empty string for nil");
        }
    }
    {
        // c.Bytes: wrap a raw buffer into []byte.
        void* buf = stdlib_Malloc(4);
        if (buf == NULL) {
            so_panic("malloc failed");
        }
        so_byte* ptr = (so_byte*)buf;
        *ptr = 'H';
        so_Slice slice = c_Bytes(ptr, 4);
        if (so_len(slice) != 4) {
            so_panic("want len == 4");
        }
        if (so_at(so_byte, slice, 0) != 'H') {
            so_panic("want slice[0] == 'H'");
        }
        stdlib_Free(buf);
    }
    {
        // Passing (char*) strings to C functions.
        so_byte buf[64] = {0};
        strcat(c_CharPtr(&buf[0]), "Hello, ");
        strcat(c_CharPtr(&buf[0]), "world!");
        so_String s = c_String(&buf[0]);
        so_println("%.*s", s.len, s.ptr);
    }
    {
        // Returning (char*) strings from C functions.
        so_byte buf[64] = {0};
        strcat(c_CharPtr(&buf[0]), "Hello, ");
        so_String s = c_String(strcat(c_CharPtr(&buf[0]), "world!"));
        so_println("%.*s", s.len, s.ptr);
    }
}
