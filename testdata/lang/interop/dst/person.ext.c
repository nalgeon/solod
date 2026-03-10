#include <inttypes.h>
#include "person.ext.h"

int64_t account_inc_balance(Account* a, int64_t amount) {
    int64_t balBefore = a->balance;
    so_byte* flags = a->flags.ptr;
    printf("name = %s balance = %" PRId64 " flags[0] = %u\n",
           a->name.ptr, balBefore, a->flags.len > 0 ? flags[0] : 0);
    a->balance += amount;
    return balBefore;
}
