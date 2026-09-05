#include "main.h"

// -- Types --

typedef struct rand_ rand_;

// Conflicts with rand from stdlib.h.
typedef struct rand_ {
    so_int free;
} rand_;

// -- Forward declarations --
static int64_t div_(int64_t d, int64_t r);

// -- Variables and constants --

// Conflicts with remove from stdio.h.
static so_unused so_int remove_ = 10;

// Conflicts with index from string.h.
static const so_unused int64_t index_ = 2;

// -- Implementation --

// Conflicts with div from stdlib.h.
static int64_t div_(int64_t d, int64_t r) {
    return so_div(d, r) + 1;
}

// An exported name gets the package prefix.
int64_t main_Abs(int64_t x) {
    if (x < 0) {
        return -x;
    }
    return x;
}

int main(void) {
    {
        // A block-scope shadows the file-scope one.
        so_int div = 12;
        so_int free = 3;
        (void)div;
        (void)free;
    }
    int64_t n = div_(12, 5);
    (void)n;
    (void)main_Abs(-1);
    rand_ w = (rand_){.free = remove_ + index_};
    (void)w.free;
    return 0;
}
