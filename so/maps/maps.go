// Package maps provides a generic allocated map implementation.
package maps

import "solod.dev/so/mem"

//so:embed maps.h
var maps_h string

// Map is a generic hashmap similar to Go's built-in map[K]V.
//
//so:extern
type Map[K comparable, V any] struct{}

// New creates a new Map with the given minimal capacity
// using the provided allocator (or the default allocator if nil).
//
// Map will automatically grow and shrink as needed,
// but will not shrink below minCap.
//
// If the allocator is nil, uses the system allocator.
// The caller is responsible for freeing map resources
// with [Map.Free] when done using it.
//
//so:extern
func New[K comparable, V any](a mem.Allocator, size int) Map[K, V] {
	return Map[K, V]{}
}

// Has returns true if the given key is in the map.
//
//so:extern
func (m *Map[K, V]) Has(key K) bool {
	return false
}

// Get returns the value for the given key,
// or the zero value if the key is not in the map.
//
//so:extern
func (m *Map[K, V]) Get(key K) V {
	var zero V
	return zero
}

// Set sets the value for the given key,
// overwriting any existing value.
//
//so:extern
func (m *Map[K, V]) Set(key K, value V) {
}

// Delete removes the key and its value from the map.
// If the key is not in the map, does nothing.
//
//so:extern
func (m *Map[K, V]) Delete(key K) {
}

// Len returns the number of key-value pairs in the map.
//
//so:extern
func (m *Map[K, V]) Len() int {
	return 0
}

// Free frees internal resources used by the map.
// If the map is already freed, does nothing.
// The map must not be used after calling Free.
//
//so:extern
func (m *Map[K, V]) Free() {
}
