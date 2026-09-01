#define SUB_TAG so_str("sub")

typedef struct {
    void (*Write)(const char* format, ...);
} Stream;

static inline void Discard(const char* format, ...) {
    (void)format;
}
