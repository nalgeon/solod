#include "so/builtin/builtin.h"

// Forward declarations for ByteMap-related types and functions.
typedef struct maps_ByteMap maps_ByteMap;
so_int maps_HashBytes(so_Slice key);
bool maps_EqualBytes(so_Slice a, so_Slice b);

// hashString hashes a string key by its content rather than pointer.
static inline so_int maps_hashString(so_Slice key) {
    so_String* s = (so_String*)key.ptr;
    so_Slice content = {(void*)s->ptr, s->len, s->len};
    return maps_HashBytes(content);
}

// equalString compares two string keys by their content.
static inline bool maps_equalString(so_Slice a, so_Slice b) {
    return so_string_eq(*(so_String*)a.ptr, *(so_String*)b.ptr);
}

// Map is a generic hashmap similar to Go's built-in map[K]V.
typedef maps_ByteMap maps_Map;

// New creates a new Map with the given minimal capacity
// using the provided allocator (or the default allocator if nil).
#define maps_New(K, V, a, size) ({                                    \
    maps_ByteMap _m = maps_NewByteMap((a), (size), (so_int)sizeof(K), \
                                      (so_int)sizeof(V));             \
    _m.hashFn = _Generic((K){0},                                      \
        so_String: maps_hashString,                                   \
        default: maps_HashBytes);                                     \
    _m.equalFn = _Generic((K){0},                                     \
        so_String: maps_equalString,                                  \
        default: maps_EqualBytes);                                    \
    _m;                                                               \
})

// Has returns true if the given key is in the map.
#define maps_Map_Has(K, V, m, key) ({           \
    K _k = (key);                               \
    so_Slice _ks = {&_k, sizeof(K), sizeof(K)}; \
    so_Slice _vs = {0};                         \
    maps_ByteMap_Get((m), _ks, _vs);            \
})

// Get returns the value for the given key,
// or the zero value if the key is not in the map.
#define maps_Map_Get(K, V, m, key) ({           \
    K _k = (key);                               \
    V _v;                                       \
    memset(&_v, 0, sizeof(V));                  \
    so_Slice _ks = {&_k, sizeof(K), sizeof(K)}; \
    so_Slice _vs = {&_v, sizeof(V), sizeof(V)}; \
    maps_ByteMap_Get((m), _ks, _vs);            \
    _v;                                         \
})

// Set sets the value for the given key,
// overwriting any existing value.
#define maps_Map_Set(K, V, m, key, value) ({    \
    K _k = (key);                               \
    V _v = (value);                             \
    so_Slice _ks = {&_k, sizeof(K), sizeof(K)}; \
    so_Slice _vs = {&_v, sizeof(V), sizeof(V)}; \
    maps_ByteMap_Set((m), _ks, _vs);            \
})

// Delete removes the key and its value from the map.
// If the key is not in the map, does nothing.
#define maps_Map_Delete(K, V, m, key) ({        \
    K _k = (key);                               \
    so_Slice _ks = {&_k, sizeof(K), sizeof(K)}; \
    maps_ByteMap_Delete((m), _ks);              \
})

// Len returns the number of key-value pairs in the map.
#define maps_Map_Len(K, V, m) \
    maps_ByteMap_Len(m)

// Free frees internal resources used by the map.
// If the map is already freed, does nothing.
#define maps_Map_Free(K, V, m) \
    maps_ByteMap_Free(m)
