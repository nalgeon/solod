#include "main.h"

// -- Implementation --

int main(void) {
    {
        // make, len, cap.
        so_Slice nums = so_make_slice(so_int, 3, 3);
        // 0
        so_int n = so_at(so_int, nums, 1);
        so_at(so_int, nums, 1) = 42;
        // 3
        so_int l = so_len(nums);
        so_int c = so_cap(nums);
        (void)n;
        (void)l;
        (void)c;
    }
    {
        // Append values.
        so_Slice nums = so_make_slice(so_int, 0, 3);
        nums = so_append(so_int, nums, 1);
        nums = so_append(so_int, nums, 2, 3);
        // 3
        so_int l = so_len(nums);
        // 3
        so_int c = so_cap(nums);
        (void)l;
        (void)c;
    }
    // Resizing slices beyond their initial capacity with append() panics.
    // nums = append(nums, 4)
    {
        // Append slices.
        so_Slice nums = so_make_slice(so_int, 0, 8);
        so_Slice numsa = (so_Slice){(so_int[2]){1, 2}, 2, 2};
        nums = so_extend(so_int, nums, (numsa));
        nums = so_extend(so_int, nums, ((so_Slice){(so_int[2]){3, 4}, 2, 2}));
        // 4
        so_int l = so_len(nums);
        if (l != 4) {
            so_panic("want l = 4");
        }
        if (so_at(so_int, nums, 3) != 4) {
            so_panic("want nums[3] = 4");
        }
    }
}
