#include "main.h"

// -- Types --

typedef struct empty empty;

typedef struct empty {
} empty;

// -- Forward declarations --
static so_int takeMap(so_Map* m);
static void modifyMap(so_Map* m);
static so_int main_MapHolder_get(main_MapHolder h, so_String key);
static so_int main_MapHolder_sum(main_MapHolder h);
static so_int ten(void);
static so_int twenty(void);

// -- Implementation --

static so_int takeMap(so_Map* m) {
    return so_map_get(so_String, so_int, m, so_str("a")) + so_map_get(so_String, so_int, m, so_str("b"));
}

static void modifyMap(so_Map* m) {
    so_map_set(so_String, so_int, m, so_str("a"), 99);
    so_map_set(so_String, so_int, m, so_str("b"), 22);
}

static so_int main_MapHolder_get(main_MapHolder h, so_String key) {
    return so_map_get(so_String, so_int, h.m, key);
}

static so_int main_MapHolder_sum(main_MapHolder h) {
    so_int s = 0;
    for (so_int _i = 0; _i < h.m->cap; _i++) {
        if (!h.m->used[_i]) continue;
        so_int v = ((so_int*)h.m->vals)[_i];
        s += v;
    }
    return s;
}

static so_int ten(void) {
    return 10;
}

static so_int twenty(void) {
    return 20;
}

