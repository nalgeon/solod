#include "main.h"

typedef struct person {
    so_int age;
} person;

int main(void) {
    {
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
        so_Slice vSlice = {(so_int[3]){1, 2, 3}, 3, 3};
        (void)vSlice;
        person vStruct = {.age = 42};
        person* vPtr = &vStruct;
        (void)vPtr;
        void* vAnyVal = &(so_int){42};
        (void)vAnyVal;
        void* vAnyPtr = vPtr;
        (void)vAnyPtr;
    }
    {
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
        so_Slice vSlice = {(so_int[3]){1, 2, 3}, 3, 3};
        (void)vSlice;
        person vStruct = {.age = 42};
        person* vPtr = &vStruct;
        (void)vPtr;
        void* vAnyVal = &(so_int){42};
        (void)vAnyVal;
        void* vAnyPtr = vPtr;
        (void)vAnyPtr;
    }
    {
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
        so_Slice vSlice = {(so_int[3]){1, 2, 3}, 3, 3};
        (void)vSlice;
        person vStruct = {.age = 42};
        person* vPtr = &vStruct;
        (void)vPtr;
        void* vAnyVal = &(so_int){42};
        (void)vAnyVal;
        void* vAnyPtr = vPtr;
        (void)vAnyPtr;
    }
    {
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
        so_Slice a1 = {(so_int[2]){1, 2}, 2, 2}, a2 = {(so_int[2]){3, 4}, 2, 2};
        (void)a1;
        (void)a2;
        person p1 = {.age = 42}, p2 = {.age = 43};
        (void)p1;
        (void)p2;
    }
    {
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
        so_Slice vSlice = {(so_int[3]){1, 2, 3}, 3, 3};
        person vStruct = {.age = 42};
        (void)vSlice;
        (void)vStruct;
    }
    {
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
        so_Slice vSlice = {(so_int[3]){1, 2, 3}, 3, 3};
        person vStruct = {.age = 42};
        (void)vSlice;
        (void)vStruct;
    }
}
