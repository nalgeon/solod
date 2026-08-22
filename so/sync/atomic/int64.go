package atomic

// Int64 is an atomic int64. The zero value is zero.
// Int64 must not be copied after first use.
//
//so:attr aligned(8)
type Int64 struct {
	v int64
}

// Load atomically loads and returns the value stored in x.
func (x *Int64) Load() int64 {
	return load(&x.v)
}

// Store atomically stores val into x.
func (x *Int64) Store(val int64) {
	store(&x.v, val)
}

// Add atomically adds delta to x and returns the new value.
func (x *Int64) Add(delta int64) int64 {
	return add(&x.v, delta)
}

// Swap atomically stores new into x and returns the previous value.
func (x *Int64) Swap(new int64) int64 {
	return swap(&x.v, new)
}

// CompareAndSwap atomically sets x to new if it currently holds old,
// reporting whether the swap happened.
func (x *Int64) CompareAndSwap(old, new int64) bool {
	return cas(&x.v, old, new)
}
