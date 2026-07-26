#ifdef so_build_hosted

#include <ctype.h>
#include <math.h>

#else

#define NAN (0.0 / 0.0)
static inline int isalpha(int ch) {
    return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z');
}
static inline double sqrt(double x) {
    if (x < 0) return NAN;
    if (x == 0) return 0;
    double guess = x / 2;
    for (int i = 0; i < 10; i++) {
        guess = (guess + x / guess) / 2;
    }
    return guess;
}

#endif  // so_build_hosted

static inline const char* get_cstring(const char* s) {
    return s;
}

static inline size_t str_len(const char* s) {
    size_t n = 0;
    while (s[n]) n++;
    return n;
}

// Returns the index of ch in s, or -1 if there is none.
static inline so_ssize_t str_index(const char* s, char ch) {
    for (so_ssize_t i = 0; s[i]; i++) {
        if (s[i] == ch) return i;
    }
    return -1;
}

static inline ptrdiff_t ptr_diff(const char* a, const char* b) {
    return a - b;
}

static inline intptr_t ptr_addr(const void* p) {
    return (intptr_t)p;
}

static inline long double ld_half(long double x) {
    return x / 2;
}
