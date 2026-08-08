#include "main.h"

// -- Implementation --

// An extern type uchar comes from C header, so an exported function
// can name it even when it is unexported.
unsigned char main_FirstChar(unsigned char* buf) {
    return *buf;
}

int main(void) {
    {
        // Passing values between So and C and vice versa.
        Account acc = (Account){.name = so_str("Alice"), .balance = 100, .flags = (so_Slice){(uint8_t[1]){42}, 1, 1}};
        int64_t balBefore = account_inc_balance(&acc, 50);
        so_println("%s %.*s %s %" PRId64 " %" PRId64 " %s %u", "name =", acc.name.len, acc.name.ptr, "balance =", balBefore, acc.balance, "flags[0] =", so_at(uint8_t, acc.flags, 0));
    }
    {
        // Calling variadic C functions from So.
        printf("One: %d\n", 1);
        printf("Two: %d, %d\n", 2, 3);
        printf("Three: %d, %d, %d\n", 4, 5, 6);
    }
    {
        // Extern nodecay functions.
        Account acc = {0};
        so_String name = so_str("Alice");
        account_set_name(&acc, name);
        if (so_string_ne(acc.name, so_str("Alice"))) {
            so_panic("Extern nodecay failed");
        }
    }
    {
        // Extern constants.
        if (INT64_MAX <= (int64_t)((int64_t)1 << 62)) {
            so_panic("maxInt64 <= 1<<62");
        }
    }
    {
        // Extern variadic function.
        Account acc = (Account){.name = so_str("Bob")};
        write_acc(&acc, "Hello %s!", "world");
    }
    {
        // Extern variadic method.
        Account acc = (Account){.name = so_str("Eve")};
        Account_Log(&acc, "Balance: %d", 789);
    }
    {
        // Extern function pointer.
        Account acc = (Account){.name = so_str("Charlie"), .write = write_acc};
        acc.write(&acc, "Balance: %d", 123);
    }
    {
        // Extern function pointer on a type alias.
        Account acc = (Account){.write = write_acc};
        Account target = (Account){.name = so_str("Diana")};
        acc.write(&target, "Balance: %d", 456);
    }
    {
        // Extern function pointer from a different package.
        Stream s = {0};
        s.Write = Discard;
        s.Write("Hello, %s!", "world");
    }
    {
        // Multi-word type names.
        so_byte b = 'a';
        unsigned char ch = (unsigned char)(b);
        if ((so_byte)(ch) != b) {
            so_panic("unexpected uchar value");
        }
        if (main_FirstChar(&ch) != ch) {
            so_panic("unexpected FirstChar value");
        }
    }
    return 0;
}
