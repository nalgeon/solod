#include "main.h"

// -- Implementation --

// File-level constants.
static const so_int fInt = 42;
static const so_String fString = so_strlit("file");
const main_HttpStatus main_StatusOK = 200;
const main_HttpStatus main_StatusNotFound = 404;
const main_HttpStatus main_StatusError = 500;
static const main_HttpStatus statusSecret = 999;
const main_ServerState main_StateIdle = so_strlit("idle");
const main_ServerState main_StateConnected = so_strlit("connected");
const main_ServerState main_StateError = so_strlit("error");

int main(void) {
    {
        const so_int lInt = 500000000;
        (void)lInt;
        const double lFloat = 3e20 / lInt;
        (void)lFloat;
        const so_String lString = so_strlit("local");
        (void)lString;
    }
    {
        main_HttpStatus status = main_StatusOK;
        (void)(status != main_StatusNotFound);
        main_HttpStatus secret = statusSecret;
        (void)(secret > main_StatusOK);
        main_ServerState state = main_StateConnected;
        (void)so_string_eq(state, main_StateIdle);
    }
    {
        (void)fInt;
        (void)fString;
    }
}
