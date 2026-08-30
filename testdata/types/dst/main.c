#include "main.h"

// -- Types --

typedef struct point point;
typedef struct empty empty;
typedef struct tagged tagged;

// Unexported struct type.
typedef struct point {
    so_int x;
    so_int y;
} point;

typedef struct empty {
} empty;

// Struct with an empty first field.
typedef struct tagged {
    empty e;
    so_int n;
} tagged;

// -- Forward declarations --
static main_Person newPerson(so_String name);
static so_String personName(main_Person p);

// -- Implementation --

static main_Person newPerson(so_String name) {
    main_Person p = (main_Person){.name = name};
    p.age = 42;
    return p;
}

// Methods on aliases.
so_int main_Person_Age(void* self) {
    main_Person* h = self;
    return h->age;
}

so_int main_ID_GetVal(main_ID aid) {
    return (so_int)(aid);
}

so_int main_ID_GetPtr(void* self) {
    main_ID* aid = self;
    return (so_int)(*aid);
}

static so_String personName(main_Person p) {
    return p.name;
}

int main(void) {
    {
        // Primitive types.
        main_ID id = 123;
        (void)id;
        main_ID aid = 456;
        (void)aid;
        main_AlsoID alsoID = 789;
        (void)alsoID;
        main_Rune r = 'A';
        (void)r;
    }
    {
        // Complex types.
        main_Name n = so_str("Alice");
        (void)n;
        main_IntArray arr = {1, 2, 3};
        (void)arr;
        main_IntSlice slice = (so_Slice){(so_int[3]){4, 5, 6}, 3, 3};
        (void)slice;
    }
    {
        // Struct types.
        main_Person bob = (main_Person){so_str("Bob"), 20};
        (void)bob;
        main_Person alice = (main_Person){.name = so_str("Alice"), .age = 30};
        (void)alice;
        main_Person fred = (main_Person){.name = so_str("Fred")};
        (void)fred;
        main_Person* ann = &(main_Person){.name = so_str("Ann"), .age = 40};
        *ann = newPerson(so_str("Jon"));
        (void)ann;
        main_Person sean = {};
        sean.name = so_str("Sean");
        sean.age = 50;
        main_Person* sp = &sean;
        sp->age = 51;
        (void)sean;
    }
    {
        // Empty struct types.
        empty e = {};
        (void)e;
        empty* ep = &(empty){};
        (void)ep;
        empty earr[3] = {};
        (void)earr;
        tagged tag = {};
        tag.n = 60;
        if (tag.n != 60) {
            so_panic("tag.n != 60");
        }
    }
    {
        // Anonymous struct type.
        so_auto dog = (struct {
            so_String name;
            bool isGood;
        }){
            .name = so_str("Rex"),
            .isGood = true,
        };
        (void)dog;
    }
    {
        // Named struct type inside a function.
        typedef struct Point {
            so_int x;
            so_int y;
        } Point;
        Point p = (Point){1, 2};
        (void)p;
    }
    {
        // Inner struct.
        main_Benchmark b1 = (main_Benchmark){.name = so_str("Test")};
        b1.loop.n = 100;
        if (b1.loop.n != 100) {
            so_panic("b1.loop.n != 100");
        }
        main_Benchmark b2 = (main_Benchmark){.name = so_str("Test2"), .loop = {.n = 200, .i = 10}};
        if (b2.loop.n != 200) {
            so_panic("b2.loop.n != 200");
        }
        main_Benchmark b3 = (main_Benchmark){.name = so_str("Test3"), .loop = {300, 30}};
        if (b3.loop.n != 300) {
            so_panic("b3.loop.n != 300");
        }
        main_Benchmark b4 = {};
        if (b4.loop.n != 0) {
            so_panic("b4.loop.n != 0");
        }
    }
    {
        // Type aliases.
        main_Person h = (main_Person){.name = so_str("Alice"), .age = 30};
        so_int age = main_Person_Age(&h);
        if (age != 30) {
            so_panic("h.Age() != 30");
        }
        main_ID aid = (main_ID)(123);
        if (main_ID_GetVal(aid) != 123) {
            so_panic("aid.GetVal() != 123");
        }
        if (main_ID_GetPtr(&aid) != 123) {
            so_panic("aid.GetPtr() != 123");
        }
        main_ID id = aid;
        if (main_ID_GetVal(id) != 123) {
            so_panic("id.GetVal() != 123");
        }
    }
    {
        // Alias for a pointer type: the method call unaliases the receiver.
        main_Person* hp = &(main_Person){.name = so_str("Zoe"), .age = 60};
        if (main_Person_Age(hp) != 60) {
            so_panic("hp.Age() != 60");
        }
    }
    {
        // Alias for a function type.
        so_String (*namer)(main_Person) = personName;
        if (so_string_ne(namer((main_Person){.name = so_str("Ivy")}), so_str("Ivy"))) {
            so_panic("namer(Ivy) != Ivy");
        }
    }
    {
        // Conversion between structs with the same underlying type.
        main_Person p = (main_Person){.name = so_str("Nina"), .age = 25};
        main_Employee e = (main_Employee)(p);
        if (so_string_ne(e.name, so_str("Nina")) || e.age != 25) {
            so_panic("Employee(p) lost the fields");
        }
        main_Person back = (main_Person)(e);
        if (so_string_ne(back.name, so_str("Nina")) || back.age != 25) {
            so_panic("Person(e) lost the fields");
        }
    }
    return 0;
}
