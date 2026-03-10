#pragma once
#include "so/builtin/builtin.h"

// -- Embeds --

#define newObj(T) (alloca(sizeof(T)))
#define freeObj(T, ptr) ((void)(ptr))

// -- Variables and constants --
