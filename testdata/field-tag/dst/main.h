#pragma once
#include "so/builtin/builtin.h"
#include "sdl/sdl.h"

// -- Types --

typedef struct main_Event main_Event;
typedef struct main_Box main_Box;

// Regular struct.
typedef struct main_Event {
    uint32_t type;
    int32_t data;
} main_Event;

// Generic struct.
typedef struct main_Box {
    uint32_t ident;
} main_Box;
