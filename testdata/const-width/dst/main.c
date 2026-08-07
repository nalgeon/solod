#include "main.h"

// -- Variables and constants --

// Constants whose value depends on the width of int. Go computes them at the
// host's width, so folding them into the C would freeze that width into the
// output. C reproduces them on its own, at the target's width.
static const so_unused so_int maxInt = (so_int)((uint64_t)(~(so_uint)(0)) >> 1);
static const so_unused so_uint maxUint = (so_uint)(~(so_uint)(0));
static const so_unused int64_t uintSize = ((int64_t)32 << ((uint64_t)(~(so_uint)(0)) >> 63));
static const so_unused int64_t ptrSize = ((int64_t)4 << ((uint64_t)(~(uintptr_t)(0)) >> 63));
static const so_unused uint64_t hiBits = ((uint64_t)0x8080808080808080u >> (64 - 8 * ptrSize));

// Constants C cannot reproduce, so it gets the computed value instead: an
// untyped intermediate above int64, and a shift count that reaches the width
// of the type it evaluates in.
static const so_unused int64_t untypedBig = 1024;
static const so_unused uint32_t wideShift = 0;

// Here only the intermediate is beyond C's reach. It folds on its own, and the
// width-dependent shift around it stays.
static const so_unused uint64_t wideMask = ((18446744073709551615u) >> (64 - 8 * ptrSize));

// -- Implementation --

int main(void) {
    // unsafe.Sizeof is also constant in Go, but it is emitted as C sizeof,
    // so these are the widths the C compiler actually chose.
    so_int intSize = 8 * (so_int)(unsafe_Sizeof((so_int)(0)));
    so_int realPtrSize = (so_int)(unsafe_Sizeof((uintptr_t)(0)));
    if (uintSize != intSize) {
        so_panic("uintSize");
    }
    if (ptrSize != realPtrSize) {
        so_panic("ptrSize");
    }
    if (maxUint != (so_uint)(maxInt) * 2 + 1) {
        so_panic("maxInt");
    }
    if ((maxUint >> (uintSize - 1)) != 1) {
        so_panic("maxUint");
    }
    // hiBits is the high bit of every byte of a word.
    uintptr_t want = 0;
    for (so_int i = 0; i < ptrSize; i++) {
        want |= ((uintptr_t)0x80 << (8 * i));
    }
    if ((uintptr_t)(hiBits) != want) {
        so_panic("hiBits");
    }
    // wideMask is every bit of a word set, and no more. Comparing as uint64
    // keeps a value too wide for the target from being truncated into agreeing.
    if ((uint64_t)(wideMask) != (uint64_t)(~(uintptr_t)(0))) {
        so_panic("wideMask");
    }
    if (untypedBig != 1024 || wideShift != 0) {
        so_panic("folded");
    }
    return 0;
}
