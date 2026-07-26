#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct sdl_CommonEvent sdl_CommonEvent;

// The c tag override must apply in packages that import this one.
typedef struct sdl_CommonEvent {
    uint32_t type;
    uint64_t Timestamp;
} sdl_CommonEvent;
