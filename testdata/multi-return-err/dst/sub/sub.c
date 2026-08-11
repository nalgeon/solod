#include "sub.h"

// -- Implementation --

sub_PointResult sub_MakePoint(so_int x, so_int y) {
    return (sub_PointResult){.val = (sub_Point){.X = x, .Y = y}, .err = (so_Error){}};
}
