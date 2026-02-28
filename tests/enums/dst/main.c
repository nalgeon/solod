#include "main.h"
const main_HttpStatus main_StatusOK = 200;
const main_HttpStatus main_StatusNotFound = 404;
const main_HttpStatus main_StatusError = 500;
const main_ServerState main_StateIdle = so_strlit("idle");
const main_ServerState main_StateConnected = so_strlit("connected");
const main_ServerState main_StateError = so_strlit("error");

int main(void) {
    main_HttpStatus status = main_StatusOK;
    so_println("%lld", status);
    main_ServerState state = main_StateConnected;
    if (so_string_ne(state, main_StateIdle)) {
        so_println("%d", true);
    }
}
