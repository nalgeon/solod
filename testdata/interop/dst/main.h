#pragma once
#include "so/builtin/builtin.h"
#include <stdint.h>
#include "person.ext.h"
#include "interop/src/sub/sub.h"

// -- Functions and methods --

// An extern type uchar comes from C header, so an exported function
// can name it even when it is unexported.
unsigned char main_FirstChar(unsigned char* buf);
