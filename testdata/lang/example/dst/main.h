#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_Person {
    so_String Name;
    so_int Age;
    so_int Nums[3];
} main_Person;

// -- Functions and methods --
so_int main_Person_Sleep(void* self);
