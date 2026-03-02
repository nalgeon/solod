#pragma once
#include "so/builtin/builtin.h"

// -- Types --

// Typedefed constant group.
typedef so_int main_HttpStatus;

// Regular constant group.
typedef so_String main_ServerState;

// -- Variables and constants --
extern const main_HttpStatus main_StatusOK;
extern const main_HttpStatus main_StatusNotFound;
extern const main_HttpStatus main_StatusError;
extern const main_ServerState main_StateIdle;
extern const main_ServerState main_StateConnected;
extern const main_ServerState main_StateError;
