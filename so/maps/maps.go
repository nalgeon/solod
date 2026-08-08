// Package maps provides a generic allocated map implementation.
package maps

import (
	"unsafe"

	"solod.dev/so/c"
	"solod.dev/so/mem"
)

//so:embed maps.h
var maps_h string

// Map is a generic hashmap similar to Go's built-in map[K]V.
// It automatically grows as needed, but does not shrink.
//
// The key type must be a number, a bool, a pointer, or a string.
// The map does not support array and struct keys.
//
// The map compares float keys byte by byte, not with the == operator.
// Negative zero and positive zero are different keys. A NaN key stays
// reachable after Set.
//
// The zero Map is not usable. Use [New] to create a Map.
//
// A Map must have a single owner. Two copies of a Map share the same
// buckets. If one copy grows, the other copy points to freed memory.
type Map[K comparable, V any] struct {
	rm rawMap
}

// New creates a new Map with the given minimal capacity.
//
//so:inline
func New[K comparable, V any](a mem.Allocator, size int) Map[K, V] {
	rm := newRawMap(a, size, c.Sizeof[K](), c.Sizeof[V]())
	return Map[K, V]{rm: rm}
}

// Has returns true if the given key is in the map.
//
//so:inline
func (m *Map[K, V]) Has(key K) bool {
	_key := key
	_found := false
	_m := m.rm
	if len(_m.hdib) > 0 {
		_hash := keyHash(&_key, _m.seed)
		_i := _hash & _m.mask
		_hdib := unsafe.SliceData(_m.hdib)
		_keys := c.PtrAs[K](unsafe.SliceData(_m.keys))
		_dist := 1
		for {
			_ehdib := *c.PtrAt(_hdib, _i)
			if int(_ehdib&0xFFFF) < _dist {
				break
			}
			if int(_ehdib>>16) == _hash &&
				keyEqual(&_key, c.PtrAt(_keys, _i)) {
				_found = true
				break
			}
			_i = (_i + 1) & _m.mask
			_dist++
		}
	}
	return _found
}

// Get returns the value for the given key,
// or the zero value if the key is not in the map.
//
//so:inline
func (m *Map[K, V]) Get(key K) V {
	_key := key
	_val := c.Zero[V]()
	_m := m.rm
	if len(_m.hdib) > 0 {
		_hash := keyHash(&_key, _m.seed)
		_i := _hash & _m.mask
		_hdib := unsafe.SliceData(_m.hdib)
		_keys := c.PtrAs[K](unsafe.SliceData(_m.keys))
		_vals := c.PtrAs[V](unsafe.SliceData(_m.vals))
		_dist := 1
		for {
			_ehdib := *c.PtrAt(_hdib, _i)
			if int(_ehdib&0xFFFF) < _dist {
				break
			}
			if int(_ehdib>>16) == _hash &&
				keyEqual(&_key, c.PtrAt(_keys, _i)) {
				_val = *c.PtrAt(_vals, _i)
				break
			}
			_i = (_i + 1) & _m.mask
			_dist++
		}
	}
	return _val
}

// Set sets the value for the given key,
// overwriting any existing value.
//
//so:inline
func (m *Map[K, V]) Set(key K, value V) {
	_key := key
	_val := value
	_m := &m.rm
	if _m.len >= _m.growAt {
		_m.Resize(len(_m.hdib) * 2)
	}
	_hash := keyHash(&_key, _m.seed)
	_ehdib := (uint64(_hash) << 16) | 1
	_i := _hash & _m.mask
	_hdib := unsafe.SliceData(_m.hdib)
	_keys := c.PtrAs[K](unsafe.SliceData(_m.keys))
	_vals := c.PtrAs[V](unsafe.SliceData(_m.vals))
	_ekey := _key
	_eval := _val
	for {
		_hdi := c.PtrAt(_hdib, _i)
		if *_hdi&0xFFFF == 0 {
			*_hdi = _ehdib
			*c.PtrAt(_keys, _i) = _ekey
			*c.PtrAt(_vals, _i) = _eval
			_m.len++
			break
		}
		if _ehdib>>16 == *_hdi>>16 &&
			keyEqual(&_ekey, c.PtrAt(_keys, _i)) {
			*c.PtrAt(_vals, _i) = _eval
			break
		}
		if *_hdi&0xFFFF < _ehdib&0xFFFF {
			mem.Swap(_hdi, &_ehdib)
			mem.Swap(c.PtrAt(_keys, _i), &_ekey)
			mem.Swap(c.PtrAt(_vals, _i), &_eval)
		}
		_i = (_i + 1) & _m.mask
		_ehdib++
	}
}

// Delete removes the key and its value from the map.
// If the key is not in the map, does nothing.
//
//so:inline
func (m *Map[K, V]) Delete(key K) {
	_key := key
	_m := &m.rm
	_hash := keyHash(&_key, _m.seed)
	_i := _hash & _m.mask
	_hdib := unsafe.SliceData(_m.hdib)
	_keys := c.PtrAs[K](unsafe.SliceData(_m.keys))
	_vals := c.PtrAs[V](unsafe.SliceData(_m.vals))
	_dist := 1
	for len(_m.hdib) > 0 {
		_hdi := c.PtrAt(_hdib, _i)
		if int(*_hdi&0xFFFF) < _dist {
			break
		}
		if int(*_hdi>>16) == _hash && keyEqual(&_key, c.PtrAt(_keys, _i)) {
			for {
				_prev := _i
				_i = (_i + 1) & _m.mask
				if *c.PtrAt(_hdib, _i)&0xFFFF <= 1 {
					*c.PtrAt(_hdib, _prev) = 0
					mem.Clear(c.PtrAt(_keys, _prev), c.Sizeof[K]())
					mem.Clear(c.PtrAt(_vals, _prev), c.Sizeof[V]())
					break
				}
				*c.PtrAt(_hdib, _prev) = *c.PtrAt(_hdib, _i) - 1
				*c.PtrAt(_keys, _prev) = *c.PtrAt(_keys, _i)
				*c.PtrAt(_vals, _prev) = *c.PtrAt(_vals, _i)
			}
			_m.len--
			break
		}
		_i = (_i + 1) & _m.mask
		_dist++
	}
}

// Iter returns an iterator over the map's key-value pairs.
// The map must not be modified or freed while iterating.
//
//so:inline
func (m *Map[K, V]) Iter() Iter[K, V] {
	return Iter[K, V]{hdib: m.rm.hdib, keys: m.rm.keys, vals: m.rm.vals}
}

// Len returns the number of key-value pairs in the map.
//
//so:inline
func (m *Map[K, V]) Len() int {
	return m.rm.Len()
}

// Clear removes all key-value pairs from the map, resetting
// it to an empty state. Does not free map resources;
// the map can be reused after Clear.
//
//so:inline
func (m *Map[K, V]) Clear() {
	m.rm.Clear()
}

// Free frees internal resources used by the map.
// If the map is already freed, does nothing.
// The map must not be used after calling Free.
//
//so:inline
func (m *Map[K, V]) Free() {
	m.rm.Free()
}

//so:extern maps_keyHash
func keyHash[K comparable](key *K, seed uint64) int {
	switch p := any(key).(type) {
	case *string:
		return maps_hashString(p, seed)
	default:
		return maps_hash((*byte)(unsafe.Pointer(key)), c.Sizeof[K](), seed)
	}
}

//so:extern maps_keyEqual
func keyEqual[K comparable](a, b *K) bool {
	return *a == *b
}
