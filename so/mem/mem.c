//go:build ignore
#include "mem.h"

#if !defined(so_build_hosted) && SO_HEAP_SIZE > 0

// The heap of a freestanding host, and the
// position of the next allocation in it.
char so_heap[SO_HEAP_SIZE];
size_t so_heap_offset = 0;

#endif
