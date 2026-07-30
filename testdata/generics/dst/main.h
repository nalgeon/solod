#pragma once
#include "so/builtin/builtin.h"

// -- Embeds --

#include "so/builtin/builtin.h"

#define newObj(T) (alloca(sizeof(T)))
#define freeObj(T, ptr) ((void)(ptr))
#define newMap(K, V, size) ((main_Map){.len = (size)})
#define main_Map_Len(K, V, m) ((m)->len)

typedef struct {
    int len;
} main_Map;

// -- Types --

typedef struct main_Stringer {
    void* self;
    so_String (*String)(void* self);
} main_Stringer;

// -- Functions and methods --

#define add(T, a_, b_) ({ \
    a_ + b_; \
})

#define first(T, a_, b_) ({ \
    a_; \
})

#define same(T, v_) ({ \
    v_; \
})
