#pragma once
#include "so/builtin/builtin.h"

typedef struct main_Reader {
    void* self;
    so_Result (*Read)(void* self, so_int buf);
} main_Reader;

typedef struct main_File {
    so_int size;
} main_File;
so_Result main_File_Read(void* self, so_int buf);
