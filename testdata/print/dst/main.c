#include "main.h"

// -- Types --

typedef struct person person;

typedef struct person {
    so_String name;
} person;

// -- Implementation --

int main(void) {
    so_int vInt = 42;
    double vFloat = 3.14;
    bool vBool = true;
    so_byte vByte = 'x';
    so_rune vRune = 0x672c;
    so_String vString = so_str("hello");
    person alice = (person){.name = so_str("alice")};
    person* vPtr = &alice;
    so_println("%" PRIdINT " %f %s %u %d %.*s %p", vInt, vFloat, so_bool_cstr(vBool), vByte, vRune, vString.len, vString.ptr, vPtr);
    so_String* pString = &vString;
    so_println("%.*s", (*pString).len, (*pString).ptr);
    so_println("%s %s %s", so_bool_cstr(true), so_bool_cstr(!vBool), so_bool_cstr(vInt > 0));
    so_print("%s", "a");
    so_print("");
    so_print("%s", "b");
    so_println("");
    return 0;
}
