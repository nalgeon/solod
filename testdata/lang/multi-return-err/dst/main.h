#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_Reader {
    void* self;
    so_R_int_err (*Read)(void* self, so_int buf);
} main_Reader;

typedef struct main_File {
    so_int size;
} main_File;

typedef struct main_FileResult {
    main_File val;
    so_Error err;
} main_FileResult;

// -- Functions and methods --
so_R_int_err main_File_Read(void* self, so_int buf);
