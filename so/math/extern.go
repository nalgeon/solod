// Package math provides basic constants and mathematical functions.
package math

import "math" // for testing

//so:include <float.h>
//so:include <limits.h>
//so:include <math.h>
//so:include <stdint.h>

// Floating-point limit values.
// Max is the largest finite value representable by the type.
// SmallestNonzero is the smallest positive, non-zero value representable by the type.

//so:extern FLT_MAX
const MaxFloat32 = math.MaxFloat32

//so:extern FLT_TRUE_MIN
const SmallestNonzeroFloat32 = math.SmallestNonzeroFloat32

//so:extern DBL_MAX
const MaxFloat64 = math.MaxFloat64

//so:extern DBL_TRUE_MIN
const SmallestNonzeroFloat64 = math.SmallestNonzeroFloat64

// Integer limit values.

//so:extern INT8_MAX
const MaxInt8 = math.MaxInt8

//so:extern INT8_MIN
const MinInt8 = math.MinInt8

//so:extern INT16_MAX
const MaxInt16 = math.MaxInt16

//so:extern INT16_MIN
const MinInt16 = math.MinInt16

//so:extern INT32_MAX
const MaxInt32 = math.MaxInt32

//so:extern INT32_MIN
const MinInt32 = math.MinInt32

//so:extern INT64_MAX
const MaxInt64 = math.MaxInt64

//so:extern INT64_MIN
const MinInt64 = math.MinInt64

//so:extern UINT8_MAX
const MaxUint8 = math.MaxUint8

//so:extern UINT16_MAX
const MaxUint16 = math.MaxUint16

//so:extern UINT32_MAX
const MaxUint32 = math.MaxUint32

//so:extern UINT64_MAX
const MaxUint64 = math.MaxUint64

// Basic operations.

//so:extern
func fabs(x float64) float64 { return math.Abs(x) }

//so:extern
func fmod(x, y float64) float64 { return math.Mod(x, y) }

//so:extern
func remainder(x, y float64) float64 { return math.Remainder(x, y) }

//so:extern
func fma(x, y, z float64) float64 { return math.FMA(x, y, z) }

//so:extern
func fmax(x, y float64) float64 { return math.Max(x, y) }

//so:extern
func fmin(x, y float64) float64 { return math.Min(x, y) }

//so:extern
func fdim(x, y float64) float64 { return math.Dim(x, y) }

// Exponential functions.

//so:extern
func exp(x float64) float64 { return math.Exp(x) }

//so:extern
func exp2(x float64) float64 { return math.Exp2(x) }

//so:extern
func expm1(x float64) float64 { return math.Expm1(x) }

//so:extern
func log(x float64) float64 { return math.Log(x) }

//so:extern
func log10(x float64) float64 { return math.Log10(x) }

//so:extern
func log2(x float64) float64 { return math.Log2(x) }

//so:extern
func log1p(x float64) float64 { return math.Log1p(x) }

// Power functions.

//so:extern
func pow(x, y float64) float64 { return math.Pow(x, y) }

//so:extern
func sqrt(x float64) float64 { return math.Sqrt(x) }

//so:extern
func cbrt(x float64) float64 { return math.Cbrt(x) }

//so:extern
func hypot(x, y float64) float64 { return math.Hypot(x, y) }

// Trigonometric functions.

//so:extern
func sin(x float64) float64 { return math.Sin(x) }

//so:extern
func cos(x float64) float64 { return math.Cos(x) }

//so:extern
func tan(x float64) float64 { return math.Tan(x) }

//so:extern
func asin(x float64) float64 { return math.Asin(x) }

//so:extern
func acos(x float64) float64 { return math.Acos(x) }

//so:extern
func atan(x float64) float64 { return math.Atan(x) }

//so:extern
func atan2(y, x float64) float64 { return math.Atan2(y, x) }

// Hyperbolic functions.

//so:extern
func sinh(x float64) float64 { return math.Sinh(x) }

//so:extern
func cosh(x float64) float64 { return math.Cosh(x) }

//so:extern
func tanh(x float64) float64 { return math.Tanh(x) }

//so:extern
func asinh(x float64) float64 { return math.Asinh(x) }

//so:extern
func acosh(x float64) float64 { return math.Acosh(x) }

//so:extern
func atanh(x float64) float64 { return math.Atanh(x) }

// Error and gamma functions.

//so:extern
func erf(x float64) float64 { return math.Erf(x) }

//so:extern
func erfc(x float64) float64 { return math.Erfc(x) }

//so:extern
func tgamma(x float64) float64 { return math.Gamma(x) }

//so:extern
func lgamma(x float64) float64 { f, _ := math.Lgamma(x); return f }

// Nearest integer floating-point operations.

//so:extern
func ceil(x float64) float64 { return math.Ceil(x) }

//so:extern
func floor(x float64) float64 { return math.Floor(x) }

//so:extern
func trunc(x float64) float64 { return math.Trunc(x) }

//so:extern
func round(x float64) float64 { return math.Round(x) }

// Floating-point manipulation functions.

//so:extern
func frexp(f float64, exp *int32) float64 {
	frac, e := math.Frexp(f)
	*exp = int32(e)
	return frac
}

//so:extern
func ldexp(frac float64, exp int32) float64 { return math.Ldexp(frac, int(exp)) }

//so:extern
func modf(f float64, intp *float64) float64 {
	ip, frac := math.Modf(f)
	*intp = ip
	return frac
}

//so:extern
func ilogb(x float64) int { return math.Ilogb(x) }

//so:extern
func logb(x float64) float64 { return math.Logb(x) }

//so:extern
func nextafterf(x, y float32) float32 { return math.Nextafter32(x, y) }

//so:extern
func nextafter(x, y float64) float64 { return math.Nextafter(x, y) }

//so:extern
func copysign(x, y float64) float64 { return math.Copysign(x, y) }

// Classification and comparison.

//so:extern
func signbit(x float64) bool { return math.Signbit(x) }
