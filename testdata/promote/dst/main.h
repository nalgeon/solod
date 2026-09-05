#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_counter main_counter;
typedef struct main_Stats main_Stats;
typedef main_counter main_alias;

// counter is unexported, but so:promote emits it in the header
// so the Stats can reference it.
//
typedef struct main_counter {
    so_int val;
} main_counter;

// Alias renames an so:promote type.
//
typedef main_counter main_alias;

// Stats is exported and has a field of the unexported so:promote type.
// Its constructor and method are emitted in the header because of so:inline.
typedef struct main_Stats {
    main_counter c;
} main_Stats;

// -- Variables and constants --
extern so_int main_defaultCap;
static const int64_t main_version = 3;

// -- Functions and methods --

// newCounter is called from an inline (header) NewStats,
// so it needs to be promoted to the header.
//
main_counter main_newCounter(void);

// inc is called from an inline (header) Stats.Inc method,
// so it needs to be promoted to the header.
//
void main_counter_inc(void* self);

static inline main_Stats main_NewStats(void) {
    return (main_Stats){.c = main_newCounter()};
}

static inline void main_Stats_Inc(void* self) {
    main_Stats* w = self;
    main_counter_inc(&w->c);
}

// GetCounter is a non-inline function that returns a promoted type.
main_counter main_GetCounter(void);
