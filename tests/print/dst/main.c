#include "main.h"

typedef struct person {
    so_String name;
} person;

int main(void) {
    so_int a = 42;
    double b = 3.14;
    bool c = true;
    uint8_t d = U'x';
    so_String e = so_strlit("hello");
    person alice = {.name = so_strlit("alice")};
    person* f = &alice;
    so_println("%lld %f %d %u %s %p", a, b, c, d, e.ptr, f);
}
