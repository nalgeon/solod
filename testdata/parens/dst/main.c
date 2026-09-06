#include "main.h"

// -- Types --

typedef struct point point;
typedef struct rect rect;

typedef struct point {
    so_int x;
    so_int y;
} point;

typedef struct shape {
    void* self;
    so_int (*area)(void* self);
} shape;

static inline so_unused so_int shape_area(shape self) {
    return self.area(self.self);
}

typedef struct rect {
    so_int width;
    so_int height;
} rect;

// -- Forward declarations --
static so_int rect_area(void* self);
static so_R_int_int pair(void);

// -- Implementation --

static so_int rect_area(void* self) {
    rect* r = self;
    return r->width * r->height;
}

static so_R_int_int pair(void) {
    return (so_R_int_int){.val = 7, .val2 = 8};
}

int main(void) {
    {
        // Parentheses around a blank identifier.
        (void)0;
        (void)0;
    }
    {
        // Parentheses around a variable, a dereference, an index, and a field.
        so_int n = 0;
        n = 1;
        if (n != 1) {
            so_panic("want n == 1");
        }
        so_int* p = &n;
        *p = 2;
        if (n != 2) {
            so_panic("want n == 2");
        }
        so_Slice s = (so_Slice){(so_int[1]){0}, 1, 1};
        so_at(so_int, s, 0) = 3;
        if (so_at(so_int, s, 0) != 3) {
            so_panic("want s[0] == 3");
        }
        point pt = (point){};
        pt.x = 4;
        if (pt.x != 4) {
            so_panic("want pt.x == 4");
        }
    }
    {
        // Parentheses around a map index.
        so_Map* m = so_make_map(so_String, so_int, 4);
        so_map_set(so_String, so_int, m, so_str("k"), 5);
        if (so_map_get(so_String, so_int, m, so_str("k")) != 5) {
            so_panic("want m[\"k\"] == 5");
        }
    }
    {
        // Parentheses around a target of a multiple assignment.
        so_int n = 0;
        (void)1;
        n = 2;
        if (n != 2) {
            so_panic("want n == 2");
        }
        so_int a = 0, b = 0;
        a = 3;
        b = 4;
        if (a != 3 || b != 4) {
            so_panic("want a == 3 and b == 4");
        }
    }
    {
        // Parentheses around a target of a multi-return assignment.
        so_int n = 0;
        so_R_int_int _res1 = pair();
        n = _res1.val2;
        if (n != 8) {
            so_panic("want n == 8");
        }
    }
    {
        // Parentheses around a target of a comma-ok assignment.
        so_Map* m = so_make_map(so_String, so_int, 4);
        so_map_set(so_String, so_int, m, so_str("k"), 9);
        so_int n = 0;
        bool ok = false;
        n = so_map_get(so_String, so_int, m, so_str("k"));
        ok = so_map_has(so_String, m, so_str("k"));
        if (n != 9 || !ok) {
            so_panic("want n == 9 and ok");
        }
        ok = so_map_has(so_String, m, so_str("none"));
        if (ok) {
            so_panic("want !ok");
        }
        shape sh = (shape){.self = &(rect){.width = 2, .height = 3}, .area = rect_area};
        ok = (sh.area == rect_area);
        if (!ok) {
            so_panic("want ok");
        }
    }
    {
        // Parentheses around a range key and a range value.
        so_int i = 0, v = 0;
        so_int sum = 0;
        so_Slice sl = (so_Slice){(so_int[3]){10, 20, 30}, 3, 3};
        for (so_int _ii = 0; _ii < so_len(sl); _ii++) {
            i = _ii;
            v = so_at(so_int, sl, _ii);
            sum += i * v;
        }
        if (sum != 80) {
            so_panic("want sum == 80");
        }
        so_int arr[2] = {1, 2};
        so_int total = 0;
        for (so_int _ii = 0; _ii < 2; _ii++) {
            i = _ii;
            v = arr[_ii];
            total += i + v;
        }
        if (total != 4) {
            so_panic("want total == 4");
        }
        so_int count = 0;
        for (so_int _ii = 0; _ii < 3; _ii++) {
            i = _ii;
            count += i;
        }
        if (count != 3) {
            so_panic("want count == 3");
        }
        so_rune r = 0;
        so_int runes = 0;
        for (so_int _ii = 0, _iw = 0; _ii < so_len(so_str("ab")); _ii += _iw) {
            i = _ii;
            _iw = 0;
            r = so_utf8_decode(so_str("ab"), _ii, &_iw);
            runes += i + (so_int)(r);
        }
        if (runes != 'a' + 'b' + 1) {
            so_panic("want runes == 'a'+'b'+1");
        }
        so_Map* m = so_make_map(so_String, so_int, 4);
        so_map_set(so_String, so_int, m, so_str("k"), 11);
        so_String k = so_str("");
        {
            so_Map* _m = m;
            for (so_int _i = 0; _m != NULL && _i < _m->cap; _i++) {
                if (!_m->used[_i]) continue;
                k = ((so_String*)_m->keys)[_i];
                if (so_string_ne(k, so_str("k"))) {
                    so_panic("want k == \"k\"");
                }
            }
        }
        if (so_string_ne(k, so_str("k"))) {
            so_panic("want k set after the loop");
        }
    }
    return 0;
}
