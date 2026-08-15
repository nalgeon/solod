//go:build ignore

// A hosted environment declares every libm function in <math.h>. A freestanding
// environment has no libm, so this file declares each libm function the package
// calls, and a call panics.
//
// The C compiler needs a declaration for each one, even in a program that calls
// none of them: the emitted math.c holds a call to every function the package
// declares with so:extern.
//
// Abs, Copysign, Dim, Inf, IsInf, IsNaN, Max, Min, NaN, Pow10, RoundToEven,
// Signbit and the Float64bits family need no libm, so they work in a
// freestanding environment.

#ifdef so_build_hosted

#include <math.h>

#else

// clang-format off

#define so_math_stub_1(name)                                       \
    static inline double name(double x) {                          \
        (void)x;                                                   \
        so_panic("math: " #name " requires a hosted environment"); \
        return 0;                                                  \
    }

#define so_math_stub_2(name)                                       \
    static inline double name(double x, double y) {                \
        (void)x;                                                   \
        (void)y;                                                   \
        so_panic("math: " #name " requires a hosted environment"); \
        return 0;                                                  \
    }

#define so_math_stub_3(name)                                       \
    static inline double name(double x, double y, double z) {      \
        (void)x;                                                   \
        (void)y;                                                   \
        (void)z;                                                   \
        so_panic("math: " #name " requires a hosted environment"); \
        return 0;                                                  \
    }

// Basic operations.
so_math_stub_2(fmod)
so_math_stub_2(remainder)
so_math_stub_3(fma)

// Exponential functions.
so_math_stub_1(exp)
so_math_stub_1(exp2)
so_math_stub_1(expm1)
so_math_stub_1(log)
so_math_stub_1(log10)
so_math_stub_1(log2)
so_math_stub_1(log1p)

// Power functions.
so_math_stub_2(pow)
so_math_stub_1(sqrt)
so_math_stub_1(cbrt)
so_math_stub_2(hypot)

// Trigonometric functions.
so_math_stub_1(sin)
so_math_stub_1(cos)
so_math_stub_1(tan)
so_math_stub_1(asin)
so_math_stub_1(acos)
so_math_stub_1(atan)
so_math_stub_2(atan2)

// Hyperbolic functions.
so_math_stub_1(sinh)
so_math_stub_1(cosh)
so_math_stub_1(tanh)
so_math_stub_1(asinh)
so_math_stub_1(acosh)
so_math_stub_1(atanh)

// Error and gamma functions.
so_math_stub_1(erf)
so_math_stub_1(erfc)
so_math_stub_1(tgamma)
so_math_stub_1(lgamma)

// Nearest integer floating-point operations.
so_math_stub_1(ceil)
so_math_stub_1(floor)
so_math_stub_1(trunc)
so_math_stub_1(round)

// Floating-point manipulation functions.
so_math_stub_1(logb)
so_math_stub_2(nextafter)

static inline double frexp(double x, int* exp) {
    (void)x;
    (void)exp;
    so_panic("math: frexp requires a hosted environment");
    return 0;
}

// clang-format on

static inline double ldexp(double frac, int exp) {
    (void)frac;
    (void)exp;
    so_panic("math: ldexp requires a hosted environment");
    return 0;
}

static inline double modf(double x, double* intp) {
    (void)x;
    (void)intp;
    so_panic("math: modf requires a hosted environment");
    return 0;
}

static inline int ilogb(double x) {
    (void)x;
    so_panic("math: ilogb requires a hosted environment");
    return 0;
}

static inline float nextafterf(float x, float y) {
    (void)x;
    (void)y;
    so_panic("math: nextafterf requires a hosted environment");
    return 0;
}

#endif  // so_build_hosted
