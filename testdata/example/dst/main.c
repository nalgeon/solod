#include "main.h"

// -- Implementation --

so_int main_Person_Sleep(void* self) {
    main_Person* p = (main_Person*)self;
    p->Age += 1;
    return p->Age;
}

int main(void) {
    main_Person p = (main_Person){.Name = so_str("Alice"), .Age = 30};
    main_Person_Sleep(&p);
    so_println("%s %s %" PRId64 " %s", p.Name.ptr, "is now", p.Age, "years old.");
    p.Nums = so_make_slice(so_int, 0, 4);
    p.Nums = so_append(so_int, p.Nums, 7, 42, 13);
    so_println("%s %" PRId64, "1st lucky number is", so_at(so_int, p.Nums, 0));
}
