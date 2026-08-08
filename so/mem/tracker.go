package mem

import "solod.dev/so/sync/atomic"

// atomicStats is a thread-safe version of [Stats].
//
//so:promote
type atomicStats struct {
	alloc      atomic.Uint64 // bytes of allocated heap objects
	totalAlloc atomic.Uint64 // cumulative bytes allocated for heap objects
	mallocs    atomic.Uint64 // cumulative count of heap objects allocated
	frees      atomic.Uint64 // cumulative count of heap objects freed
}

// get returns a snapshot of the current statistics.
func (s *atomicStats) get() Stats {
	return Stats{
		Alloc:      s.alloc.Load(),
		TotalAlloc: s.totalAlloc.Load(),
		Mallocs:    s.mallocs.Load(),
		Frees:      s.frees.Load(),
	}
}

// A Tracker wraps an [Allocator] and tracks all
// allocations and deallocations made through it.
//
// Tracker is thread-safe as long as the underlying Allocator is thread-safe.
type Tracker struct {
	Allocator Allocator
	stats     atomicStats
}

func (t *Tracker) Alloc(size int, align int) (any, error) {
	ptr, err := t.Allocator.Alloc(size, align)
	if err != nil {
		return nil, err
	}
	t.stats.alloc.Add(uint64(size))
	t.stats.totalAlloc.Add(uint64(size))
	t.stats.mallocs.Add(1)
	return ptr, nil
}

func (t *Tracker) Realloc(ptr any, oldSize int, newSize int, align int) (any, error) {
	newPtr, err := t.Allocator.Realloc(ptr, oldSize, newSize, align)
	if err != nil {
		return nil, err
	}
	if newSize > oldSize {
		t.stats.alloc.Add(uint64(newSize - oldSize))
		t.stats.totalAlloc.Add(uint64(newSize - oldSize))
	} else {
		t.stats.alloc.Sub(uint64(oldSize - newSize))
	}
	t.stats.mallocs.Add(1)
	t.stats.frees.Add(1)
	return newPtr, nil
}

// Free frees a previously allocated block of memory.
// Freeing a nil pointer does not change the statistics.
func (t *Tracker) Free(ptr any, size int, align int) {
	if ptr == nil {
		return
	}
	t.Allocator.Free(ptr, size, align)
	t.stats.alloc.Sub(uint64(size))
	t.stats.frees.Add(1)
}

// Stats returns a snapshot of the current memory statistics.
// Each counter is read atomically, but the overall snapshot is only eventually
// consistent. When accessed concurrently, the counters might not match up with
// each other.
func (t *Tracker) Stats() Stats {
	return t.stats.get()
}
