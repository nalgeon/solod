#include <stdio.h>

#define stdio_EOF EOF

#define stdio_SeekSet SEEK_SET
#define stdio_SeekCur SEEK_CUR
#define stdio_SeekEnd SEEK_END

#define stdio_File FILE

#define stdio_Stdin stdin
#define stdio_Stdout stdout
#define stdio_Stderr stderr

#define stdio_Fopen(path, mode) fopen(path.ptr, mode.ptr)
#define stdio_Fclose(stream) fclose(stream)
#define stdio_Fflush(stream) fflush(stream)
#define stdio_Fseek(stream, offset, whence) fseek(stream, offset, whence)

#define stdio_Fgetc(stream) fgetc(stream)
#define stdio_Fputc(ch, stream) fputc(ch, stream)

#define stdio_Fgets(s, n, stream) fgets((char*)s, n, stream)
#define stdio_Fputs(s, stream) fputs(s.ptr, stream)

#define stdio_Fread(ptr, size, count, stream) fread(ptr, size, count, stream)
#define stdio_Fwrite(ptr, size, count, stream) fwrite(ptr, size, count, stream)

#define stdio_Feof(stream) feof(stream)
#define stdio_Ferror(stream) ferror(stream)
