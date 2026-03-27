package main

import "solod.dev/so/math"

func main() {
	{
		// Abs.
		x := math.Abs(-2)
		if x != 2 {
			panic("Abs(-2) != 2")
		}
		y := math.Abs(2)
		if y != 2 {
			panic("Abs(2) != 2")
		}
	}
	{
		// Acos.
		x := math.Acos(1)
		if x != 0 {
			panic("Acos(1) != 0")
		}
	}
	{
		// Acosh.
		x := math.Acosh(1)
		if x != 0 {
			panic("Acosh(1) != 0")
		}
	}
	{
		// Asin.
		x := math.Asin(0)
		if x != 0 {
			panic("Asin(0) != 0")
		}
	}
	{
		// Asinh.
		x := math.Asinh(0)
		if x != 0 {
			panic("Asinh(0) != 0")
		}
	}
	{
		// Atan.
		x := math.Atan(0)
		if x != 0 {
			panic("Atan(0) != 0")
		}
	}
	{
		// Atan2.
		x := math.Atan2(0, 0)
		if x != 0 {
			panic("Atan2(0, 0) != 0")
		}
	}
	{
		// Atanh.
		x := math.Atanh(0)
		if x != 0 {
			panic("Atanh(0) != 0")
		}
	}
	{
		// Cbrt.
		x := math.Cbrt(8)
		if x != 2 {
			panic("Cbrt(8) != 2")
		}
		y := math.Cbrt(27)
		if math.Abs(y-3) > 1e-10 {
			panic("Cbrt(27) != ~3")
		}
	}
	{
		// Ceil.
		x := math.Ceil(1.49)
		if x != 2 {
			panic("Ceil(1.49) != 2")
		}
	}
	{
		// Copysign.
		x := math.Copysign(3.2, -1)
		if x != -3.2 {
			panic("Copysign(3.2, -1) != -3.2")
		}
	}
	{
		// Cos.
		x := math.Cos(0)
		if x != 1 {
			panic("Cos(0) != 1")
		}
		y := math.Cos(math.Pi / 2)
		if math.Abs(y) > 1e-10 {
			panic("Cos(Pi/2) != ~0")
		}
	}
	{
		// Cosh.
		x := math.Cosh(0)
		if x != 1 {
			panic("Cosh(0) != 1")
		}
	}
	{
		// Dim.
		x := math.Dim(4, -2)
		if x != 6 {
			panic("Dim(4, -2) != 6")
		}
		y := math.Dim(-4, 2)
		if y != 0 {
			panic("Dim(-4, 2) != 0")
		}
	}
	{
		// Exp.
		x := math.Exp(1)
		if math.Abs(x-2.7183) > 1e-4 {
			panic("Exp(1) != ~2.7183")
		}
		y := math.Exp(2)
		if math.Abs(y-7.389) > 1e-3 {
			panic("Exp(2) != ~7.389")
		}
		z := math.Exp(-1)
		if math.Abs(z-0.3679) > 1e-4 {
			panic("Exp(-1) != ~0.3679")
		}
	}
	{
		// Exp2.
		x := math.Exp2(1)
		if x != 2 {
			panic("Exp2(1) != 2")
		}
		y := math.Exp2(-3)
		if y != 0.125 {
			panic("Exp2(-3) != 0.125")
		}
	}
	{
		// Expm1.
		x := math.Expm1(0.01)
		if math.Abs(x-0.010050) > 1e-6 {
			panic("Expm1(0.01) != ~0.010050")
		}
		y := math.Expm1(-1)
		if math.Abs(y-(-0.632121)) > 1e-6 {
			panic("Expm1(-1) != ~-0.632121")
		}
	}
	{
		// Floor.
		x := math.Floor(1.51)
		if x != 1 {
			panic("Floor(1.51) != 1")
		}
	}
	{
		// Log.
		x := math.Log(1)
		if x != 0 {
			panic("Log(1) != 0")
		}
		y := math.Log(2.7183)
		if math.Abs(y-1.0) > 1e-4 {
			panic("Log(2.7183) != ~1.0")
		}
	}
	{
		// Log2.
		x := math.Log2(256)
		if x != 8 {
			panic("Log2(256) != 8")
		}
	}
	{
		// Log10.
		x := math.Log10(100)
		if x != 2 {
			panic("Log10(100) != 2")
		}
	}
	{
		// Mod.
		x := math.Mod(7, 4)
		if x != 3 {
			panic("Mod(7, 4) != 3")
		}
	}
	{
		// Modf.
		i, f := math.Modf(3.14)
		if i != 3 {
			panic("Modf(3.14) int != 3")
		}
		if math.Abs(f-0.14) > 1e-10 {
			panic("Modf(3.14) frac != ~0.14")
		}
		i2, f2 := math.Modf(-2.71)
		if i2 != -2 {
			panic("Modf(-2.71) int != -2")
		}
		if math.Abs(f2-(-0.71)) > 1e-10 {
			panic("Modf(-2.71) frac != ~-0.71")
		}
	}
	{
		// Pow.
		x := math.Pow(2, 3)
		if x != 8 {
			panic("Pow(2, 3) != 8")
		}
	}
	{
		// Pow10.
		x := math.Pow10(2)
		if x != 100 {
			panic("Pow10(2) != 100")
		}
	}
	{
		// Remainder.
		x := math.Remainder(100, 30)
		if x != 10 {
			panic("Remainder(100, 30) != 10")
		}
	}
	{
		// Round.
		x := math.Round(10.5)
		if x != 11 {
			panic("Round(10.5) != 11")
		}
		y := math.Round(-10.5)
		if y != -11 {
			panic("Round(-10.5) != -11")
		}
	}
	{
		// RoundToEven.
		x := math.RoundToEven(11.5)
		if x != 12 {
			panic("RoundToEven(11.5) != 12")
		}
		y := math.RoundToEven(12.5)
		if y != 12 {
			panic("RoundToEven(12.5) != 12")
		}
	}
	{
		// Sin.
		x := math.Sin(0)
		if x != 0 {
			panic("Sin(0) != 0")
		}
		y := math.Sin(math.Pi)
		if math.Abs(y) > 1e-10 {
			panic("Sin(Pi) != ~0")
		}
	}
	{
		// Sinh.
		x := math.Sinh(0)
		if x != 0 {
			panic("Sinh(0) != 0")
		}
	}
	{
		// Sqrt.
		x := math.Sqrt(3*3 + 4*4)
		if x != 5 {
			panic("Sqrt(25) != 5")
		}
	}
	{
		// Tan.
		x := math.Tan(0)
		if x != 0 {
			panic("Tan(0) != 0")
		}
	}
	{
		// Tanh.
		x := math.Tanh(0)
		if x != 0 {
			panic("Tanh(0) != 0")
		}
	}
	{
		// Trunc.
		x := math.Trunc(math.Pi)
		if x != 3 {
			panic("Trunc(Pi) != 3")
		}
		y := math.Trunc(-1.2345)
		if y != -1 {
			panic("Trunc(-1.2345) != -1")
		}
	}
}
