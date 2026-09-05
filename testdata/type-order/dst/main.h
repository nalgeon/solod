#pragma once
#include "so/builtin/builtin.h"

// -- Types --

typedef struct main_Node main_Node;
typedef struct main_Employee main_Employee;
typedef struct main_Pet main_Pet;
typedef struct main_Point main_Point;
typedef struct main_Rect main_Rect;
typedef struct main_Cell main_Cell;
typedef struct main_Grid main_Grid;
typedef struct main_Origin main_Origin;
typedef struct main_Anchor main_Anchor;
typedef struct main_Base main_Base;
typedef struct main_Loop main_Loop;
typedef struct main_Outer main_Outer;
typedef struct main_Payload main_Payload;
typedef struct main_Reading main_Reading;
typedef main_Origin main_Target;
typedef main_Base main_Link;
typedef main_Link main_Chain;
typedef main_Loop main_Twin;

// Self-referencing struct type.
typedef struct main_Node {
    so_int value;
    main_Node* next;
} main_Node;

// Type referencing another type defined later.
typedef struct main_Employee {
    so_String name;
    main_Pet* pet;
} main_Employee;

typedef struct main_Pet {
    so_String name;
} main_Pet;

typedef struct main_Point {
    so_int X;
    so_int Y;
} main_Point;

// Type using a type defined later by value.
typedef struct main_Rect {
    main_Point Min;
    main_Point Max;
} main_Rect;

typedef struct main_Cell {
    so_int v;
} main_Cell;

// Array of a type defined later.
typedef struct main_Grid {
    main_Cell cells[4];
} main_Grid;

typedef struct main_Origin {
    so_int v;
} main_Origin;

// Named type of a type defined later.
typedef main_Origin main_Target;

// Pointer to a named struct type defined later.
typedef struct main_Anchor {
    main_Chain* link;
} main_Anchor;

typedef struct main_Base {
    so_int v;
} main_Base;
typedef main_Base main_Link;
typedef main_Link main_Chain;

// Named struct type in the signature of a func type and an interface.
typedef void (*main_Visit)(main_Chain*);

typedef struct main_Visitor {
    void* self;
    void (*See)(void* self, main_Chain*);
} main_Visitor;

static inline void main_Visitor_See(main_Visitor self, main_Chain* _0) {
    self.See(self.self, _0);
}

// Struct that points to a named type of itself.
typedef struct main_Loop {
    main_Twin* twin;
} main_Loop;
typedef main_Loop main_Twin;

typedef main_Payload (*main_Handler)(main_Payload);

// Func type held by value: Handler must precede Outer, but the struct types
// in its signature are fine as forward declarations.
typedef struct main_Outer {
    main_Handler handle;
} main_Outer;

typedef struct main_Payload {
    so_int v;
} main_Payload;
typedef so_int main_Meters;

// Pointer to a non-struct type: Meters has no forward declaration,
// so its definition must come first.
typedef struct main_Reading {
    main_Meters* depth;
} main_Reading;
