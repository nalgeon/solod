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
        Account acc = {};
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
        // Extern nodecay variadic function: the args go flat,
        // at their So types, and every scalar widens.
        so_String name = so_str("Alice");
        int32_t i32 = 7;
        so_uint u = 4;
        uint8_t u8 = 3;
        float f32 = 1.5f;
        Account acc = {};
        so_int got = measure(so_str("ssiiiiiuudp"), name, so_str("Bob"), (so_int)(10), (so_int)(-8), (so_int)(i32), (so_int)(true), (so_int)('A'), (so_uint)(u), (so_uint)(u8), (double)(f32), &acc);
        so_int want = so_len(name) + so_len(so_str("Bob")) + 10 - 8 + (so_int)(i32) + 1 + (so_int)('A') + (so_int)(u) + (so_int)(u8) + (so_int)(f32) + 1;
        if (got != want) {
            so_panic("measure failed");
        }
        if (measure(so_str("")) != 0) {
            so_panic("empty measure failed");
        }
    }
    {
        // Extern nodecay variadic method.
        Account acc = (Account){.balance = 20};
        if (Account_Measure(&acc, so_str("is"), (so_int)(5), so_str("abc")) != 28) {
            so_panic("Measure failed");
        }
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
        Stream s = {};
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
