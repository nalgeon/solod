#include "main.h"

// -- Types --

typedef struct person person;

typedef struct person {
    so_int age;
} person;
typedef so_int* number;

// -- Forward declarations --
static so_int readCounter(void);

// -- Variables and constants --

// Package-level variables.
static so_unused so_int pkgInt = 42;
static so_unused double pkgFloat = 3.14;
static so_unused bool pkgBool = true;
static so_unused so_byte pkgByte = 'x';
static so_unused so_rune pkgRune = 0x672c;
static so_unused so_String pkgString = so_str("hello");
static so_unused so_Slice pkgSlice = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
static so_unused person pkgStruct = (person){.age = 42};
static so_unused person* pkgPtr = &pkgStruct;
static so_unused void* pkgAnyVal = &(so_int){42};
static so_unused void* pkgNil = NULL;

// Blank package-level variables.

// A package-level variable is visible to a function, so
// a call can read an assignment the same statement makes.
static so_unused so_int counter = 0;

// -- Implementation --

static so_int readCounter(void) {
    return counter;
}

int main(void) {
    {
        // Definition with var and explicit type.
        so_int vInt = 42;
        (void)vInt;
        double vFloat = 3.14;
        (void)vFloat;
        bool vBool = true;
        (void)vBool;
        so_byte vByte = 'x';
        (void)vByte;
        so_rune vRune = 0x672c;
        (void)vRune;
        so_String vString = so_str("hello");
        (void)vString;
        so_Slice vSlice = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        (void)vSlice;
        person vStruct = (person){.age = 42};
        person* vPtr = &vStruct;
        (void)vPtr;
        void* vAnyVal = &(so_int){42};
        (void)vAnyVal;
        void* vAnyPtr = vPtr;
        (void)vAnyPtr;
        void* vNil = NULL;
        (void)vNil;
    }
    {
        // Definition with var and type inference.
        so_int vInt = 42;
        (void)vInt;
        double vFloat = 3.14;
        (void)vFloat;
        bool vBool = true;
        (void)vBool;
        so_rune vByte = 'x';
        (void)vByte;
        so_rune vRune = 0x672c;
        (void)vRune;
        so_String vString = so_str("hello");
        (void)vString;
        so_Slice vSlice = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        (void)vSlice;
        person vStruct = (person){.age = 42};
        person* vPtr = &vStruct;
        (void)vPtr;
        void* vAnyVal = &(so_int){42};
        (void)vAnyVal;
        void* vAnyPtr = vPtr;
        (void)vAnyPtr;
        void* vNil = NULL;
        (void)vNil;
    }
    {
        // Definition with short variable declaration.
        so_int vInt = 42;
        (void)vInt;
        double vFloat = 3.14;
        (void)vFloat;
        bool vBool = true;
        (void)vBool;
        so_rune vByte = 'x';
        (void)vByte;
        so_rune vRune = 0x672c;
        (void)vRune;
        so_String vString = so_str("hello");
        (void)vString;
        so_Slice vSlice = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        (void)vSlice;
        person vStruct = (person){.age = 42};
        person* vPtr = &vStruct;
        (void)vPtr;
        void* vAnyVal = &(so_int){42};
        (void)vAnyVal;
        void* vAnyPtr = vPtr;
        (void)vAnyPtr;
        void* vNil = NULL;
        (void)vNil;
    }
    {
        // Zero values.
        so_int vInt = 0;
        (void)vInt;
        double vFloat = 0;
        (void)vFloat;
        bool vBool = false;
        (void)vBool;
        so_byte vByte = 0;
        (void)vByte;
        so_rune vRune = 0;
        (void)vRune;
        so_String vString = so_str("");
        (void)vString;
        so_Slice vSlice = {};
        (void)vSlice;
        person vStruct = {};
        (void)vStruct;
        person* vPtr = NULL;
        (void)vPtr;
        void* vNil = NULL;
        (void)vNil;
    }
    {
        // Multiple typed variable declaration.
        so_int a = 11, b = 22, c = 33;
        (void)a;
        (void)b;
        (void)c;
        so_byte b1 = 'a', b2 = 'b';
        (void)b1;
        (void)b2;
        so_String s1 = so_str("foo"), s2 = so_str("bar");
        (void)s1;
        (void)s2;
        so_Slice a1 = (so_Slice){(so_int[2]){1, 2}, 2, 2}, a2 = (so_Slice){(so_int[2]){3, 4}, 2, 2};
        (void)a1;
        (void)a2;
        person p1 = (person){.age = 42}, p2 = (person){.age = 43};
        (void)p1;
        (void)p2;
        person* ptr1 = &p1;
        person* ptr2 = &p2;
        (void)ptr1;
        (void)ptr2;
        number n1 = &p1.age, n2 = &p2.age;
        (void)n1;
        (void)n2;
    }
    {
        // Multiple untyped variable declaration.
        so_int vInt = 42;
        double vFloat = 3.14;
        bool vBool = true;
        (void)vInt;
        (void)vFloat;
        (void)vBool;
        so_rune vByte = 'x', vRune = 0x672c;
        so_String vString = so_str("hello");
        (void)vByte;
        (void)vRune;
        (void)vString;
        so_Slice vSlice = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        person vStruct = (person){.age = 42};
        (void)vSlice;
        (void)vStruct;
        person* ptr1 = &vStruct;
        person* ptr2 = &vStruct;
        (void)ptr1;
        (void)ptr2;
        number n1 = (number)(&vStruct.age), n2 = (number)(&vStruct.age);
        (void)n1;
        (void)n2;
    }
    {
        // Multiple variable declaration with short variable declaration.
        so_int vInt = 42;
        double vFloat = 3.14;
        bool vBool = true;
        (void)vInt;
        (void)vFloat;
        (void)vBool;
        so_rune vByte = 'x', vRune = 0x672c;
        so_String vString = so_str("hello");
        (void)vByte;
        (void)vRune;
        (void)vString;
        so_Slice vSlice = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        person vStruct = (person){.age = 42};
        (void)vSlice;
        (void)vStruct;
        person* ptr1 = &vStruct;
        person* ptr2 = &vStruct;
        (void)ptr1;
        (void)ptr2;
        number n1 = (number)(&vStruct.age), n2 = (number)(&vStruct.age);
        (void)n1;
        (void)n2;
        void* u1 = (void*)(&vStruct);
        void* u2 = (void*)(&vStruct);
        (void)u1;
        (void)u2;
    }
    {
        // Discarding values with blank identifier.
        so_int v1 = 11;
        so_int v2 = 22;
        so_int v3 = 51;
        (void)52;
        (void)61;
        so_int v4 = 62;
        (void)71;
        (void)72;
        (void)81;
        (void)v1;
        (void)v2;
        (void)v3;
        (void)v4;
    }
    {
        // Partial redeclaration with short variable declaration.
        so_int a = 11, x = 100;
        so_int b = 22;
        x = 200;
        x = 300;
        so_int c = 33;
        (void)a;
        (void)b;
        (void)c;
        (void)x;
    }
    {
        // Multiple assignment.
        so_int a = 11, b = 22;
        a = 33;
        b = 44;
        so_int x = 55, y = 66;
        a = x;
        b = y;
        if (a != 55 || b != 66) {
            so_panic("multiple assignment failed");
        }
        person p = (person){.age = 42};
        person* ptr1 = NULL;
        person* ptr2 = NULL;
        ptr1 = &p;
        ptr2 = &p;
        (void)ptr1;
        (void)ptr2;
        number n1 = NULL, n2 = NULL;
        n1 = (number)(&p.age);
        n2 = (number)(&p.age);
        (void)n1;
        (void)n2;
    }
    {
        // Evaluates the whole right side before it assigns anything,
        // so a swap works and a call sees the values from before.
        so_int a = 11, b = 22;
        so_int _asg1 = b;
        so_int _asg2 = a;
        a = _asg1;
        b = _asg2;
        if (a != 22 || b != 11) {
            so_panic("swap failed");
        }
        so_Slice s = (so_Slice){(so_int[2]){10, 20}, 2, 2};
        so_int i = 0, j = 1;
        so_int _asg3 = so_at(so_int, s, j);
        so_int _asg4 = so_at(so_int, s, i);
        so_at(so_int, s, i) = _asg3;
        so_at(so_int, s, j) = _asg4;
        if (so_at(so_int, s, 0) != 20 || so_at(so_int, s, 1) != 10) {
            so_panic("slice swap failed");
        }
        counter = 7;
        so_int _asg5 = readCounter();
        counter = 8;
        b = _asg5;
        if (counter != 8 || b != 7) {
            so_panic("assignment order failed");
        }
    }
    {
        // Variable shadowing.
        so_int age = 30;
        person p = (person){.age = age};
        {
            so_int age = p.age;
            (void)age;
        }
        {
            person age = (person){.age = 40};
            (void)age;
        }
    }
    return 0;
}
