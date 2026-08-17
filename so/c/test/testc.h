#define TEST_ANSWER 42

static inline int is_alpha(int ch) {
    return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z');
}

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
