#include "main.h"

// -- Implementation --

// File-level constants.
static const so_int fInt = 42;
static const so_String fString = so_str("file");
const main_HttpStatus main_StatusOK = 200;
const main_HttpStatus main_StatusNotFound = 404;
const main_HttpStatus main_StatusError = 500;
static const main_HttpStatus statusSecret = 999;
const main_ServerState main_StateIdle = so_str("idle");
const main_ServerState main_StateConnected = so_str("connected");
const main_ServerState main_StateError = so_str("error");
const main_Day main_Sunday = 0;
const main_Day main_Monday = 1;
const main_Day main_Tuesday = 2;

int main(void) {
    {
        // Local constants.
        const so_int lInt = 500000000;
        (void)lInt;
        const double lFloat = 3e20 / lInt;
        (void)lFloat;
        const so_String lString = so_str("local");
        (void)lString;
    }
    {
        // Using constants in expressions.
        main_HttpStatus status = main_StatusOK;
        (void)(status != main_StatusNotFound);
        main_HttpStatus secret = statusSecret;
        (void)(secret > main_StatusOK);
        main_ServerState state = main_StateConnected;
        (void)so_string_eq(state, main_StateIdle);
    }
    {
        // Using iota constants.
        main_Day day = main_Monday;
        (void)(day == main_Sunday);
    }
    {
        // Using _ on file level is not supported,
        // so silence the unused file-level constants here.
        (void)fInt;
        (void)fString;
    }
}
