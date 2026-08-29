#include "main.h"

// -- Variables and constants --

// File-level constants.
static const so_unused so_int fInt = 42;
static const so_unused so_String fString = so_str("file");
static const so_unused main_HttpStatus statusSecret = 999;
static const so_unused uint64_t halfUint64 = 9223372036854775808u;
static const so_unused uint64_t bigIota = 9223372036854775808u;
static const so_unused uint64_t bigIotaNext = 9223372036854775809u;

// An integer constant expression is folded when C cannot compute it
// step by step, and emitted as operators when it can.
static const so_unused uint64_t bigMask = 18446744073709551615u;
static const so_unused int64_t bigQuo = bigMask / 3;
static const so_unused int64_t smallMask = ((int64_t)1 << 20) - 1;

// An untyped constant is converted to the type of the use.
static const so_unused int64_t deBruijn = 0x077CB531;

// Float constant expressions are folded.
static const so_unused double ln10 = 2.30258509299404568401799145468436420760110148862877297603332790;
static const so_unused double log10E = 0.4342944819032518;

// An integral value still has to read as a float in C.
static const so_unused double area = 25.0;

// An integer literal in a float context: too large for a C integer type.
static const so_unused double bigFloat = 1e+23;

// Iota in float constants.
static const so_unused double step0 = 0.0;
static const so_unused double step1 = 0.25;
static const so_unused double step2 = 0.5;
main_Point main_PointZero = (main_Point){.X = main_Zero, .Y = main_Zero};
main_Point main_PointSubZero = (main_Point){.X = sub_Zero, .Y = sub_Zero};

// -- Implementation --

