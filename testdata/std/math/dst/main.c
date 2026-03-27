#include "main.h"

// -- Implementation --

int main(void) {
    {
        // Abs.
        double x = math_Abs(-2);
        if (x != 2) {
            so_panic("Abs(-2) != 2");
        }
        double y = math_Abs(2);
        if (y != 2) {
            so_panic("Abs(2) != 2");
        }
    }
    {
        // Acos.
        double x = math_Acos(1);
        if (x != 0) {
            so_panic("Acos(1) != 0");
        }
    }
    {
        // Acosh.
        double x = math_Acosh(1);
        if (x != 0) {
            so_panic("Acosh(1) != 0");
        }
    }
    {
        // Asin.
        double x = math_Asin(0);
        if (x != 0) {
            so_panic("Asin(0) != 0");
        }
    }
    {
        // Asinh.
        double x = math_Asinh(0);
        if (x != 0) {
            so_panic("Asinh(0) != 0");
        }
    }
    {
        // Atan.
        double x = math_Atan(0);
        if (x != 0) {
            so_panic("Atan(0) != 0");
        }
    }
    {
        // Atan2.
        double x = math_Atan2(0, 0);
        if (x != 0) {
            so_panic("Atan2(0, 0) != 0");
        }
    }
    {
        // Atanh.
        double x = math_Atanh(0);
        if (x != 0) {
            so_panic("Atanh(0) != 0");
        }
    }
    {
        // Cbrt.
        double x = math_Cbrt(8);
        if (x != 2) {
            so_panic("Cbrt(8) != 2");
        }
        double y = math_Cbrt(27);
        if (math_Abs(y - 3) > 1e-10) {
            so_panic("Cbrt(27) != ~3");
        }
    }
    {
        // Ceil.
        double x = math_Ceil(1.49);
        if (x != 2) {
            so_panic("Ceil(1.49) != 2");
        }
    }
    {
        // Copysign.
        double x = math_Copysign(3.2, -1);
        if (x != -3.2) {
            so_panic("Copysign(3.2, -1) != -3.2");
        }
    }
    {
        // Cos.
        double x = math_Cos(0);
        if (x != 1) {
            so_panic("Cos(0) != 1");
        }
        double y = math_Cos(math_Pi / 2);
        if (math_Abs(y) > 1e-10) {
            so_panic("Cos(Pi/2) != ~0");
        }
    }
    {
        // Cosh.
        double x = math_Cosh(0);
        if (x != 1) {
            so_panic("Cosh(0) != 1");
        }
    }
    {
        // Dim.
        double x = math_Dim(4, -2);
        if (x != 6) {
            so_panic("Dim(4, -2) != 6");
        }
        double y = math_Dim(-4, 2);
        if (y != 0) {
            so_panic("Dim(-4, 2) != 0");
        }
    }
    {
        // Exp.
        double x = math_Exp(1);
        if (math_Abs(x - 2.7183) > 1e-4) {
            so_panic("Exp(1) != ~2.7183");
        }
        double y = math_Exp(2);
        if (math_Abs(y - 7.389) > 1e-3) {
            so_panic("Exp(2) != ~7.389");
        }
        double z = math_Exp(-1);
        if (math_Abs(z - 0.3679) > 1e-4) {
            so_panic("Exp(-1) != ~0.3679");
        }
    }
    {
        // Exp2.
        double x = math_Exp2(1);
        if (x != 2) {
            so_panic("Exp2(1) != 2");
        }
        double y = math_Exp2(-3);
        if (y != 0.125) {
            so_panic("Exp2(-3) != 0.125");
        }
    }
    {
        // Expm1.
        double x = math_Expm1(0.01);
        if (math_Abs(x - 0.010050) > 1e-6) {
            so_panic("Expm1(0.01) != ~0.010050");
        }
        double y = math_Expm1(-1);
        if (math_Abs(y - (-0.632121)) > 1e-6) {
            so_panic("Expm1(-1) != ~-0.632121");
        }
    }
    {
        // Floor.
        double x = math_Floor(1.51);
        if (x != 1) {
            so_panic("Floor(1.51) != 1");
        }
    }
    {
        // Log.
        double x = math_Log(1);
        if (x != 0) {
            so_panic("Log(1) != 0");
        }
        double y = math_Log(2.7183);
        if (math_Abs(y - 1.0) > 1e-4) {
            so_panic("Log(2.7183) != ~1.0");
        }
    }
    {
        // Log2.
        double x = math_Log2(256);
        if (x != 8) {
            so_panic("Log2(256) != 8");
        }
    }
    {
        // Log10.
        double x = math_Log10(100);
        if (x != 2) {
            so_panic("Log10(100) != 2");
        }
    }
    {
        // Mod.
        double x = math_Mod(7, 4);
        if (x != 3) {
            so_panic("Mod(7, 4) != 3");
        }
    }
    {
        // Modf.
        so_Result _res1 = math_Modf(3.14);
        double i = _res1.val.as_double;
        double f = _res1.val2.as_double;
        if (i != 3) {
            so_panic("Modf(3.14) int != 3");
        }
        if (math_Abs(f - 0.14) > 1e-10) {
            so_panic("Modf(3.14) frac != ~0.14");
        }
        so_Result _res2 = math_Modf(-2.71);
        double i2 = _res2.val.as_double;
        double f2 = _res2.val2.as_double;
        if (i2 != -2) {
            so_panic("Modf(-2.71) int != -2");
        }
        if (math_Abs(f2 - (-0.71)) > 1e-10) {
            so_panic("Modf(-2.71) frac != ~-0.71");
        }
    }
    {
        // Pow.
        double x = math_Pow(2, 3);
        if (x != 8) {
            so_panic("Pow(2, 3) != 8");
        }
    }
    {
        // Pow10.
        double x = math_Pow10(2);
        if (x != 100) {
            so_panic("Pow10(2) != 100");
        }
    }
    {
        // Remainder.
        double x = math_Remainder(100, 30);
        if (x != 10) {
            so_panic("Remainder(100, 30) != 10");
        }
    }
    {
        // Round.
        double x = math_Round(10.5);
        if (x != 11) {
            so_panic("Round(10.5) != 11");
        }
        double y = math_Round(-10.5);
        if (y != -11) {
            so_panic("Round(-10.5) != -11");
        }
    }
    {
        // RoundToEven.
        double x = math_RoundToEven(11.5);
        if (x != 12) {
            so_panic("RoundToEven(11.5) != 12");
        }
        double y = math_RoundToEven(12.5);
        if (y != 12) {
            so_panic("RoundToEven(12.5) != 12");
        }
    }
    {
        // Sin.
        double x = math_Sin(0);
        if (x != 0) {
            so_panic("Sin(0) != 0");
        }
        double y = math_Sin(math_Pi);
        if (math_Abs(y) > 1e-10) {
            so_panic("Sin(Pi) != ~0");
        }
    }
    {
        // Sinh.
        double x = math_Sinh(0);
        if (x != 0) {
            so_panic("Sinh(0) != 0");
        }
    }
    {
        // Sqrt.
        double x = math_Sqrt(3 * 3 + 4 * 4);
        if (x != 5) {
            so_panic("Sqrt(25) != 5");
        }
    }
    {
        // Tan.
        double x = math_Tan(0);
        if (x != 0) {
            so_panic("Tan(0) != 0");
        }
    }
    {
        // Tanh.
        double x = math_Tanh(0);
        if (x != 0) {
            so_panic("Tanh(0) != 0");
        }
    }
    {
        // Trunc.
        double x = math_Trunc(math_Pi);
        if (x != 3) {
            so_panic("Trunc(Pi) != 3");
        }
        double y = math_Trunc(-1.2345);
        if (y != -1) {
            so_panic("Trunc(-1.2345) != -1");
        }
    }
}
