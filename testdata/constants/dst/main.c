#include "main.h"

// -- Variables and constants --

// File-level constants.
static const so_int fInt = 42;
static const so_String fString = so_str("file");
static const main_HttpStatus statusSecret = 999;
static const uint64_t halfUint64 = 9223372036854775808u;
static const uint64_t bigIota = 9223372036854775808u;
static const uint64_t bigIotaNext = 9223372036854775809u;
main_Point main_PointZero = (main_Point){.X = main_Zero, .Y = main_Zero};
main_Point main_PointSubZero = (main_Point){.X = sub_Zero, .Y = sub_Zero};

// -- Implementation --

int main(void) {
    {
        // Local constants.
        const int64_t lInt = 500000000;
        (void)lInt;
        const double lFloat = 3e20 / lInt;
        (void)lFloat;
        const so_String lString = so_str("local");
        (void)lString;
    }
    {
        // Using constants in expressions.
        main_HttpStatus status = main_StatusOK;
        (void)(status != main_StatusNotFound);
        main_HttpStatus secret = statusSecret;
        (void)(secret > main_StatusOK);
        main_ServerState state = main_StateConnected;
        (void)so_string_eq(state, main_StateIdle);
    }
    {
        // Using iota constants.
        main_Day day = main_Monday;
        (void)(day == main_Sunday);
    }
    {
        // Arithmetic on constants above math.MaxInt64 stays unsigned.
        uint64_t third = main_MaxUint64 / 3;
        if (third != 6148914691236517205) {
            so_panic("MaxUint64 / 3");
        }
        uint64_t shifted = (main_MaxUint64 >> 1);
        if (shifted != 9223372036854775807) {
            so_panic("MaxUint64 >> 1");
        }
        uint64_t half = halfUint64;
        if (half != 9223372036854775808u) {
            so_panic("halfUint64");
        }
        uint64_t first = bigIota;
        if (first != 9223372036854775808u) {
            so_panic("bigIota");
        }
        uint64_t next = bigIotaNext;
        if (next != 9223372036854775809u) {
            so_panic("bigIotaNext");
        }
    }
    {
        // Using _ on file level is not supported,
        // so silence the unused file-level constants here.
        (void)fInt;
        (void)fString;
    }
    return 0;
}
