#pragma once
#include "so/builtin/builtin.h"

// -- Types --

// Primitive types.
// not a different type
typedef so_int main_ID;

// also int
typedef so_int main_AliasedID;

// also int
typedef so_int main_AlsoID;
typedef int32_t main_Rune;

// Complex types.
typedef so_String main_Name;
typedef so_Slice main_IntArray;
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
