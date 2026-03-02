#include "main.h"

// -- Implementation --

int main(void) {
    {
        so_Slice nums = so_make_slice(so_int, 3, 3);
        so_int n = so_index(so_int, nums, 1);
        so_index(so_int, nums, 1) = 42;
        so_int l = so_len(nums);
        so_int c = so_cap(nums);
        (void)n;
        (void)l;
        (void)c;
    }
    {
        so_Slice nums = so_make_slice(so_int, 0, 3);
        nums = so_append(so_int, nums, 1);
        nums = so_append(so_int, nums, 2, 3);
        so_int l = so_len(nums);
        so_int c = so_cap(nums);
        (void)l;
        (void)c;
    }
    {
        so_Slice nums = so_make_slice(so_int, 0, 8);
        so_Slice numsa = (so_Slice){(so_int[2]){1, 2}, 2, 2};
        nums = so_extend(so_int, nums, (numsa));
        nums = so_extend(so_int, nums, ((so_Slice){(so_int[2]){3, 4}, 2, 2}));
        so_int l = so_len(nums);
        if (l != 4) {
            so_panic("want l = 4");
        }
        if (so_index(so_int, nums, 3) != 4) {
            so_panic("want nums[3] = 4");
        }
    }
}
