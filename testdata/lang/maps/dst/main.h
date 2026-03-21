#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_Pair {
    so_int x;
    so_int y;
} main_Pair;

typedef so_int (*main_IntFunc)();
typedef so_Map* main_StrMap;

typedef struct main_MapHolder {
    so_Map* m;
} main_MapHolder;
