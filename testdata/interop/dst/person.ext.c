#include "person.ext.h"

void Account_Log(Account* a, const char* fmt, ...) {
    va_list args;
    va_start(args, fmt);
    printf("Log %s: ", a->name.ptr);
    vprintf(fmt, args);
    printf("\n");
    va_end(args);
}

int64_t account_inc_balance(Account* a, int64_t amount) {
    int64_t balBefore = a->balance;
    so_byte* flags = a->flags.ptr;
    printf("name = %s balance = %" PRId64 " flags[0] = %u\n",
           a->name.ptr, balBefore, a->flags.len > 0 ? flags[0] : 0);
    a->balance += amount;
    return balBefore;
}

void account_set_name(Account* a, so_String name) {
    a->name = name;
}

// vmeasure reads one argument per kind and sums the arguments. A nodecay call
// widens every scalar, so the reads use five types only.
static so_int vmeasure(so_String kinds, va_list args) {
    so_int total = 0;
    for (so_int i = 0; i < kinds.len; i++) {
        switch (kinds.ptr[i]) {
        case 'i':
            total += va_arg(args, so_int);
            break;
        case 'u':
            total += (so_int)va_arg(args, so_uint);
            break;
        case 'd':
            total += (so_int)va_arg(args, double);
            break;
        case 's':
            total += va_arg(args, so_String).len;
            break;
        case 'p':
            total += va_arg(args, void*) != NULL;
            break;
        }
    }
    return total;
}

so_int measure(so_String kinds, ...) {
    va_list args;
    va_start(args, kinds);
    so_int total = vmeasure(kinds, args);
    va_end(args);
    return total;
}

so_int Account_Measure(Account* a, so_String kinds, ...) {
    va_list args;
    va_start(args, kinds);
    so_int total = vmeasure(kinds, args) + a->balance;
    va_end(args);
    return total;
}

void write_acc(Account* a, const char* fmt, ...) {
    va_list args;
    va_start(args, fmt);
    printf("Account %s: ", a->name.ptr);
    vprintf(fmt, args);
    printf("\n");
    va_end(args);
}
