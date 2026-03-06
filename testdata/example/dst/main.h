#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_Person {
    so_String Name;
    so_int Age;
    so_Slice Nums;
} main_Person;

// -- Functions and methods --
so_int main_Person_Sleep(void* self);
