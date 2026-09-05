#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_Empty main_Empty;
typedef struct main_Person main_Person;
typedef struct main_Benchmark main_Benchmark;
typedef main_Person main_Human;
typedef main_Person main_Employee;

// Primitive types.
// same type in C
typedef so_int main_ID;

// also int
typedef so_int main_AlsoID;

// also int
typedef so_int main_AliasInt;

// also int
typedef so_int main_AliasID;
typedef so_rune main_Rune;

// Complex types.
typedef so_String main_Name;
typedef so_int main_IntArray[3];
typedef so_Slice main_IntSlice;
typedef so_int* main_IntPtr;
typedef void* main_Any;

typedef struct main_Empty {
} main_Empty;

// Struct type.
typedef struct main_Person {
    so_String name;
    so_int age;
} main_Person;

// Alias for a struct type.
typedef main_Person main_Human;
typedef main_Person main_Employee;

// Alias for a pointer type.
typedef main_Person* main_HumanPtr;

// Alias for a function type.
typedef so_String (*main_Namer)(main_Person);

// Inner struct.
typedef struct main_Benchmark {
    so_String name;
    struct {
        so_int n;
        so_int i;
    } loop;
} main_Benchmark;

// -- Functions and methods --

// Methods on aliases.
so_int main_Person_Age(void* self);
so_int main_ID_GetVal(main_ID aid);
so_int main_ID_GetPtr(void* self);