int main(void) {
    {
        // Key type: string (eq_str)
        so_Map* m = so_map_lit(so_String, so_int, 2, ((so_String[]){so_str("a"), so_str("b")}), ((so_int[]){1, 2}));
        if (so_map_get(so_String, so_int, m, so_str("a")) != 1 || so_map_get(so_String, so_int, m, so_str("b")) != 2) {
            so_panic("string key");
        }
    }
    {
        // Key type: int (eq_8)
        so_Map* m = so_map_lit(so_int, so_int, 2, ((so_int[]){1, 2}), ((so_int[]){10, 20}));
        if (so_map_get(so_int, so_int, m, 1) != 10 || so_map_get(so_int, so_int, m, 2) != 20) {
            so_panic("int key");
        }
    }
    {
        // Key type: float32 (eq_4)
        so_Map* m = so_map_lit(float, so_int, 2, ((float[]){1.5f, 2.5f}), ((so_int[]){10, 20}));
        if (so_map_get(float, so_int, m, 1.5f) != 10 || so_map_get(float, so_int, m, 2.5f) != 20) {
            so_panic("float32 key");
        }
    }
    {
        // Key type: uint16 (eq_2)
        so_Map* m = so_map_lit(uint16_t, so_int, 2, ((uint16_t[]){1, 2}), ((so_int[]){10, 20}));
        if (so_map_get(uint16_t, so_int, m, 1) != 10 || so_map_get(uint16_t, so_int, m, 2) != 20) {
            so_panic("uint16 key");
        }
    }
    {
        // Key type: bool (eq_1)
        so_Map* m = so_map_lit(bool, so_int, 2, ((bool[]){true, false}), ((so_int[]){1, 0}));
        if (so_map_get(bool, so_int, m, true) != 1 || so_map_get(bool, so_int, m, false) != 0) {
            so_panic("bool key");
        }
    }
    {
        // Key type: uint8 (eq_1)
        so_Map* m = so_map_lit(uint8_t, so_int, 2, ((uint8_t[]){1, 2}), ((so_int[]){10, 20}));
        if (so_map_get(uint8_t, so_int, m, 1) != 10 || so_map_get(uint8_t, so_int, m, 2) != 20) {
            so_panic("uint8 key");
        }
    }
    {
        // Key type: *int (pointer, eq_8)
        so_int a = 1;
        so_int b = 2;
        so_Map* m = so_map_lit(so_int*, so_String, 2, ((so_int*[]){&a, &b}), ((so_String[]){so_str("first"), so_str("second")}));
        if (so_string_ne(so_map_get(so_int*, so_String, m, &a), so_str("first")) || so_string_ne(so_map_get(so_int*, so_String, m, &b), so_str("second"))) {
            so_panic("pointer key");
        }
    }
    {
        // Value type: string
        so_Map* m = so_map_lit(so_int, so_String, 2, ((so_int[]){1, 2}), ((so_String[]){so_str("a"), so_str("b")}));
        if (so_string_ne(so_map_get(so_int, so_String, m, 1), so_str("a")) || so_string_ne(so_map_get(so_int, so_String, m, 2), so_str("b"))) {
            so_panic("string value");
        }
    }
    {
        // Value type: bool
        so_Map* m = so_map_lit(so_String, bool, 2, ((so_String[]){so_str("yes"), so_str("no")}), ((bool[]){true, false}));
        if (!so_map_get(so_String, bool, m, so_str("yes")) || so_map_get(so_String, bool, m, so_str("no"))) {
            so_panic("bool value");
        }
    }
    {
        // Value type: float64
        so_Map* m = so_map_lit(so_String, double, 2, ((so_String[]){so_str("pi"), so_str("e")}), ((double[]){3.14, 2.71}));
        if (so_map_get(so_String, double, m, so_str("pi")) != 3.14 || so_map_get(so_String, double, m, so_str("e")) != 2.71) {
            so_panic("float64 value");
        }
    }
    {
        // Key and value: string
        so_Map* m = so_map_lit(so_String, so_String, 2, ((so_String[]){so_str("a"), so_str("b")}), ((so_String[]){so_str("x"), so_str("y")}));
        if (so_string_ne(so_map_get(so_String, so_String, m, so_str("a")), so_str("x")) || so_string_ne(so_map_get(so_String, so_String, m, so_str("b")), so_str("y"))) {
            so_panic("string string");
        }
    }
    {
        // Value type: struct
        so_Map* m = so_map_lit(so_String, main_Pair, 2, ((so_String[]){so_str("origin"), so_str("point")}), ((main_Pair[]){(main_Pair){0, 0}, (main_Pair){3, 4}}));
        if (so_map_get(so_String, main_Pair, m, so_str("origin")).x != 0 || so_map_get(so_String, main_Pair, m, so_str("point")).x != 3 || so_map_get(so_String, main_Pair, m, so_str("point")).y != 4) {
            so_panic("struct value");
        }
    }
    {
        // Value type: slice
        so_Slice s1 = so_make_slice(so_int, 2, 2);
        so_at(so_int, s1, 0) = 10;
        so_at(so_int, s1, 1) = 20;
        so_Slice s2 = so_make_slice(so_int, 2, 2);
        so_at(so_int, s2, 0) = 30;
        so_at(so_int, s2, 1) = 40;
        so_Map* m = so_map_lit(so_String, so_Slice, 2, ((so_String[]){so_str("a"), so_str("b")}), ((so_Slice[]){s1, s2}));
        if (so_at(so_int, so_map_get(so_String, so_Slice, m, so_str("a")), 0) != 10 || so_at(so_int, so_map_get(so_String, so_Slice, m, so_str("b")), 1) != 40) {
            so_panic("slice value");
        }
    }
    {
        // Value type: *int (pointer)
        so_int a = 42;
        so_int b = 99;
        so_Map* m = so_map_lit(so_String, so_int*, 2, ((so_String[]){so_str("a"), so_str("b")}), ((so_int*[]){&a, &b}));
        if (*so_map_get(so_String, so_int*, m, so_str("a")) != 42 || *so_map_get(so_String, so_int*, m, so_str("b")) != 99) {
            so_panic("pointer value");
        }
    }
    {
        // Value type: map (nested)
        so_Map* inner1 = so_map_lit(so_String, so_int, 1, ((so_String[]){so_str("x")}), ((so_int[]){1}));
        so_Map* inner2 = so_map_lit(so_String, so_int, 1, ((so_String[]){so_str("y")}), ((so_int[]){2}));
        so_Map* m = so_map_lit(so_String, so_Map*, 2, ((so_String[]){so_str("a"), so_str("b")}), ((so_Map*[]){inner1, inner2}));
        if (so_map_get(so_String, so_int, so_map_get(so_String, so_Map*, m, so_str("a")), so_str("x")) != 1 || so_map_get(so_String, so_int, so_map_get(so_String, so_Map*, m, so_str("b")), so_str("y")) != 2) {
            so_panic("nested map value");
        }
    }
    {
        // Value type: function
        so_Map* m = so_map_lit(so_String, main_IntFunc, 2, ((so_String[]){so_str("ten"), so_str("twenty")}), ((main_IntFunc[]){ten, twenty}));
        if (so_map_get(so_String, main_IntFunc, m, so_str("ten"))() != 10 || so_map_get(so_String, main_IntFunc, m, so_str("twenty"))() != 20) {
            so_panic("func value");
        }
    }
    {
        // Single element literal
        so_Map* m = so_map_lit(so_String, so_int, 1, ((so_String[]){so_str("only")}), ((so_int[]){42}));
        if (so_map_get(so_String, so_int, m, so_str("only")) != 42) {
            so_panic("single element");
        }
    }
    {
        // Empty literal
        so_Map* m = &(so_Map){};
        if (m->len != 0) {
            so_panic("empty literal");
        }
    }
    {
        // Make and populate
        so_Map* m = so_make_map(so_String, so_int, 3);
        if (m->len != 0) {
            so_panic("make initial len");
        }
        so_map_set(so_String, so_int, m, so_str("a"), 10);
        so_map_set(so_String, so_int, m, so_str("b"), 20);
        so_map_set(so_String, so_int, m, so_str("c"), 30);
        if (so_map_get(so_String, so_int, m, so_str("a")) != 10 || so_map_get(so_String, so_int, m, so_str("b")) != 20 || so_map_get(so_String, so_int, m, so_str("c")) != 30) {
            so_panic("make values");
        }
        if (m->len != 3) {
            so_panic("make final len");
        }
    }
    {
        // Make with int key
        so_Map* m = so_make_map(so_int, so_String, 2);
        so_map_set(so_int, so_String, m, 1, so_str("one"));
        so_map_set(so_int, so_String, m, 2, so_str("two"));
        if (so_string_ne(so_map_get(so_int, so_String, m, 1), so_str("one")) || so_string_ne(so_map_get(so_int, so_String, m, 2), so_str("two"))) {
            so_panic("make int key");
        }
    }
    {
        // Missing key: zero value int
        so_Map* m = so_map_lit(so_String, so_int, 1, ((so_String[]){so_str("a")}), ((so_int[]){1}));
        if (so_map_get(so_String, so_int, m, so_str("missing")) != 0) {
            so_panic("zero int");
        }
    }
    {
        // Missing key: zero value string
        so_Map* m = so_map_lit(so_int, so_String, 1, ((so_int[]){1}), ((so_String[]){so_str("a")}));
        if (so_string_ne(so_map_get(so_int, so_String, m, 99), so_str(""))) {
            so_panic("zero string");
        }
    }
    {
        // Missing key: zero value bool
        so_Map* m = so_map_lit(so_String, bool, 1, ((so_String[]){so_str("a")}), ((bool[]){true}));
        if (so_map_get(so_String, bool, m, so_str("missing"))) {
            so_panic("zero bool");
        }
    }
    {
        // Missing key: zero value struct
        so_Map* m = so_map_lit(so_String, main_Pair, 1, ((so_String[]){so_str("a")}), ((main_Pair[]){(main_Pair){1, 2}}));
        main_Pair p = so_map_get(so_String, main_Pair, m, so_str("missing"));
        if (p.x != 0 || p.y != 0) {
            so_panic("zero struct");
        }
    }
    {
        // Overwrite existing key
        so_Map* m = so_map_lit(so_String, so_int, 1, ((so_String[]){so_str("a")}), ((so_int[]){1}));
        so_map_set(so_String, so_int, m, so_str("a"), 99);
        if (so_map_get(so_String, so_int, m, so_str("a")) != 99) {
            so_panic("overwrite value");
        }
        if (m->len != 1) {
            so_panic("overwrite len");
        }
    }
    {
        // Map value in arithmetic
        so_Map* m = so_map_lit(so_String, so_int, 2, ((so_String[]){so_str("a"), so_str("b")}), ((so_int[]){10, 20}));
        so_int sum = so_map_get(so_String, so_int, m, so_str("a")) + so_map_get(so_String, so_int, m, so_str("b"));
        if (sum != 30) {
            so_panic("arithmetic");
        }
    }
    {
        // Map value in nested expression
        so_Map* m = so_map_lit(so_String, so_int, 2, ((so_String[]){so_str("a"), so_str("b")}), ((so_int[]){2, 3}));
        so_int result = so_map_get(so_String, so_int, m, so_str("a")) * so_map_get(so_String, so_int, m, so_str("b")) + so_map_get(so_String, so_int, m, so_str("a"));
        if (result != 8) {
            so_panic("nested expr");
        }
    }
    {
        // Map bool value in condition
        so_Map* m = so_map_lit(so_String, bool, 1, ((so_String[]){so_str("flag")}), ((bool[]){true}));
        if (!so_map_get(so_String, bool, m, so_str("flag"))) {
            so_panic("bool condition");
        }
    }
    {
        // Comma-ok: define then assign
        so_Map* m = so_map_lit(so_String, so_int, 2, ((so_String[]){so_str("a"), so_str("b")}), ((so_int[]){1, 2}));
        // Define, key exists.
        so_int v = so_map_get(so_String, so_int, m, so_str("a"));
        bool ok = so_map_has(so_String, m, so_str("a"));
        if (!ok || v != 1) {
            so_panic("comma-ok define hit");
        }
        // Assign, key missing.
        v = so_map_get(so_String, so_int, m, so_str("missing"));
        ok = so_map_has(so_String, m, so_str("missing"));
        if (ok || v != 0) {
            so_panic("comma-ok assign miss");
        }
        // Assign, key exists.
        v = so_map_get(so_String, so_int, m, so_str("b"));
        ok = so_map_has(so_String, m, so_str("b"));
        if (!ok || v != 2) {
            so_panic("comma-ok assign hit");
        }
    }
    {
        // Comma-ok: blank value
        so_Map* m = so_map_lit(so_String, so_int, 1, ((so_String[]){so_str("a")}), ((so_int[]){1}));
        bool ok = so_map_has(so_String, m, so_str("a"));
        if (!ok) {
            so_panic("comma-ok blank value hit");
        }
        ok = so_map_has(so_String, m, so_str("missing"));
        if (ok) {
            so_panic("comma-ok blank value miss");
        }
    }
    {
        // Comma-ok: blank ok
        so_Map* m = so_map_lit(so_String, so_int, 1, ((so_String[]){so_str("a")}), ((so_int[]){1}));
        so_int v = so_map_get(so_String, so_int, m, so_str("a"));
        if (v != 1) {
            so_panic("comma-ok blank ok");
        }
    }
    {
        // Comma-ok: with string value
        so_Map* m = so_map_lit(so_int, so_String, 1, ((so_int[]){1}), ((so_String[]){so_str("hello")}));
        so_String v = so_map_get(so_int, so_String, m, 1);
        bool ok = so_map_has(so_int, m, 1);
        if (!ok || so_string_ne(v, so_str("hello"))) {
            so_panic("comma-ok string value");
        }
        v = so_map_get(so_int, so_String, m, 99);
        ok = so_map_has(so_int, m, 99);
        if (ok || so_string_ne(v, so_str(""))) {
            so_panic("comma-ok string value miss");
        }
    }
    {
        // Range: key + value
        so_Map* m = so_map_lit(so_int, so_int, 3, ((so_int[]){1, 2, 3}), ((so_int[]){10, 20, 30}));
        so_int sum = 0;
        for (so_int _i = 0; _i < m->cap; _i++) {
            if (!m->used[_i]) continue;
            so_int k = ((so_int*)m->keys)[_i];
            so_int v = ((so_int*)m->vals)[_i];
            sum += k * v;
        }
        // 1*10 + 2*20 + 3*30 = 10 + 40 + 90 = 140
        if (sum != 140) {
            so_panic("range k,v define");
        }
    }
    {
        // Range: key only
        so_Map* m = so_map_lit(so_int, so_int, 2, ((so_int[]){10, 20}), ((so_int[]){100, 200}));
        so_int sum = 0;
        for (so_int _i = 0; _i < m->cap; _i++) {
            if (!m->used[_i]) continue;
            so_int k = ((so_int*)m->keys)[_i];
            sum += k;
        }
        if (sum != 30) {
            so_panic("range k only");
        }
    }
    {
        // Range: value only
        so_Map* m = so_map_lit(so_int, so_int, 2, ((so_int[]){10, 20}), ((so_int[]){100, 200}));
        so_int sum = 0;
        for (so_int _i = 0; _i < m->cap; _i++) {
            if (!m->used[_i]) continue;
            so_int v = ((so_int*)m->vals)[_i];
            sum += v;
        }
        if (sum != 300) {
            so_panic("range v only");
        }
    }
    {
        // Range: key + value (assign, not define)
        so_Map* m = so_map_lit(so_int, so_int, 2, ((so_int[]){1, 2}), ((so_int[]){10, 20}));
        so_int k = 0;
        so_int v = 0;
        so_int sum = 0;
        for (so_int _i = 0; _i < m->cap; _i++) {
            if (!m->used[_i]) continue;
            k = ((so_int*)m->keys)[_i];
            v = ((so_int*)m->vals)[_i];
            sum += k + v;
        }
        // 1+10 + 2+20 = 33
        if (sum != 33) {
            so_panic("range assign");
        }
    }
    {
        // Range: string keys and values
        so_Map* m = so_map_lit(so_String, so_String, 2, ((so_String[]){so_str("hello"), so_str("foo")}), ((so_String[]){so_str("world"), so_str("bar")}));
        so_String keys = so_str("");
        so_String vals = so_str("");
        for (so_int _i = 0; _i < m->cap; _i++) {
            if (!m->used[_i]) continue;
            so_String k = ((so_String*)m->keys)[_i];
            so_String v = ((so_String*)m->vals)[_i];
            keys = so_string_add(keys, k);
            vals = so_string_add(vals, v);
        }
        if (so_string_ne(keys, so_str("hellofoo")) && so_string_ne(keys, so_str("foohello"))) {
            so_panic("range string keys");
        }
        if (so_string_ne(vals, so_str("worldbar")) && so_string_ne(vals, so_str("barworld"))) {
            so_panic("range string values");
        }
    }
    {
        // Range: over struct values
        so_Map* m = so_map_lit(so_String, main_Pair, 2, ((so_String[]){so_str("a"), so_str("b")}), ((main_Pair[]){(main_Pair){1, 2}, (main_Pair){3, 4}}));
        so_int sum = 0;
        for (so_int _i = 0; _i < m->cap; _i++) {
            if (!m->used[_i]) continue;
            main_Pair v = ((main_Pair*)m->vals)[_i];
            sum += v.x + v.y;
        }
        if (sum != 10) {
            so_panic("range struct value");
        }
    }
    {
        // len: literal
        so_Map* m = so_map_lit(so_String, so_int, 3, ((so_String[]){so_str("a"), so_str("b"), so_str("c")}), ((so_int[]){1, 2, 3}));
        if (m->len != 3) {
            so_panic("len literal");
        }
    }
    {
        // len: empty
        so_Map* m = &(so_Map){};
        if (m->len != 0) {
            so_panic("len empty");
        }
    }
    {
        // len: grows with set
        so_Map* m = so_make_map(so_String, so_int, 3);
        if (m->len != 0) {
            so_panic("len make 0");
        }
        so_map_set(so_String, so_int, m, so_str("a"), 1);
        if (m->len != 1) {
            so_panic("len make 1");
        }
        so_map_set(so_String, so_int, m, so_str("b"), 2);
        if (m->len != 2) {
            so_panic("len make 2");
        }
    }
    {
        // len: overwrite does not change len
        so_Map* m = so_map_lit(so_String, so_int, 1, ((so_String[]){so_str("a")}), ((so_int[]){1}));
        so_map_set(so_String, so_int, m, so_str("a"), 99);
        if (m->len != 1) {
            so_panic("len overwrite");
        }
    }
    {
        // len: in expression
        so_Map* m = so_map_lit(so_String, so_int, 2, ((so_String[]){so_str("a"), so_str("b")}), ((so_int[]){1, 2}));
        so_int n = m->len + 1;
        if (n != 3) {
            so_panic("len expr");
        }
    }
    {
        // Nil: non-nil literal
        so_Map* m = so_map_lit(so_String, so_int, 1, ((so_String[]){so_str("a")}), ((so_int[]){1}));
        if (m == NULL) {
            so_panic("non-nil");
        }
    }
    {
        // Nil: assign and check. Nil works on either side.
        so_Map* m = so_map_lit(so_String, so_int, 1, ((so_String[]){so_str("a")}), ((so_int[]){1}));
        m = NULL;
        if (m != NULL) {
            so_panic("nil after assign");
        }
        if (m != NULL) {
            so_panic("nil after assign");
        }
    }
    {
        // Pass map to function
        so_Map* m = so_map_lit(so_String, so_int, 2, ((so_String[]){so_str("a"), so_str("b")}), ((so_int[]){10, 20}));
        if (takeMap(m) != 30) {
            so_panic("pass to func");
        }
    }
    {
        // Function modifies map
        so_Map* m = so_make_map(so_String, so_int, 2);
        so_map_set(so_String, so_int, m, so_str("a"), 11);
        modifyMap(m);
        if (so_map_get(so_String, so_int, m, so_str("a")) != 99) {
            so_panic("func modify a");
        }
        if (so_map_get(so_String, so_int, m, so_str("b")) != 22) {
            so_panic("func modify b");
        }
        if (m->len != 2) {
            so_panic("func modify len");
        }
    }
    {
        // Method on struct with map field
        main_MapHolder h = (main_MapHolder){.m = so_map_lit(so_String, so_int, 2, ((so_String[]){so_str("x"), so_str("y")}), ((so_int[]){10, 20}))};
        if (main_MapHolder_get(h, so_str("x")) != 10) {
            so_panic("method get");
        }
        if (main_MapHolder_sum(h) != 30) {
            so_panic("method sum");
        }
    }
    {
        // Named map type
        main_StrMap m = so_map_lit(so_String, so_int, 2, ((so_String[]){so_str("a"), so_str("b")}), ((so_int[]){1, 2}));
        if (so_map_get(so_String, so_int, m, so_str("a")) != 1 || so_map_get(so_String, so_int, m, so_str("b")) != 2) {
            so_panic("named type literal");
        }
    }
    {
        // Named map type: set and get
        main_StrMap m = so_make_map(so_String, so_int, 2);
        so_map_set(so_String, so_int, m, so_str("x"), 10);
        so_map_set(so_String, so_int, m, so_str("y"), 20);
        if (so_map_get(so_String, so_int, m, so_str("x")) != 10 || so_map_get(so_String, so_int, m, so_str("y")) != 20) {
            so_panic("named type make");
        }
    }
    {
        // Named map type: comma-ok
        main_StrMap m = so_map_lit(so_String, so_int, 1, ((so_String[]){so_str("a")}), ((so_int[]){1}));
        so_int v = so_map_get(so_String, so_int, m, so_str("a"));
        bool ok = so_map_has(so_String, m, so_str("a"));
        if (!ok || v != 1) {
            so_panic("named type comma-ok");
        }
    }
    {
        // Named map type: range
        main_StrMap m = so_map_lit(so_String, so_int, 2, ((so_String[]){so_str("a"), so_str("b")}), ((so_int[]){1, 2}));
        so_int sum = 0;
        for (so_int _i = 0; _i < m->cap; _i++) {
            if (!m->used[_i]) continue;
            so_int v = ((so_int*)m->vals)[_i];
            sum += v;
        }
        if (sum != 3) {
            so_panic("named type range");
        }
    }
    {
        // Map assigned to another variable
        so_Map* m1 = so_map_lit(so_String, so_int, 1, ((so_String[]){so_str("a")}), ((so_int[]){1}));
        so_Map* m2 = m1;
        so_map_set(so_String, so_int, m2, so_str("a"), 99);
        // In Go maps are references, m1 sees the change.
        if (so_map_get(so_String, so_int, m1, so_str("a")) != 99) {
            so_panic("map alias");
        }
    }
    {
        // Map in if-init statement
        so_Map* m = so_map_lit(so_String, so_int, 1, ((so_String[]){so_str("a")}), ((so_int[]){1}));
        {
            so_int v = so_map_get(so_String, so_int, m, so_str("a"));
            bool ok = so_map_has(so_String, m, so_str("a"));
            if (ok) {
                if (v != 1) {
                    so_panic("if-init value");
                }
            }
        }
    }
    {
        // Map in if-init statement miss
        so_Map* m = so_map_lit(so_String, so_int, 1, ((so_String[]){so_str("a")}), ((so_int[]){1}));
        {
            bool ok = so_map_has(so_String, m, so_str("missing"));
            if (ok) {
                so_panic("if-init miss");
            }
        }
    }
    {
        // Map with a value type with no fields.
        so_Map* set = so_make_map(so_String, empty, 4);
        so_map_set(so_String, empty, set, so_str("a"), ((empty){}));
        empty v = so_map_get(so_String, empty, set, so_str("a"));
        (void)v;
        {
            bool ok = so_map_has(so_String, set, so_str("b"));
            if (ok) {
                so_panic("empty value miss");
            }
        }
        if (set->len != 1) {
            so_panic("empty value len");
        }
    }
    {
        // Map increment: m[key]++
        so_Map* m = so_map_lit(so_String, so_int, 1, ((so_String[]){so_str("a")}), ((so_int[]){1}));
        // m[key]++ and m[key] += 1 are not supported
        so_map_set(so_String, so_int, m, so_str("a"), so_map_get(so_String, so_int, m, so_str("a")) + 1);
        if (so_map_get(so_String, so_int, m, so_str("a")) != 2) {
            so_panic("increment");
        }
    }
    return 0;
}
