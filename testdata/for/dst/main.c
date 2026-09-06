#include "main.h"

// -- Types --

typedef struct node node;

typedef struct node {
    so_int value;
    node* next;
} node;

// -- Forward declarations --
static so_int two(void);
static so_int incr(so_int* n);
static void testBasic(void);
static void testClause(void);

// -- Implementation --

static so_int two(void) {
    return 2;
}

static so_int incr(so_int* n) {
    (*n)++;
    return *n;
}

int main(void) {
    testBasic();
    testClause();
    return 0;
}

static void testBasic(void) {
    so_int i = 1;
    for (; i <= 3;) {
        so_println("%" PRIdINT, i);
        i = i + 1;
    }
    for (so_int j = 0; j < 3; j++) {
        so_println("%" PRIdINT, j);
    }
    so_int start = 5;
    for (start--; start >= 0; start--) {
        if (start == 2) {
            break;
        }
    }
    for (start = 5; start >= 0; start--) {
    }
    for (so_int k = 0; k < 3; k++) {
        so_println("%s %" PRIdINT, "range", k);
    }
    // The loop assigns to k2, so it keeps the value of the last iteration.
    so_int k2 = 0;
    for (so_int _k2i = 0; _k2i < 3; _k2i++) {
        k2 = _k2i;
    }
    if (k2 != 2) {
        so_panic("want k2 == 2");
    }
    for (so_int _i = 0; _i < 3; _i++) {
    }
    for (;;) {
        so_println("%s", "loop");
        break;
    }
    for (so_int n = 0; n < 6; n++) {
        if (n % 2 == 0) {
            continue;
        }
        so_println("%" PRIdINT, n);
    }
}

static void testClause(void) {
    {
        // Init statements that go before the loop.
        {
            so_int i = 0, j = 3;
            for (; i < j; i++) {
                (void)i;
            }
        }
        so_int src[2] = {1, 2};
        {
            so_int arr[2];
            memcpy(arr, src, sizeof(arr));
            for (; false;) {
                (void)arr;
            }
        }
        so_Map* m = so_make_map(so_String, so_int, 2);
        {
            so_int v = so_map_get(so_String, so_int, m, so_str("k"));
            bool ok = so_map_has(so_String, m, so_str("k"));
            for (; false;) {
                (void)v;
                (void)ok;
            }
        }
        {
            so_map_set(so_String, so_int, m, so_str("k"), 5);
            for (; false;) {
            }
        }
        if (so_map_get(so_String, so_int, m, so_str("k")) != 5) {
            so_panic("want m[\"k\"] == 5");
        }
        void* a = NULL;
        {
            a = &(so_int){1};
            for (; true;) {
                if ((*(so_int*)a) != 1) {
                    so_panic("want a.(int) == 1");
                }
                break;
            }
        }
        for ((void)0; false;) {
        }
    }
    {
        // Post statements that fit in the clause.
        so_int sum = 0;
        for (so_int i = 0; i < 9; i += 3) {
            sum += i;
        }
        if (sum != 9) {
            so_panic("want sum == 9");
        }
        so_int bits = 0;
        for (so_int i = 8; i > 0; i >>= 1) {
            bits++;
        }
        if (bits != 4) {
            so_panic("want bits == 4");
        }
        // A map variable emits so_Map*, a scalar.
        so_Map* mm = so_make_map(so_String, so_int, 2);
        for (so_int i = 0; i < 1; mm = NULL) {
            i++;
        }
        if (mm != NULL) {
            so_panic("want mm == nil");
        }
        // The division keeps its zero divisor guard.
        so_int n = 10;
        for (so_int i = 0; i < 1; n = so_div(n, two())) {
            i++;
        }
        if (n != 5) {
            so_panic("want n == 5");
        }
        // A pointer walk and a discarded call.
        node tail = (node){.value = 2};
        node head = (node){.value = 1, .next = &tail};
        so_int count = 0;
        for (node* p = &head; p != NULL; p = p->next) {
            count++;
        }
        if (count != 2) {
            so_panic("want count == 2");
        }
        so_int calls = 0;
        for (so_int i = 0; i < 2; (void)incr(&calls)) {
            i++;
        }
        if (calls != 2) {
            so_panic("want calls == 2");
        }
    }
}
