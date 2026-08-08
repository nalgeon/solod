package math_test

import (
	"solod.dev/so/math"
	"solod.dev/so/testing"
)

func TestAbs(t *testing.T) {
	if math.Abs(-2) != 2 {
		t.Error("Abs(-2) != 2")
	}
	if math.Abs(2) != 2 {
		t.Error("Abs(2) != 2")
	}
}

func TestAcos(t *testing.T) {
	if math.Acos(1) != 0 {
		t.Error("Acos(1) != 0")
	}
}

func TestAcosh(t *testing.T) {
	if math.Acosh(1) != 0 {
		t.Error("Acosh(1) != 0")
	}
}

func TestAsin(t *testing.T) {
	if math.Asin(0) != 0 {
		t.Error("Asin(0) != 0")
	}
}

func TestAsinh(t *testing.T) {
	if math.Asinh(0) != 0 {
		t.Error("Asinh(0) != 0")
	}
}

func TestAtan(t *testing.T) {
	if math.Atan(0) != 0 {
		t.Error("Atan(0) != 0")
	}
}

func TestAtan2(t *testing.T) {
	if math.Atan2(0, 0) != 0 {
		t.Error("Atan2(0, 0) != 0")
	}
}

func TestAtanh(t *testing.T) {
	if math.Atanh(0) != 0 {
		t.Error("Atanh(0) != 0")
	}
}

func TestCbrt(t *testing.T) {
	if math.Cbrt(8) != 2 {
		t.Error("Cbrt(8) != 2")
	}
	if math.Abs(math.Cbrt(27)-3) > 1e-10 {
		t.Error("Cbrt(27) != ~3")
	}
}

func TestCeil(t *testing.T) {
	if math.Ceil(1.49) != 2 {
		t.Error("Ceil(1.49) != 2")
	}
}

func TestConst(t *testing.T) {
	// Log2E and Log10E are constant expressions. Go computes them in arbitrary
	// precision and rounds once, so these are the values it arrives at.
	var log2E float64 = math.Log2E
	if log2E != 1.4426950408889634 {
		t.Error("Log2E != 1.4426950408889634")
	}
	var log10E float64 = math.Log10E
	if log10E != 0.4342944819032518 {
		t.Error("Log10E != 0.4342944819032518")
	}
}

func TestCopysign(t *testing.T) {
	if math.Copysign(3.2, -1) != -3.2 {
		t.Error("Copysign(3.2, -1) != -3.2")
	}
}

func TestCos(t *testing.T) {
	if math.Cos(0) != 1 {
		t.Error("Cos(0) != 1")
	}
	if math.Abs(math.Cos(math.Pi/2)) > 1e-10 {
		t.Error("Cos(Pi/2) != ~0")
	}
}

func TestCosh(t *testing.T) {
	if math.Cosh(0) != 1 {
		t.Error("Cosh(0) != 1")
	}
}

func TestDim(t *testing.T) {
	if math.Dim(4, -2) != 6 {
		t.Error("Dim(4, -2) != 6")
	}
	if math.Dim(-4, 2) != 0 {
		t.Error("Dim(-4, 2) != 0")
	}
}

func TestExp(t *testing.T) {
	if math.Abs(math.Exp(1)-2.7183) > 1e-4 {
		t.Error("Exp(1) != ~2.7183")
	}
	if math.Abs(math.Exp(2)-7.389) > 1e-3 {
		t.Error("Exp(2) != ~7.389")
	}
	if math.Abs(math.Exp(-1)-0.3679) > 1e-4 {
		t.Error("Exp(-1) != ~0.3679")
	}
}

func TestExp2(t *testing.T) {
	if math.Exp2(1) != 2 {
		t.Error("Exp2(1) != 2")
	}
	if math.Exp2(-3) != 0.125 {
		t.Error("Exp2(-3) != 0.125")
	}
}

func TestExpm1(t *testing.T) {
	if math.Abs(math.Expm1(0.01)-0.010050) > 1e-6 {
		t.Error("Expm1(0.01) != ~0.010050")
	}
	if math.Abs(math.Expm1(-1)-(-0.632121)) > 1e-6 {
		t.Error("Expm1(-1) != ~-0.632121")
	}
}

func TestFloor(t *testing.T) {
	if math.Floor(1.51) != 1 {
		t.Error("Floor(1.51) != 1")
	}
}

func TestLog(t *testing.T) {
	if math.Log(1) != 0 {
		t.Error("Log(1) != 0")
	}
	if math.Abs(math.Log(2.7183)-1.0) > 1e-4 {
		t.Error("Log(2.7183) != ~1.0")
	}
}

func TestLog2(t *testing.T) {
	if math.Log2(256) != 8 {
		t.Error("Log2(256) != 8")
	}
}

func TestLog10(t *testing.T) {
	if math.Log10(100) != 2 {
		t.Error("Log10(100) != 2")
	}
}

func TestMod(t *testing.T) {
	if math.Mod(7, 4) != 3 {
		t.Error("Mod(7, 4) != 3")
	}
}

func TestModf(t *testing.T) {
	i, f := math.Modf(3.14)
	if i != 3 {
		t.Error("Modf(3.14) int != 3")
	}
	if math.Abs(f-0.14) > 1e-10 {
		t.Error("Modf(3.14) frac != ~0.14")
	}
	i2, f2 := math.Modf(-2.71)
	if i2 != -2 {
		t.Error("Modf(-2.71) int != -2")
	}
	if math.Abs(f2-(-0.71)) > 1e-10 {
		t.Error("Modf(-2.71) frac != ~-0.71")
	}
}

func TestPow(t *testing.T) {
	if math.Pow(2, 3) != 8 {
		t.Error("Pow(2, 3) != 8")
	}
}

func TestPow10(t *testing.T) {
	if math.Pow10(2) != 100 {
		t.Error("Pow10(2) != 100")
	}
}

func TestRemainder(t *testing.T) {
	if math.Remainder(100, 30) != 10 {
		t.Error("Remainder(100, 30) != 10")
	}
}

func TestRound(t *testing.T) {
	if math.Round(10.5) != 11 {
		t.Error("Round(10.5) != 11")
	}
	if math.Round(-10.5) != -11 {
		t.Error("Round(-10.5) != -11")
	}
}

func TestRoundToEven(t *testing.T) {
	if math.RoundToEven(11.5) != 12 {
		t.Error("RoundToEven(11.5) != 12")
	}
	if math.RoundToEven(12.5) != 12 {
		t.Error("RoundToEven(12.5) != 12")
	}
}

func TestSin(t *testing.T) {
	if math.Sin(0) != 0 {
		t.Error("Sin(0) != 0")
	}
	if math.Abs(math.Sin(math.Pi)) > 1e-10 {
		t.Error("Sin(Pi) != ~0")
	}
}

func TestSinh(t *testing.T) {
	if math.Sinh(0) != 0 {
		t.Error("Sinh(0) != 0")
	}
}

func TestSqrt(t *testing.T) {
	if math.Sqrt(3*3+4*4) != 5 {
		t.Error("Sqrt(25) != 5")
	}
}

func TestTan(t *testing.T) {
	if math.Tan(0) != 0 {
		t.Error("Tan(0) != 0")
	}
}

func TestTanh(t *testing.T) {
	if math.Tanh(0) != 0 {
		t.Error("Tanh(0) != 0")
	}
}

func TestTrunc(t *testing.T) {
	if math.Trunc(math.Pi) != 3 {
		t.Error("Trunc(Pi) != 3")
	}
	if math.Trunc(-1.2345) != -1 {
		t.Error("Trunc(-1.2345) != -1")
	}
}
