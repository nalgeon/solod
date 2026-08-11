#include "main.h"

// -- Forward declarations --
static so_int sum(so_Slice nums);

// -- Implementation --

void main_Sum_Add(void* self, so_Slice nums) {
    main_Sum* s = self;
    for (so_int _ = 0; _ < so_len(nums); _++) {
        so_int num = so_at(so_int, nums, _);
        s->v += num;
    }
}

static so_int sum(so_Slice nums) {
    so_int total = 0;
    for (so_int _ = 0; _ < so_len(nums); _++) {
        so_int num = so_at(so_int, nums, _);
        total += num;
    }
    return total;
}

int main(void) {
    {
        // Variadic function call.
        sum((so_Slice){(so_int[2]){1, 2}, 2, 2});
        so_int total = sum((so_Slice){(so_int[3]){1, 2, 3}, 3, 3});
        if (total != 6) {
            so_panic("wrong sum");
        }
        so_Slice nums = (so_Slice){(so_int[4]){1, 2, 3, 4}, 4, 4};
        total = sum(nums);
        if (total != 10) {
            so_panic("wrong sum");
        }
    }
    {
        // Variadic method call.
        main_Sum s = {};
        main_Sum_Add(&s, (so_Slice){(so_int[2]){1, 2}, 2, 2});
        main_Sum_Add(&s, (so_Slice){(so_int[3]){1, 2, 3}, 3, 3});
        if (s.v != 9) {
            so_panic("wrong sum");
        }
        so_Slice nums = (so_Slice){(so_int[4]){1, 2, 3, 4}, 4, 4};
        main_Sum_Add(&s, nums);
        if (s.v != 19) {
            so_panic("wrong sum");
        }
    }
    return 0;
}
