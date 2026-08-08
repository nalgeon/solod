package maps

import (
	"unsafe"

	"solod.dev/so/c"
)

// Iter is an iterator over a Map's key-value pairs.
type Iter[K comparable, V any] struct {
	hdib []uint64
	keys []byte
	vals []byte
	i    int
}

// Next advances the iterator to the next key-value pair, which will
// then be available through the [Iter.Key] and [Iter.Value] methods.
// It returns false if there are no more pairs to iterate over.
//
//so:inline
func (it *Iter[K, V]) Next() bool {
	_found := false
	_hdib := unsafe.SliceData(it.hdib)
	_n := len(it.hdib)
	for it.i < _n {
		_hd := c.PtrAt(_hdib, it.i)
		c.Assume(_hd != nil)
		if *_hd&0xFFFF != 0 {
			it.i++
			_found = true
			break
		}
		it.i++
	}
	return _found
}

// Key returns the key of the current key-value pair.
//
//so:inline
func (it *Iter[K, V]) Key() K {
	c.Assert(it.i > 0, "maps: Iter.Key called before Next")
	_keys := c.PtrAs[K](unsafe.SliceData(it.keys))
	return *c.PtrAt(_keys, it.i-1)
}

// Value returns the value of the current key-value pair.
//
//so:inline
func (it *Iter[K, V]) Value() V {
	c.Assert(it.i > 0, "maps: Iter.Value called before Next")
	_vals := c.PtrAs[V](unsafe.SliceData(it.vals))
	return *c.PtrAt(_vals, it.i-1)
}
