#include "main.h"

// -- Forward declarations (types) --
typedef struct person person;

// -- Implementation --

typedef struct person {
    so_String name;
} person;

int main(void) {
    so_int vInt = 42;
    double vFloat = 3.14;
    bool vBool = true;
    uint8_t vByte = 'x';
    int32_t vRune = U'本';
    so_String vString = so_str("hello");
    person alice = (person){.name = so_str("alice")};
    person* vPtr = &alice;
    so_println("%" PRId64 " %f %d %u %d %s %p", vInt, vFloat, vBool, vByte, vRune, vString.ptr, vPtr);
    so_print("%s", "a");
    so_print("");
    so_print("%s", "b");
    so_println("");
}
