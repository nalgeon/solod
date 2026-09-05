#include "main.h"

// -- Types --

typedef struct movie movie;

// A function pointer field with a conflicting parameter name.
typedef struct movie {
    so_int (*rate)(so_int);
} movie;

// An interface method with a conflicting parameter name.
typedef struct rater {
    void* self;
    so_int (*rate)(void* self, so_int);
} rater;

static inline so_unused so_int rater_rate(rater self, so_int register_) {
    return self.rate(self.self, register_);
}

// -- Forward declarations --
static so_int typeof_(void);
static so_int scale(so_int long_, so_int register_);
static so_int shadow(so_int long_);
static so_int switchTemp(so_int x);
static so_R_int_int pair(void);
static so_int resultTemp(void);
static so_int varDecl(void);

// -- Variables and constants --

// An exported identifier gets a package prefix.
so_int main_NULL = 0;

// A conflicting unexported package-level identifier.
static so_unused so_int stderr_ = 1;

// -- Implementation --

static so_int typeof_(void) {
    return stderr_;
}

// Conflicting C keywords used as parameter names.
static so_int scale(so_int long_, so_int register_) {
    so_int total = long_ * register_;
    return total;
}

// A mangled parameter (long -> long_) and a same-named local in a nested
// block are a legal C shadow, not a conflict, so both are accepted.
static so_int shadow(so_int long_) {
    if (long_ > 0) {
        so_int long_ = 99;
        return long_;
    }
    return long_;
}

// A name that looks like a generated temporary is not reserved: the switch
// tag temporary picks the next free name instead of shadowing the variable.
static so_int switchTemp(so_int x) {
    so_int _sw1 = 99;
    {
        so_int _sw2 = x;
        if (_sw2 == 1) {
            return _sw1;
        }
    }
    return 0;
}

static so_R_int_int pair(void) {
    return (so_R_int_int){.val = 1, .val2 = 2};
}

// The same for the temporary that holds a multi-value call result.
static so_int resultTemp(void) {
    so_int _res1 = 99;
    so_R_int_int _res2 = pair();
    so_int a = _res2.val;
    so_int b = _res2.val2;
    return _res1 + a + b;
}

// Conflicting local function variables.
static so_int varDecl(void) {
    so_int double_ = 5;
    so_int union_ = 1, enum_ = 2;
    return double_ + union_ + enum_;
}

int main(void) {
    // Conflicting local variables.
    so_int long_ = 10;
    so_int short_ = 20;
    so_int value = scale(long_, short_);
    (void)value;
    (void)shadow(value);
    (void)switchTemp(1);
    (void)typeof_();
    (void)resultTemp();
    (void)varDecl();
    // The name is mangled everywhere it is used.
    for (so_int bool_ = 0; bool_ < long_; bool_++) {
        so_int b = bool_;
        (void)b;
    }
    movie m = {};
    rater r = {};
    (void)m;
    (void)r;
    return 0;
}
