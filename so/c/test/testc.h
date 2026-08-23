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

static inline unsigned sum_bytes(const unsigned char* p, size_t n) {
    unsigned sum = 0;
    for (size_t i = 0; i < n; i++) sum += p[i];
    return sum;
}

// Returns the index of the first item accepted by match, or -1 if there is none.
static inline so_ssize_t find_first(const void* items, size_t count, size_t size,
                                    bool (*match)(const void*)) {
    const unsigned char* base = (const unsigned char*)items;
    for (size_t i = 0; i < count; i++) {
        if (match(base + i * size)) return (so_ssize_t)i;
    }
    return -1;
}
