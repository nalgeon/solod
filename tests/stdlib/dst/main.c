#include "main.h"
#include "so/so.h"
#include "so/errors/errors.h"
#include "so/math/math.h"
static so_ResInt work(so_int n);
so_Error main_Err42 = errors_New(so_strlit("42"));

static so_ResInt work(so_int n) {
    if (n == 42) {
        return {.Err = main_Err42};
    }
    return {.Val = 42};
}

int main(void) {
    double x = math_Sqrt(4.0);
    (void)x;
    so_ResInt r1 = work(11);
    if (r1.Err != NULL) {
        so_panic("unexpected error");
    }
    so_ResInt r2 = work(42);
    if (r2.Err != main_Err42) {
        so_panic("expected Err42");
    }
}
