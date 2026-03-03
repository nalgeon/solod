#include "main.h"

// -- Forward declarations (functions and methods) --
static main_Person newPerson(so_String name);

// -- Implementation --

static main_Person newPerson(so_String name) {
    main_Person p = (main_Person){.name = name};
    p.age = 42;
    return p;
}

int main(void) {
    {
        // Primitive types.
        main_ID id = 123;
        (void)id;
        so_int aid = 456;
        (void)aid;
        main_AlsoID alsoID = 789;
        (void)alsoID;
        main_Rune r = U'A';
        (void)r;
    }
    {
        // Complex types.
        main_Name n = so_strlit("Alice");
        (void)n;
        main_IntArray arr = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        (void)arr;
        main_IntSlice slice = (so_Slice){(so_int[3]){4, 5, 6}, 3, 3};
        (void)slice;
    }
    {
        // Struct types.
        main_Person bob = (main_Person){so_strlit("Bob"), 20};
        (void)bob;
        main_Person alice = (main_Person){.name = so_strlit("Alice"), .age = 30};
        (void)alice;
        main_Person fred = (main_Person){.name = so_strlit("Fred")};
        (void)fred;
        main_Person* ann = &(main_Person){.name = so_strlit("Ann"), .age = 40};
        *ann = newPerson(so_strlit("Jon"));
        (void)ann;
        main_Person sean = {0};
        sean.name = so_strlit("Sean");
        sean.age = 50;
        main_Person* sp = &sean;
        sp->age = 51;
        (void)sean;
        so_auto dog = (struct {
            so_String name;
            bool isGood;
        }){
            .name = so_strlit("Rex"),
            .isGood = true,
        };
        (void)dog;
    }
}
