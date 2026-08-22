#include "so/builtin/builtin.h"

// SwapByte swaps n bytes between a and b.
// Panics if either a or b is nil.
//
// SwapByte temporarily allocates a buffer of size n
// on the stack, so it's not suitable for large n.
static inline void mem_SwapByte(void* a, void* b, so_int n) {
    so_assert(a != NULL, "mem: nil pointer");
    so_assert(b != NULL, "mem: nil pointer");
    so_assert(n >= 0, "mem: negative size");
    if (n == 0) return;

    size_t size = (size_t)n;
    char tmp[size];
    memcpy(tmp, a, size);
    memcpy(a, b, size);
    memcpy(b, tmp, size);
}

// A counter is one statistic of a Tracker.
//
// A target with a lock-free 64-bit atomic gets an atomic add. Every other
// target gets a plain add, which is not thread-safe.
typedef struct {
    alignas(8) uint64_t v;
} mem_counter;

#if __GCC_ATOMIC_LLONG_LOCK_FREE == 2

#define mem_counter_Load(c) \
    (__atomic_load_n(&(c)->v, __ATOMIC_SEQ_CST))

#define mem_counter_Add(c, delta) \
    ((void)__atomic_add_fetch(&(c)->v, (delta), __ATOMIC_SEQ_CST))

#else

#define mem_counter_Load(c) ((c)->v)
#define mem_counter_Add(c, delta) ((void)((c)->v += (delta)))

#endif  // __GCC_ATOMIC_LLONG_LOCK_FREE == 2

#ifndef so_build_hosted

// Bump allocator over a static buffer for freestanding environments.
// Memory is never reclaimed: free is a no-op, realloc copies into a new bump.
// Suitable for short-lived programs that don't need much memory.
// The heap is off by default, enable with -DSO_HEAP_SIZE=N.

#ifndef SO_HEAP_SIZE
#define SO_HEAP_SIZE (0)  // in bytes
#endif

#if SO_HEAP_SIZE > 0

// The whole program shares a single heap.
extern alignas(16) char so_heap[SO_HEAP_SIZE];
extern size_t so_heap_offset;

// so_heap_next rounds cur up to a 16 byte boundary and adds size to it. It
// writes the rounded offset to off and the position of the next allocation
// to next, and reports whether the allocation fits.
static inline bool so_heap_next(size_t cur, size_t size, size_t* off, size_t* next) {
    // The rounded offset can pass the end of the heap,
    // so check it before the subtraction.
    size_t offset = (cur + 15) & ~(size_t)15;
    if (offset > SO_HEAP_SIZE || size > SO_HEAP_SIZE - offset) {
        return false;
    }
    *off = offset;
    *next = offset + size;
    return true;
}

// malloc takes the next range of the heap. It uses a compare and exchange,
// not a fetch and add: one atomic operation cannot align the offset and add
// to it, and a failed add still advances the offset.
static inline void* malloc(size_t size) {
    if (size == 0) return NULL;
    size_t offset, next;
#if __GCC_ATOMIC_POINTER_LOCK_FREE == 2
    // The target has a lock-free atomic of pointer width. so_heap_offset
    // is modified in a compare-and-swap loop, so malloc is thread-safe.
    size_t cur = __atomic_load_n(&so_heap_offset, __ATOMIC_RELAXED);
    do {
        if (!so_heap_next(cur, size, &offset, &next)) return NULL;
    } while (!__atomic_compare_exchange_n(&so_heap_offset, &cur, next, false,
                                          __ATOMIC_RELAXED, __ATOMIC_RELAXED));
#else
    // The target has no lock-free atomic of pointer width.
    // so_heap_offset is modified with a plain read and write,
    // so malloc is not thread-safe.
    if (!so_heap_next(so_heap_offset, size, &offset, &next)) return NULL;
    so_heap_offset = next;
#endif
    return &so_heap[offset];
}

static inline void* calloc(size_t num, size_t size) {
    if (num != 0 && size > SIZE_MAX / num) return NULL;
    size_t total = num * size;
    void* ptr = malloc(total);
    if (ptr) memset(ptr, 0, total);
    return ptr;
}

static inline void* realloc(void* ptr, size_t new_size) {
    if (new_size == 0) return NULL;
    void* new_ptr = malloc(new_size);
    if (ptr && new_ptr) {
        // We don't track allocation sizes, so we copy new_size bytes.
        // When growing, this over-reads from the old allocation into
        // adjacent bump memory - harmless but yields garbage in the tail.
        memcpy(new_ptr, ptr, new_size);
    }
    return new_ptr;
}

#else

static inline void* malloc(size_t size) {
    (void)size;
    return NULL;
}

static inline void* calloc(size_t num, size_t size) {
    (void)num;
    (void)size;
    return NULL;
}

static inline void* realloc(void* ptr, size_t new_size) {
    (void)ptr;
    (void)new_size;
    return NULL;
}

#endif  // SO_HEAP_SIZE > 0

static inline void free(void* ptr) {
    (void)ptr;
}

#endif  // so_build_hosted

// so_heap_mark returns the position of the next allocation in the heap.
// so_heap_release returns the heap to the position mark, which loses every
// allocation made after so_heap_mark read the mark.
//
// A hosted host reclaims memory through free, so the mark is always 0 and the
// release does nothing. The freestanding heap is a bump allocator that never
// reclaims, so the release is the only way to reuse the memory.
//
// The mark and the release are not thread-safe. They are intended only
// for use by the testing package.

#if !defined(so_build_hosted) && SO_HEAP_SIZE > 0

static inline size_t so_heap_mark(void) {
    return so_heap_offset;
}

static inline void so_heap_release(size_t mark) {
    so_heap_offset = mark;
}

#else

static inline size_t so_heap_mark(void) {
    return 0;
}

static inline void so_heap_release(size_t mark) {
    (void)mark;
}

#endif
