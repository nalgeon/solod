#pragma once
#include "so/builtin/builtin.h"
#include <stdint.h>
#include "person.ext.h"
#include "interop/src/sub/sub.h"

// -- Variables and constants --

// MaxHalf references an extern constant, which comes from a C header
// and needs no declaration of its own.
static const int64_t main_MaxHalf = INT64_MAX / 2;

// -- Functions and methods --

// An extern type uchar comes from C header, so an exported function
// can reference it even when it is unexported.
unsigned char main_FirstChar(unsigned char* buf);
