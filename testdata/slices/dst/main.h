#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_Pair main_Pair;
typedef struct main_SliceHolder main_SliceHolder;

typedef struct main_Pair {
    so_int x;
    so_int y;
} main_Pair;
typedef so_Slice main_Bytes;

typedef struct main_SliceHolder {
    so_Slice nums;
} main_SliceHolder;
typedef so_Slice main_IntSlice;
