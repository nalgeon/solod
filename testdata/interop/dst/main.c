#include "main.h"
#include <stdio.h>
#include "person.ext.h"

// -- Implementation --

int main(void) {
    {
        // Passing values between So and C and vice versa.
        Account acc = (Account){.name = so_str("Alice"), .balance = 100, .flags = (so_Slice){(uint8_t[1]){42}, 1, 1}};
        int64_t balBefore = account_inc_balance(&acc, 50);
        so_println("%s %s %s %" PRId64 " %" PRId64 " %s %u", "name =", acc.name.ptr, "balance =", balBefore, acc.balance, "flags[0] =", so_at(uint8_t, acc.flags, 0));
    }
    {
        // Calling variadic C functions from So.
        printf("One: %d\n", 1);
        printf("Two: %d, %d\n", 2, 3);
        printf("Three: %d, %d, %d\n", 4, 5, 6);
    }
}
