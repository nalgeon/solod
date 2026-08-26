#include "main.h"

// -- Types --

typedef struct packed packed;
typedef struct aligned aligned;

// Packed struct with so:attr.
//
typedef struct __attribute__((packed)) packed {
    so_byte a;
    so_int b;
} packed;

// Struct with multiple attrs.
//
typedef struct __attribute__((packed, aligned(16))) aligned {
    so_int x;
} aligned;

// Typedef alias with so:attr.
//
typedef __attribute__((aligned(8))) so_int myInt;

// -- Forward declarations --
static __attribute__((noinline)) void helper(void);

// -- Variables and constants --

// Unexported volatile variable.
//
static so_unused volatile so_int counter = 0;

// Exported volatile variable.
//
volatile so_int main_Counter = 0;

// Unexported thread-local variable.
//
static so_unused _Thread_local so_int perThread = 0;

// Exported thread-local variable.
//
_Thread_local so_int main_PerThread = 0;

// Combined volatile + thread-local.
//
static so_unused _Thread_local volatile so_int flags = 0;

// -- Implementation --

// Exported function with so:attr.
//
__attribute__((noinline)) void main_Work(void) {
}

// Unexported function with so:attr.
//
static __attribute__((noinline)) void helper(void) {
}

int main(void) {
    counter = 1;
    main_Counter = 2;
    perThread = 3;
    main_PerThread = 4;
    flags = 5;
    (void)(packed){.a = 1, .b = 2};
    (void)(aligned){.x = 3};
    (void)(main_Exported){.v = 4};
    myInt m = 5;
    (void)m;
    main_Work();
    helper();
    return 0;
}
