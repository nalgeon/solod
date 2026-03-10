#include "main.h"

// -- Implementation --

int main(void) {
    double pi = math_Pi;
    (void)pi;
    double x = math_Sqrt(16.0);
    (void)x;
    double y = math_Pow(2.0, 10.0);
    (void)y;
    double z = math_Abs(-3.14);
    (void)z;
    double f = math_Floor(2.7);
    (void)f;
    double c = math_Ceil(2.3);
    (void)c;
    double r = math_Round(2.5);
    (void)r;
    double s = math_Sin(math_Pi);
    (void)s;
    double a = math_Atan2(1.0, 1.0);
    (void)a;
    double m = math_Fmin(3.0, 5.0);
    (void)m;
    double lg = math_Log(math_E);
    (void)lg;
    double fm = math_Fmod(5.5, 2.0);
    (void)fm;
}
