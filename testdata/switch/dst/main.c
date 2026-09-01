#include "main.h"

// -- Types --
typedef so_int array[3];

// -- Forward declarations --
static so_int inc(void);
static so_String name(void);
static so_int twice(so_int v);

// -- Variables and constants --
static so_unused so_int calls = 0;

// -- Implementation --

// inc counts its calls, so a switch on it
// shows how many times the tag is evaluated.
static so_int inc(void) {
    calls++;
    return calls;
}

static so_String name(void) {
    calls++;
    return so_str("hello");
}

static so_int twice(so_int v) {
    return v * 2;
}

so_int main_Square_Area(void* self) {
    main_Square* s = self;
    return s->side * s->side;
}

int main(void) {
    {
        // Empty switch statement.
        if (false) {
        }
    }
    {
        // Empty switch statement with a tag.
        so_int i = 1;
        (void)i;
        if (false) {
        }
    }
    {
        // Switch on int with cases and default.
        so_int i = 2;
        {
            so_int _sw1 = i;
            if (_sw1 == 1) {
                so_panic("unexpected i == 1");
            } else if (_sw1 == 2) {
            } else if (_sw1 == 3) {
                so_panic("unexpected i == 3");
            } else {
                so_panic("unexpected default");
            }
        }
    }
    {
        // Tagless switch (bool conditions).
        so_int x = 10;
        if (x > 100) {
            so_panic("unexpected x > 100");
        } else if (x > 0) {
        } else {
            so_panic("unexpected default");
        }
    }
    {
        // Multiple values per case.
        so_int y = 3;
        {
            so_int _sw2 = y;
            if (_sw2 == 1 || _sw2 == 2 || _sw2 == 3) {
                if (y != 3) {
                    so_panic("want y == 3");
                }
            } else if (_sw2 == 4 || _sw2 == 5 || _sw2 == 6) {
                so_panic("unexpected y == 4, 5, 6");
            } else {
                so_panic("unexpected default");
            }
        }
    }
    {
        // Switch with init statement.
        {
            so_int n = 42;
            so_int _sw3 = n;
            if (_sw3 == 42) {
            } else {
                so_panic("unexpected default");
            }
        }
    }
    {
        // Switch on string.
        so_String s = so_str("hello");
        {
            so_String _sw4 = s;
            if (so_string_eq(_sw4, so_str("hello"))) {
            } else if (so_string_eq(_sw4, so_str("bye"))) {
                so_panic("unexpected s == bye");
            } else {
                so_panic("unexpected default");
            }
        }
    }
    {
        // Cases without default. There is no default to catch a missed
        // case, so a flag records that the switch selected one.
        so_int z = 5;
        bool matched = false;
        {
            so_int _sw5 = z;
            if (_sw5 == 1) {
                so_panic("unexpected z == 1");
            } else if (_sw5 == 5) {
                matched = true;
            }
        }
        if (!matched) {
            so_panic("want z == 5");
        }
    }
    {
        // The tag is evaluated only once.
        calls = 0;
        {
            so_int _sw6 = inc();
            if (_sw6 == 2) {
                so_panic("unexpected inc() == 2");
            } else if (_sw6 == 3) {
                so_panic("unexpected inc() == 3");
            } else {
                if (calls != 1) {
                    so_panic("want 1 inc() call");
                }
            }
        }
    }
    {
        // Same for a string tag.
        calls = 0;
        {
            so_String _sw7 = name();
            if (so_string_eq(_sw7, so_str("bye"))) {
                so_panic("unexpected name() == bye");
            } else if (so_string_eq(_sw7, so_str("hello"))) {
                if (calls != 1) {
                    so_panic("want 1 name() call");
                }
            } else {
                so_panic("unexpected default");
            }
        }
    }
    {
        // Init statement and a tag to evaluate.
        calls = 0;
        {
            so_int base = 10;
            so_int _sw8 = base + inc();
            if (_sw8 == 11) {
                if (calls != 1) {
                    so_panic("want 1 inc() call");
                }
            } else {
                so_panic("unexpected default");
            }
        }
    }
    {
        // No cases to compare against, but the tag still runs.
        calls = 0;
        (void)inc();
        {
            if (calls != 1) {
                so_panic("want 1 inc() call");
            }
        }
    }
    {
        // A case expression may convert and use builtins.
        so_String s = so_str("abc");
        int32_t n = 3;
        {
            so_int _sw9 = so_len(s);
            if (_sw9 == (so_int)(n)) {
            } else {
                so_panic("unexpected default");
            }
        }
    }
    {
        // Case expressions run in order, as in Go, so they may call functions.
        calls = 0;
        if (inc() > 5) {
            so_panic("unexpected inc() > 5");
        } else if (inc() > 0) {
            if (calls != 2) {
                so_panic("want 2 inc() calls");
            }
        } else {
            so_panic("unexpected default");
        }
    }
    {
        // The same for a tagged switch: the tag is already in a temporary,
        // so a case expression may change it.
        calls = 0;
        {
            so_int _sw10 = calls;
            if (_sw10 == inc()) {
                so_panic("unexpected calls == inc()");
            } else if (_sw10 == 0) {
                if (calls != 1) {
                    so_panic("want 1 inc() call");
                }
            } else {
                so_panic("unexpected default");
            }
        }
    }
    {
        // Once a case matches, the later case expressions do not run,
        // neither in the same case nor in the following ones.
        calls = 0;
        {
            so_int _sw11 = 1;
            if (_sw11 == 1 || _sw11 == inc()) {
            } else if (_sw11 == inc()) {
                so_panic("unexpected second case");
            } else {
                so_panic("unexpected default");
            }
        }
        if (calls != 0) {
            so_panic("want no inc() calls");
        }
    }
    {
        // Case expressions that C groups differently than Go.
        so_int a = 1, b = 2;
        bool ok = true;
        {
            bool _sw12 = ok;
            if (_sw12 == (a < b && b < 3)) {
            } else {
                so_panic("unexpected default");
            }
        }
        {
            bool _sw13 = ok;
            if (_sw13 == (a == b)) {
                so_panic("unexpected a == b");
            } else {
            }
        }
        {
            so_int _sw14 = 3;
            if (_sw14 == (a + b)) {
            } else {
                so_panic("unexpected default");
            }
        }
        {
            so_int _sw15 = 3;
            if (_sw15 == (a | b)) {
            } else {
                so_panic("unexpected default");
            }
        }
    }
    {
        // A binary case expression on a string tag compares with a call.
        so_String s1 = so_str("he"), s2 = so_str("llo");
        {
            so_String _sw16 = so_str("hello");
            if (so_string_eq(_sw16, (so_string_add(s1, s2)))) {
            } else {
                so_panic("unexpected default");
            }
        }
    }
    {
        // Nested switches use separate tag temporaries.
        so_int x = 1, y = 2;
        {
            so_int _sw17 = x;
            if (_sw17 == 1) {
                {
                    so_int _sw18 = y;
                    if (_sw18 == 2) {
                    } else {
                        so_panic("unexpected inner default");
                    }
                }
            } else {
                so_panic("unexpected outer default");
            }
        }
    }
    {
        // A default-only body gets its own scope, and the tag still runs.
        calls = 0;
        so_int v = 1;
        (void)inc();
        {
            so_int v = 2;
            if (v != 2) {
                so_panic("want inner v == 2");
            }
        }
        if (calls != 1) {
            so_panic("want 1 inc() call");
        }
        if (v != 1) {
            so_panic("want outer v == 1");
        }
    }
    {
        // A break in a case body is not supported, but a labeled break is.
        so_int steps = 0;
        done:;
        {
            so_int _sw19 = 1;
            if (_sw19 == 1) {
                steps++;
                goto done_end;
            } else {
                so_panic("unexpected default");
            }
        }
        done_end:;
        if (steps != 1) {
            so_panic("want 1 step");
        }
    }
    {
        // A break inside a loop in a case body leaves the loop.
        so_int n = 0;
        {
            so_int _sw20 = 1;
            if (_sw20 == 1) {
                for (so_int _i = 0; _i < 5; _i++) {
                    n++;
                    if (n == 2) {
                        break;
                    }
                }
            } else {
                so_panic("unexpected default");
            }
        }
        if (n != 2) {
            so_panic("want n == 2");
        }
    }
    {
        // Switch on a slice compares to nil.
        so_Slice s = {};
        {
            so_Slice _sw21 = s;
            if (_sw21.cap == 0) {
            } else {
                so_panic("unexpected default");
            }
        }
    }
    {
        // Switch on a map compares to nil.
        so_Map* m = NULL;
        {
            so_Map* _sw22 = m;
            if (_sw22 == NULL) {
            } else {
                so_panic("unexpected default");
            }
        }
    }
    {
        // Switch on an interface compares to nil.
        main_Shape sh = (main_Shape){.self = &(main_Square){2}, .Area = main_Square_Area};
        {
            main_Shape _sw23 = sh;
            if (_sw23.self == NULL) {
                so_panic("unexpected sh == nil");
            } else {
                if (main_Shape_Area(sh) != 4) {
                    so_panic("want sh.Area() == 4");
                }
            }
        }
    }
    {
        // Pointer to array. A pointer to an unnamed array type
        // is not supported, so the case uses a named type.
        array a = {1, 2, 3};
        array* p = &a;
        {
            array* _sw24 = p;
            if (_sw24 == NULL) {
                so_panic("unexpected p == nil");
            } else if (_sw24 == &a) {
            } else {
                so_panic("unexpected default");
            }
        }
    }
    {
        // Function pointer.
        so_int (*fn)(so_int) = twice;
        {
            so_int (*_sw25)(so_int) = fn;
            if (_sw25 == NULL) {
                so_panic("unexpected fn == nil");
            } else {
                if (fn(2) != 4) {
                    so_panic("want fn(2) == 4");
                }
            }
        }
    }
    return 0;
}
