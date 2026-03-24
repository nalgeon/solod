#include "main.h"

// -- Implementation --

int main(void) {
    {
        // SetGet: insert 3 entries, verify all values
        maps_Map m = maps_New(so_String, so_int, (mem_Allocator){0}, 0);
        maps_Map_Set(so_String, so_int, &m, so_str("abc"), 11);
        maps_Map_Set(so_String, so_int, &m, so_str("def"), 22);
        maps_Map_Set(so_String, so_int, &m, so_str("xyz"), 33);
        if (maps_Map_Get(so_String, so_int, &m, so_str("abc")) != 11) {
            so_panic("want abc = 11");
        }
        so_String key = so_str("abc");
        if (maps_Map_Get(so_String, so_int, &m, key) != 11) {
            so_panic("want abc = 11 for key = abc");
        }
        if (maps_Map_Get(so_String, so_int, &m, so_str("def")) != 22) {
            so_panic("want def = 22");
        }
        if (maps_Map_Get(so_String, so_int, &m, so_str("xyz")) != 33) {
            so_panic("want xyz = 33");
        }
        if (maps_Map_Get(so_String, so_int, &m, so_str("missing")) != 0) {
            so_panic("want missing = 0");
        }
        if (maps_Map_Len(so_String, so_int, &m) != 3) {
            so_panic("want len = 3");
        }
        maps_Map_Free(so_String, so_int, &m);
    }
    {
        // String values.
        maps_Map m = maps_New(int32_t, so_String, (mem_Allocator){0}, 0);
        maps_Map_Set(int32_t, so_String, &m, 11, so_str("abc"));
        maps_Map_Set(int32_t, so_String, &m, 22, so_str("def"));
        maps_Map_Set(int32_t, so_String, &m, 33, so_str("xyz"));
        if (so_string_ne(maps_Map_Get(int32_t, so_String, &m, 11), so_str("abc"))) {
            so_panic("want 11 = abc");
        }
        if (so_string_ne(maps_Map_Get(int32_t, so_String, &m, 22), so_str("def"))) {
            so_panic("want 22 = def");
        }
        if (so_string_ne(maps_Map_Get(int32_t, so_String, &m, 33), so_str("xyz"))) {
            so_panic("want 33 = xyz");
        }
        if (so_string_ne(maps_Map_Get(int32_t, so_String, &m, 44), so_str(""))) {
            so_panic("want 44 = empty string");
        }
        maps_Map_Free(int32_t, so_String, &m);
    }
    {
        // Has.
        maps_Map m = maps_New(so_String, so_int, (mem_Allocator){0}, 0);
        maps_Map_Set(so_String, so_int, &m, so_str("abc"), 11);
        maps_Map_Set(so_String, so_int, &m, so_str("def"), 22);
        if (!maps_Map_Has(so_String, so_int, &m, so_str("abc"))) {
            so_panic("want has(abc)");
        }
        if (!maps_Map_Has(so_String, so_int, &m, so_str("def"))) {
            so_panic("want has(def)");
        }
        if (maps_Map_Has(so_String, so_int, &m, so_str("missing"))) {
            so_panic("want has(missing) == false");
        }
        maps_Map_Free(so_String, so_int, &m);
    }
    {
        // Delete: insert 3 entries, delete one, verify
        maps_Map m = maps_New(so_String, so_int, (mem_Allocator){0}, 0);
        maps_Map_Set(so_String, so_int, &m, so_str("abc"), 11);
        maps_Map_Set(so_String, so_int, &m, so_str("def"), 22);
        maps_Map_Set(so_String, so_int, &m, so_str("xyz"), 33);
        maps_Map_Delete(so_String, so_int, &m, so_str("def"));
        // no-op
        maps_Map_Delete(so_String, so_int, &m, so_str("missing"));
        if (maps_Map_Get(so_String, so_int, &m, so_str("def")) != 0) {
            so_panic("want def = 0 after delete");
        }
        if (maps_Map_Get(so_String, so_int, &m, so_str("abc")) != 11) {
            so_panic("want abc = 11 after delete");
        }
        if (maps_Map_Get(so_String, so_int, &m, so_str("xyz")) != 33) {
            so_panic("want xyz = 33 after delete");
        }
        if (maps_Map_Len(so_String, so_int, &m) != 2) {
            so_panic("want len = 2 after delete");
        }
        maps_Map_Free(so_String, so_int, &m);
    }
    {
        // Overwrite: set same key twice, verify latest value wins
        maps_Map m = maps_New(so_String, so_int, (mem_Allocator){0}, 0);
        maps_Map_Set(so_String, so_int, &m, so_str("key"), 100);
        maps_Map_Set(so_String, so_int, &m, so_str("key"), 200);
        if (maps_Map_Get(so_String, so_int, &m, so_str("key")) != 200) {
            so_panic("want key = 200 after overwrite");
        }
        if (maps_Map_Len(so_String, so_int, &m) != 1) {
            so_panic("want len = 1 after overwrite");
        }
        maps_Map_Free(so_String, so_int, &m);
    }
    {
        // Missing: get non-existent key returns zero value
        maps_Map m = maps_New(so_String, so_int, (mem_Allocator){0}, 0);
        if (maps_Map_Get(so_String, so_int, &m, so_str("missing")) != 0) {
            so_panic("want missing = 0");
        }
        maps_Map_Free(so_String, so_int, &m);
    }
    {
        // Grow: insert 100 int-keyed entries, verify all retrievable
        maps_Map m = maps_New(so_int, so_int, (mem_Allocator){0}, 0);
        for (so_int i = 0; i < 100; i++) {
            maps_Map_Set(so_int, so_int, &m, i, i * 10);
        }
        for (so_int i = 0; i < 100; i++) {
            if (maps_Map_Get(so_int, so_int, &m, i) != i * 10) {
                so_panic("wrong value after grow");
            }
        }
        if (maps_Map_Len(so_int, so_int, &m) != 100) {
            so_panic("want len = 100 after grow");
        }
        maps_Map_Free(so_int, so_int, &m);
    }
}
