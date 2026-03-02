#include "main.h"

// -- Forward declarations (functions and methods) --
static so_int sum(so_Slice nums);

// -- Implementation --

static so_int sum(so_Slice nums) {
    so_int total = 0;
    for (so_int _ = 0; _ < so_len(nums); _++) {
        so_int num = so_index(so_int, nums, _);
        total += num;
    }
    return total;
}

int main(void) {
    sum((so_Slice){(so_int[2]){1, 2}, 2, 2});
    sum((so_Slice){(so_int[3]){1, 2, 3}, 3, 3});
    so_Slice nums = (so_Slice){(so_int[4]){1, 2, 3, 4}, 4, 4};
    sum(nums);
}
