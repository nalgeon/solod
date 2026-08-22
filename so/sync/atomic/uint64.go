package atomic

// Uint64 is an atomic uint64. The zero value is zero.
// Uint64 must not be copied after first use.
//
//so:attr aligned(8)
type Uint64 struct {
	v uint64
}

// Load atomically loads and returns the value stored in x.
func (x *Uint64) Load() uint64 {
	return load(&x.v)
}

// Store atomically stores val into x.
func (x *Uint64) Store(val uint64) {
	store(&x.v, val)
}

// Add atomically adds delta to x and returns the new value.
func (x *Uint64) Add(delta uint64) uint64 {
	return add(&x.v, delta)
}

// Sub atomically subtracts delta from x and returns the new value.
func (x *Uint64) Sub(delta uint64) uint64 {
	return add(&x.v, ^(delta - 1))
}

// Swap atomically stores new into x and returns the previous value.
func (x *Uint64) Swap(new uint64) uint64 {
	return swap(&x.v, new)
}

// CompareAndSwap atomically sets x to new if it currently holds old,
// reporting whether the swap happened.
func (x *Uint64) CompareAndSwap(old, new uint64) bool {
	return cas(&x.v, old, new)
}
