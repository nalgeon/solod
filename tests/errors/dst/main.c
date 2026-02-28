#include "main.h"
so_Error main_ErrOutOfTea = so_error("no more tea available");

static so_Error makeTea(so_int arg) {
    if (arg == 42) {
        return main_ErrOutOfTea;
    }
    return NULL;
}

int main(void) {
    so_Error err = makeTea(7);
    if (err != NULL) {
        so_panic("err != nil");
    }
    err = makeTea(42);
    if (err == NULL) {
        so_panic("err == nil");
    }
    if (err != main_ErrOutOfTea) {
        so_panic("err != ErrOutOfTea");
    }
}
