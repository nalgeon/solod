//go:build ignore
#include "mem.h"

#if !defined(so_build_hosted) && SO_HEAP_SIZE > 0

// The heap of a freestanding host, and the position of the next allocation in it.
// malloc rounds the offset to 16 bytes, so the heap must start
// at a 16 byte boundary to give back aligned pointers.
alignas(16) char so_heap[SO_HEAP_SIZE];
size_t so_heap_offset = 0;

#endif
