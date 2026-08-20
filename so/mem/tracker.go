package mem

// A Tracker wraps an [Allocator] and tracks all
// allocations and deallocations made through it.
//
// Tracker is thread-safe as long as the underlying Allocator is thread-safe
// and the target has a lock-free 64-bit atomic. On a target without one,
// Tracker's stats could be inaccurate if it's shared across multiple threads.
type Tracker struct {
	Allocator Allocator
	stats     counters
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
// Each counter is read independently, and the overall snapshot is only
// eventually consistent. When accessed concurrently, the counters might
// not match up with each other.
func (t *Tracker) Stats() Stats {
	return t.stats.get()
}

// A counter is one statistic of a [Tracker].
//
// A target with a lock-free 64-bit atomic gets an atomic add. Every other
// target gets a plain add, which is not thread-safe.
//
//so:extern mem_counter
type counter struct {
	v uint64
}

// Load returns the current value of the counter.
//
//so:extern
func (c *counter) Load() uint64 {
	return c.v
}

// Add adds delta to the counter.
//
//so:extern
func (c *counter) Add(delta uint64) {
	c.v += delta
}

// Sub subtracts delta from the counter.
func (c *counter) Sub(delta uint64) {
	c.Add(^(delta - 1))
}

// counters holds the statistics of a [Tracker].
//
//so:promote
type counters struct {
	alloc      counter // bytes of allocated heap objects
	totalAlloc counter // cumulative bytes allocated for heap objects
	mallocs    counter // cumulative count of heap objects allocated
	frees      counter // cumulative count of heap objects freed
}

// get returns a snapshot of the current statistics.
func (s *counters) get() Stats {
	return Stats{
		Alloc:      s.alloc.Load(),
		TotalAlloc: s.totalAlloc.Load(),
		Mallocs:    s.mallocs.Load(),
		Frees:      s.frees.Load(),
	}
}
