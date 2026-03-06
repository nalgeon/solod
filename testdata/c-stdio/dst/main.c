#include "main.h"

// -- Implementation --

int main(void) {
    {
        // Output stream.
        stdio_File* f = stdio_Fopen("/tmp/test.txt", "w");
        if (f == NULL) {
            so_panic("failed to open file");
        }
        stdio_Fputs("hello", f);
        stdio_Fputc(10, f);
        stdio_Fflush(f);
        uint8_t buf[64] = {0};
        stdio_Fwrite(&buf[0], 1, 64, f);
        stdio_Fclose(f);
    }
    {
        // Input stream.
        stdio_File* f = stdio_Fopen("/tmp/test.txt", "r");
        if (f == NULL) {
            so_panic("failed to open file");
        }
        so_int ch = stdio_Fgetc(f);
        if (ch == stdio_EOF) {
            so_panic("unexpected EOF");
        }
        uint8_t buf[64] = {0};
        stdio_Fseek(f, 0, 0);
        stdio_Fgets(&buf[0], 64, f);
        stdio_Fread(&buf[0], 1, 64, f);
        so_int pos = stdio_Ftell(f);
        if (pos < 0) {
            so_panic("ftell error");
        }
        if (stdio_Feof(f)) {
            so_panic("unexpected EOF");
        }
        if (stdio_Ferror(f)) {
            so_panic("stream error");
        }
        stdio_Fclose(f);
    }
    {
        // Formatted output.
        stdio_Printf("hello %d\n", 42);
        stdio_Fprintf(stdio_Stdout, "value: %d\n", 100);
        uint8_t buf[64] = {0};
        stdio_Snprintf(&buf[0], 64, "count: %d", 10);
    }
    {
        // Formatted input.
        int32_t n = 0;
        stdio_Sscanf("42", "%d", &n);
    }
}
