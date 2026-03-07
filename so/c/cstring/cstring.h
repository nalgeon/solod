#include <string.h>

#define cstring_Memcpy(dst, src, n) memcpy(dst, src, n)
#define cstring_Memmove(dst, src, n) memmove(dst, src, n)
#define cstring_Memset(ptr, value, n) memset(ptr, value, n)
#define cstring_Memcmp(a, b, n) memcmp(a, b, n)
