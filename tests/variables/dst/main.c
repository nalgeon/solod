#include "main.h"

// -- Forward declarations (types) --
typedef struct person person;

// -- Implementation --

typedef struct person {
    so_int age;
} person;

int main(void) {
    {
        // Definition with var and explicit type.
        so_int vInt = 42;
        (void)vInt;
        double vFloat = 3.14;
        (void)vFloat;
        bool vBool = true;
        (void)vBool;
        uint8_t vByte = 'x';
        (void)vByte;
        int32_t vRune = U'本';
        (void)vRune;
        so_String vString = so_strlit("hello");
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
        int32_t vByte = U'x';
        (void)vByte;
        int32_t vRune = U'本';
        (void)vRune;
        so_String vString = so_strlit("hello");
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
        int32_t vByte = U'x';
        (void)vByte;
        int32_t vRune = U'本';
        (void)vRune;
        so_String vString = so_strlit("hello");
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
        uint8_t vByte = 0;
        (void)vByte;
        int32_t vRune = 0;
        (void)vRune;
        so_String vString = so_strlit("");
        (void)vString;
        so_Slice vSlice = {0};
        (void)vSlice;
        person vStruct = {0};
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
        uint8_t b1 = 'a', b2 = 'b';
        (void)b1;
        (void)b2;
        so_String s1 = so_strlit("foo"), s2 = so_strlit("bar");
        (void)s1;
        (void)s2;
        so_Slice a1 = (so_Slice){(so_int[2]){1, 2}, 2, 2}, a2 = (so_Slice){(so_int[2]){3, 4}, 2, 2};
        (void)a1;
        (void)a2;
        person p1 = (person){.age = 42}, p2 = (person){.age = 43};
        (void)p1;
        (void)p2;
    }
    {
        // Multiple untyped variable declaration.
        so_int vInt = 42;
        double vFloat = 3.14;
        bool vBool = true;
        (void)vInt;
        (void)vFloat;
        (void)vBool;
        int32_t vByte = U'x', vRune = U'本';
        so_String vString = so_strlit("hello");
        (void)vByte;
        (void)vRune;
        (void)vString;
        so_Slice vSlice = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        person vStruct = (person){.age = 42};
        (void)vSlice;
        (void)vStruct;
    }
    {
        // Multiple variable declaration with short variable declaration.
        so_int vInt = 42;
        double vFloat = 3.14;
        bool vBool = true;
        (void)vInt;
        (void)vFloat;
        (void)vBool;
        int32_t vByte = U'x', vRune = U'本';
        so_String vString = so_strlit("hello");
        (void)vByte;
        (void)vRune;
        (void)vString;
        so_Slice vSlice = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3};
        person vStruct = (person){.age = 42};
        (void)vSlice;
        (void)vStruct;
    }
    {
        // Discarding values with blank identifier.
        so_int v1 = 11;
        so_int v2 = 22;
        so_int v3 = 51;
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
}
