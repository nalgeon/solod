#include "so/builtin/builtin.h"

// Atomic operations require either a hosted build or
// a freestanding target with lock-free instructions.
#if !defined(so_build_hosted) && __GCC_ATOMIC_BOOL_LOCK_FREE != 2
#error "so/sync/atomic: the target has no lock-free atomic of any width, and a freestanding build links no libatomic"
#endif

// This package maps Go-style atomics onto the compiler's __atomic builtins,
// which operate on ordinary (non-_Atomic) objects. Every operation uses
// sequentially consistent ordering (__ATOMIC_SEQ_CST), matching Go's
// sync/atomic. The T macro argument is the C type of the value.

// so_atomic_load atomically loads the value at p.
#define so_atomic_load(T, p) \
    (__atomic_load_n((p), __ATOMIC_SEQ_CST))

// so_atomic_store atomically stores v at p.
#define so_atomic_store(T, p, v) \
    (__atomic_store_n((p), (v), __ATOMIC_SEQ_CST))

// so_atomic_add atomically adds delta to *p and returns the new value.
#define so_atomic_add(T, p, delta) \
    (__atomic_add_fetch((p), (delta), __ATOMIC_SEQ_CST))

// so_atomic_swap atomically stores v at p and returns the previous value.
#define so_atomic_swap(T, p, v) \
    (__atomic_exchange_n((p), (v), __ATOMIC_SEQ_CST))

// so_atomic_cas atomically sets *p to new if it equals old,
// reporting whether the swap happened.
#define so_atomic_cas(T, p, old, new) ({                             \
    T _old = (old);                                                  \
    __atomic_compare_exchange_n((p), &_old, (new), false,            \
                                __ATOMIC_SEQ_CST, __ATOMIC_SEQ_CST); \
})

// atomic_Pointer is the backing store for atomic.Pointer[T]: a single machine
// pointer. T is erased to void* in C; the wrappers cast back at each access.
typedef struct {
    void* v;
} atomic_Pointer;

#define atomic_Pointer_Load(T, p) \
    ((T*)__atomic_load_n(&(p)->v, __ATOMIC_SEQ_CST))

#define atomic_Pointer_Store(T, p, val) \
    (__atomic_store_n(&(p)->v, (void*)(val), __ATOMIC_SEQ_CST))

#define atomic_Pointer_Swap(T, p, val) \
    ((T*)__atomic_exchange_n(&(p)->v, (void*)(val), __ATOMIC_SEQ_CST))

#define atomic_Pointer_CompareAndSwap(T, p, old, new) ({             \
    void* _old = (void*)(old);                                       \
    __atomic_compare_exchange_n(&(p)->v, &_old, (void*)(new), false, \
                                __ATOMIC_SEQ_CST, __ATOMIC_SEQ_CST); \
})
