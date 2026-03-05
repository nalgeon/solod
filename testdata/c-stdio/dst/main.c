#include "main.h"

// -- Implementation --

int main(void) {
    {
        // Output stream.
        stdio_File* f = stdio_Fopen(so_str("/tmp/test.txt"), so_str("w"));
        if (f == NULL) {
            so_panic("failed to open file");
        }
        stdio_Fputs(so_str("hello"), f);
        stdio_Fputc(10, f);
        stdio_Fflush(f);
        so_Slice buf = (so_Slice){(uint8_t[64]){0}, 64, 64};
        stdio_Fwrite(&so_index(uint8_t, buf, 0), 1, 64, f);
        stdio_Fclose(f);
    }
    {
        // Input stream.
        stdio_File* f = stdio_Fopen(so_str("/tmp/test.txt"), so_str("r"));
        if (f == NULL) {
            so_panic("failed to open file");
        }
        so_int ch = stdio_Fgetc(f);
        if (ch == stdio_EOF) {
            so_panic("unexpected EOF");
        }
        so_Slice buf = (so_Slice){(uint8_t[64]){0}, 64, 64};
        stdio_Fseek(f, 0, 0);
        stdio_Fgets(&so_index(uint8_t, buf, 0), 64, f);
        stdio_Fread(&so_index(uint8_t, buf, 0), 1, 64, f);
        if (stdio_Feof(f)) {
            so_panic("unexpected EOF");
        }
        if (stdio_Ferror(f)) {
            so_panic("stream error");
        }
        stdio_Fclose(f);
    }
}
