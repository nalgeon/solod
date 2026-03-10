#include "main.h"

// -- Implementation --

int main(void) {
    uintptr_t size = 4 * unsafe_Sizeof((so_byte)(0));
    // Memset: fill memory with a value.
    void* buf = stdlib_Malloc(size);
    if (buf == NULL) {
        so_panic("malloc failed");
    }
    cstring_Memset(buf, 65, size);
    // Memcpy: copy to a new buffer.
    void* buf2 = stdlib_Malloc(size);
    if (buf2 == NULL) {
        so_panic("malloc failed");
    }
    cstring_Memcpy(buf2, buf, size);
    // Memcmp: compare the two buffers.
    so_int result = cstring_Memcmp(buf, buf2, size);
    if (result != 0) {
        so_panic("want result == 0");
    }
    // Memmove: overlapping copy within buf.
    cstring_Memmove(buf, buf, size);
    result = cstring_Memcmp(buf, buf2, size);
    if (result != 0) {
        so_panic("want result == 0");
    }
    stdlib_Free(buf);
    stdlib_Free(buf2);
}
