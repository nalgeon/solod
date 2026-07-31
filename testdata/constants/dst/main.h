#pragma once
#include "so/builtin/builtin.h"
#include "sub/sub.h"

// -- Types --

typedef struct main_Point main_Point;

// Typedefed constant group.
typedef so_int main_HttpStatus;

// Regular constant group.
typedef so_String main_ServerState;

// Iota constant group.
typedef so_int main_Day;

typedef struct main_Point {
    so_int X;
    so_int Y;
} main_Point;

// -- Variables and constants --
static const main_HttpStatus main_StatusOK = 200;
static const main_HttpStatus main_StatusNotFound = 404;
static const main_HttpStatus main_StatusError = 500;
static const main_ServerState main_StateIdle = so_str("idle");
static const main_ServerState main_StateConnected = so_str("connected");
static const main_ServerState main_StateError = so_str("error");
static const main_Day main_Sunday = 0;
static const main_Day main_Monday = 1;
static const main_Day main_Tuesday = 2;

// Using constants in other definitions.
static const int64_t main_Zero = 42;
static const int64_t main_FortyTwo = main_Zero + 42;

// Untyped constants above math.MaxInt64 are declared as uint64.
static const uint64_t main_MaxUint64 = 18446744073709551615u;
extern main_Point main_PointZero;
extern main_Point main_PointSubZero;
