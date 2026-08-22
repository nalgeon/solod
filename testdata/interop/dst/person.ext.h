#include <stdarg.h>
#include <stdint.h>
#include "so/builtin/builtin.h"

#if !defined(so_build_hosted)
static inline int printf(const char* format, ...) {
    (void)format;
    return 0;
}

static inline int vprintf(const char* format, va_list args) {
    (void)format;
    (void)args;
    return 0;
}
#endif

typedef struct Account Account;

typedef void (*write_func_t)(Account* a, const char* fmt, ...);

struct Account {
    so_String name;
    int64_t balance;
    so_Slice flags;
    write_func_t write;
};

void Account_Log(Account* a, const char* fmt, ...);

int64_t account_inc_balance(Account* a, int64_t amount);

void account_set_name(Account* a, so_String name);

void write_acc(Account* a, const char* fmt, ...);

so_int measure(so_String kinds, ...);

so_int Account_Measure(Account* a, so_String kinds, ...);
