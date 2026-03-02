#include <stdio.h>
#include <stdint.h>
#include "so/builtin/builtin.h"

typedef struct {
    so_String name;
    int64_t balance;
    so_Slice flags;
} Account;

int64_t account_inc_balance(Account* a, int64_t amount);