int main(void) {
    {
        // Local constants.
        const so_unused int64_t lInt = 500000000;
        const so_unused double lFloat = 6e+11;
        const so_unused so_String lString = so_str("local");
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
        // An intermediate value above int64 is folded away.
        uint64_t mask = 18446744073709551615u;
        if (mask != main_MaxUint64) {
            so_panic("1<<64 - 1");
        }
        uint64_t wide = 9223372036854775808u;
        if (wide != 9223372036854775808u) {
            so_panic("1<<100 >> 37");
        }
        // An intermediate above uint64 reached without a shift. C would wrap
        // it before the division brings it back into range.
        uint64_t over = 18446744073709551615u;
        if (over != main_MaxUint64) {
            so_panic("MaxUint64 * 3 / 3");
        }
        uint64_t sum = ((uint64_t)1 << 63) + 1;
        if (sum != 9223372036854775809u) {
            so_panic("1<<63 + 1");
        }
        int64_t neg = INT64_MIN;
        if (neg != INT64_MIN) {
            so_panic("-(1 << 63)");
        }
        uint64_t quo = (uint64_t)bigQuo;
        if (quo != 6148914691236517205) {
            so_panic("bigMask / 3");
        }
        uint64_t mixed = 18446744073709551614u;
        if (mixed != 18446744073709551614u) {
            so_panic("MaxUint64 + (-1)");
        }
        // Here the negative value never meets the one above int64, so C can
        // compute the expression itself.
        uint64_t apart = ((uint64_t)1 << 63) + (-1 + 2);
        if (apart != 9223372036854775809u) {
            so_panic("1<<63 + (-1 + 2)");
        }
        // The value above int64 is two levels down, and the small intermediate
        // value hides it. C still evaluates the subtraction in uint64.
        int16_t down = -1;
        if (down != -1) {
            so_panic("MaxUint64&0xffff - 1<<16");
        }
        int16_t halved = -1;
        if (halved != -1) {
            so_panic("(MaxUint64&0xfffe - 1<<16) / 2");
        }
        bool below = -1 < 0;
        if (!below) {
            so_panic("MaxUint64&0xffff-1<<16 < 0");
        }
        // A shift count of 64 is undefined in C, even with a value in range.
        uint64_t none = 0;
        if (none != 0) {
            so_panic("0 << 64");
        }
        int64_t gone = 0;
        if (gone != 0) {
            so_panic("1 >> 64");
        }
        // C does not define a shift of a negative value.
        int64_t low = INT64_MIN;
        if (low != INT64_MIN) {
            so_panic("-1 << 63");
        }
        int64_t ones = -1;
        if (ones != -1) {
            so_panic("-1 << 63 >> 63");
        }
        int32_t high = -256;
        if (high != -256) {
            so_panic("^0 << 8");
        }
    }
    {
        // C gives a constant expression the type of its literals, so the left
        // operand of a shift is converted to the type of the shift.
        int64_t product = ((int64_t)(2 * 1) << 61);
        if (product != 4611686018427387904) {
            so_panic("2 * 1 << 61");
        }
        int64_t sum = ((int64_t)(1 + 1) << 61);
        if (sum != 4611686018427387904) {
            so_panic("(1 + 1) << 61");
        }
        // A named constant is declared with a C type, so it needs no conversion.
        const so_unused int64_t one = 1;
        int64_t named = (one << 62);
        if (named != 4611686018427387904) {
            so_panic("one << 62");
        }
        // The left operand of a shift keeps its own type where that type holds
        // more than the type of the shift.
        uint64_t wide = (((uint64_t)1 << 63) >> 3);
        if (wide != 1152921504606846976) {
            so_panic("1<<63 >> 3");
        }
        // The left operand folds, so the shift around it needs the cast.
        int64_t folded = ((int64_t)1 << 40);
        if (folded != 1099511627776) {
            so_panic("(1<<64 >> 64) << 40");
        }
    }
    {
        // Expressions that C computes correctly are left as operators.
        int32_t narrow = (((int64_t)1 << 40) >> 20);
        if (narrow != 1048576) {
            so_panic("1<<40 >> 20");
        }
        int64_t small = smallMask + 1;
        if (small != 1048576) {
            so_panic("smallMask + 1");
        }
        int64_t flags = (((int64_t)1 << 20) | ((int64_t)1 << 10));
        if (flags != 1049600) {
            so_panic("1<<20 | 1<<10");
        }
    }
    {
        // An untyped constant takes the type of the use, so the product wraps
        // in uint32. The declared type of the constant is wider.
        uint32_t x = 0x40;
        if (((x & -x) * (uint32_t)deBruijn >> 27) != 27) {
            so_panic("(x&-x) * deBruijn >> 27");
        }
    }
    {
        // Float constant expressions are folded.
        // Compare through a variable, so that the comparison happens at
        // runtime on the emitted value instead of at compile time.
        double quo = log10E;
        if (quo != 0.4342944819032518) {
            so_panic("1 / ln10");
        }
        double whole = area;
        if (whole != 25.0) {
            so_panic("3*3 + 4*4");
        }
        double sum = 0.3;
        if (sum != 0.3) {
            so_panic("0.1 + 0.2");
        }
        double wide = 1e+200;
        if (wide != 1e200) {
            so_panic("1e200 * 1e200 / 1e200");
        }
        double tiny = 1e-300;
        if (tiny != 1e-300) {
            so_panic("1e-300 * 1e-300 * 1e300");
        }
        double lost = 1.0;
        if (lost != 1.0) {
            so_panic("(1e20 + 1) - 1e20");
        }
        double step = step2;
        if (step != 0.5) {
            so_panic("0.25 * iota");
        }
    }
    {
        // An integer literal in a float context.
        double big = 1e+22;
        if (big != 1e22) {
            so_panic("1e22");
        }
        double wider = bigFloat;
        if (wider != 1e23) {
            so_panic("bigFloat");
        }
        double hex = 255.0;
        if (hex != 255.0) {
            so_panic("0xFF");
        }
    }
    {
        // A float literal in an integer context. C reads 1e9 as a double and
        // promotes the other operand to match, which loses the low bits.
        int64_t sec = 1323195642;
        int64_t nsec = sec * 1000000000 + 80043132;
        if (nsec != 1323195642080043132) {
            so_panic("sec * 1e9");
        }
        so_int count = 3000000;
        if (count != 3000000) {
            so_panic("3 * 1e6");
        }
        int64_t neg = -1000000000;
        if (neg != -1000000000) {
            so_panic("-1e9");
        }
    }
    {
        // Same for constants narrowed to float32.
        float sum = 0.3f;
        if (sum != 0.3f) {
            so_panic("float32 0.1 + 0.2");
        }
        float wide = 100000.0f;
        if (wide != 100000.0f) {
            so_panic("float32 1e200 * 1e200 / 1e200 / 1e195");
        }
        float tiny = 1e-40f;
        if (tiny != 1e-40f) {
            so_panic("float32 1e-40 * 1e40 * 1e-40");
        }
        // A float32 literal compares as a float, not as a double.
        float lit = 0.1f;
        if (lit != 0.1f) {
            so_panic("float32 0.1");
        }
    }
    return 0;
}
