#include "main.h"

// -- Types --

typedef struct point point;

typedef struct point {
    so_int x;
    so_int y;
} point;

// -- Result types --

typedef struct pointResult {
    point val;
    so_Error err;
} pointResult;

// -- Forward declarations --
static main_FileResult makeFile(so_int size);
static pointResult makePoint(so_int x, so_int y);
static sub_PointResult makeSubPoint(so_int x, so_int y);
static so_R_int_err divide(so_int a, so_int b);
static so_R_int_err returnInt(void);
static so_R_rune_err returnRune(void);
static so_R_str_err returnString(void);
static so_R_slice_err returnSlice(so_Slice s);
static main_FileResult returnStruct(void);
static so_R_ptr_err returnAny(void);
static so_R_ptr_err returnPtr(void);
static so_R_slice_err returnNilSlice(void);
static so_R_ptr_err returnNilPtr(void);
static so_R_ptr_err returnAnyInt(void* v);
static so_R_ptr_bool returnAnyPair(void* v);
static so_R_int_err forwardInt(void);
static so_R_rune_err forwardRune(void);
static so_R_str_err forwardString(void);
static so_R_slice_err forwardSlice(so_Slice s);
static main_FileResult forwardStruct(void);
static so_R_ptr_err forwardAny(void);
static so_R_ptr_err forwardPtr(void);
static void testBasic(void);
static void testIf(void);
static void testReturnTypes(void);
static void testNilValues(void);
static void testForwarding(void);
static void testStructExported(void);
static void testStructUnexported(void);
static void testStructOtherPackage(void);

// -- Variables and constants --
static so_unused main_File file = {};

// -- Implementation --

static main_FileResult makeFile(so_int size) {
    return (main_FileResult){.val = (main_File){.size = size}, .err = (so_Error){}};
}

so_R_int_err main_File_Read(void* self, so_int buf) {
    main_File* f = self;
    (void)buf;
    return (so_R_int_err){.val = f->size, .err = (so_Error){}};
}

static pointResult makePoint(so_int x, so_int y) {
    return (pointResult){.val = (point){.x = x, .y = y}, .err = (so_Error){}};
}

static sub_PointResult makeSubPoint(so_int x, so_int y) {
    return (sub_PointResult){.val = (sub_Point){.X = x, .Y = y}, .err = (so_Error){}};
}

static so_R_int_err divide(so_int a, so_int b) {
    return (so_R_int_err){.val = so_div(a, b), .err = (so_Error){}};
}

static so_R_int_err returnInt(void) {
    return (so_R_int_err){.val = 42, .err = (so_Error){}};
}

static so_R_rune_err returnRune(void) {
    return (so_R_rune_err){.val = 'x', .err = (so_Error){}};
}

static so_R_str_err returnString(void) {
    return (so_R_str_err){.val = so_str("hello"), .err = (so_Error){}};
}

static so_R_slice_err returnSlice(so_Slice s) {
    return (so_R_slice_err){.val = s, .err = (so_Error){}};
}

static main_FileResult returnStruct(void) {
    return (main_FileResult){.val = (main_File){.size = 42}, .err = (so_Error){}};
}

static so_R_ptr_err returnAny(void) {
    return (so_R_ptr_err){.val = &file, .err = (so_Error){}};
}

static so_R_ptr_err returnPtr(void) {
    return (so_R_ptr_err){.val = &file, .err = (so_Error){}};
}

// Nil and untyped values need the same conversion as a single return.
static so_R_slice_err returnNilSlice(void) {
    return (so_R_slice_err){.val = (so_Slice){}, .err = (so_Error){}};
}

static so_R_ptr_err returnNilPtr(void) {
    return (so_R_ptr_err){.val = NULL, .err = (so_Error){}};
}

static so_R_ptr_err returnAnyInt(void* v) {
    return (so_R_ptr_err){.val = v, .err = (so_Error){}};
}

static so_R_ptr_bool returnAnyPair(void* v) {
    return (so_R_ptr_bool){.val = v, .val2 = true};
}

// func returnIface() (Reader, error)  { return &file, nil }
static so_R_int_err forwardInt(void) {
    return returnInt();
}

static so_R_rune_err forwardRune(void) {
    return returnRune();
}

static so_R_str_err forwardString(void) {
    return returnString();
}

static so_R_slice_err forwardSlice(so_Slice s) {
    return returnSlice(s);
}

static main_FileResult forwardStruct(void) {
    return returnStruct();
}

static so_R_ptr_err forwardAny(void) {
    return returnAny();
}

static so_R_ptr_err forwardPtr(void) {
    return returnPtr();
}

static void testBasic(void) {
    // Destructure into new variables.
    so_R_int_err _res1 = divide(10, 3);
    so_int q = _res1.val;
    so_Error err = _res1.err;
    (void)q;
    (void)err;
    // Blank identifier.
    so_R_int_err _res2 = divide(10, 3);
    so_Error err2 = _res2.err;
    (void)err2;
    so_R_int_err _res3 = divide(10, 3);
    so_int r3 = _res3.val;
    (void)r3;
    // Partial reassignment.
    so_R_int_err _res4 = divide(10, 3);
    so_int r4 = _res4.val;
    err2 = _res4.err;
    (void)r4;
    // Assign to existing variables.
    q = 0;
    err = (so_Error){};
    so_R_int_err _res5 = divide(20, 7);
    q = _res5.val;
    err = _res5.err;
}

static void testIf(void) {
    {
        so_R_int_err _res1 = divide(10, 3);
        so_int n = _res1.val;
        so_Error err = _res1.err;
        if (err.self != NULL) {
            (void)n;
        }
    }
}

static void testReturnTypes(void) {
    so_Error err = {};
    (void)err;
    so_R_int_err _res1 = returnInt();
    so_int i = _res1.val;
    err = _res1.err;
    (void)i;
    so_R_rune_err _res2 = returnRune();
    so_rune run = _res2.val;
    err = _res2.err;
    (void)run;
    so_R_str_err _res3 = returnString();
    so_String str = _res3.val;
    err = _res3.err;
    (void)str;
    so_R_slice_err _res4 = returnSlice((so_Slice){});
    so_Slice slice = _res4.val;
    err = _res4.err;
    (void)slice;
    main_FileResult _res5 = returnStruct();
    main_File struc = _res5.val;
    err = _res5.err;
    (void)struc;
    // iface, err := returnIface()
    // _ = iface
    so_R_ptr_err _res6 = returnAny();
    void* a = _res6.val;
    err = _res6.err;
    (void)a;
    so_R_ptr_err _res7 = returnPtr();
    main_File* ptr = _res7.val;
    err = _res7.err;
    (void)ptr;
}

static void testNilValues(void) {
    so_R_slice_err _res1 = returnNilSlice();
    so_Slice nilSlice = _res1.val;
    so_Error err = _res1.err;
    if (nilSlice.ptr != NULL || so_len(nilSlice) != 0 || err.self != NULL) {
        so_panic("Nil slice in multi-return failed");
    }
    so_R_ptr_err _res2 = returnNilPtr();
    main_File* nilPtr = _res2.val;
    err = _res2.err;
    if (nilPtr != NULL || err.self != NULL) {
        so_panic("Nil pointer in multi-return failed");
    }
    so_int n = 42;
    so_R_ptr_err _res3 = returnAnyInt(&n);
    void* anyInt = _res3.val;
    err = _res3.err;
    if ((*(so_int*)anyInt) != n || err.self != NULL) {
        so_panic("Any value in multi-return failed");
    }
    so_R_ptr_bool _res4 = returnAnyPair(&n);
    void* anyPair = _res4.val;
    bool ok = _res4.val2;
    if ((*(so_int*)anyPair) != n || !ok) {
        so_panic("Any value in (T, T) return failed");
    }
}

static void testForwarding(void) {
    so_Error err = {};
    (void)err;
    so_R_int_err _res1 = forwardInt();
    so_int i = _res1.val;
    err = _res1.err;
    (void)i;
    so_R_rune_err _res2 = forwardRune();
    so_rune r = _res2.val;
    err = _res2.err;
    (void)r;
    so_R_str_err _res3 = forwardString();
    so_String str = _res3.val;
    err = _res3.err;
    (void)str;
    so_R_slice_err _res4 = forwardSlice((so_Slice){});
    so_Slice slice = _res4.val;
    err = _res4.err;
    (void)slice;
    main_FileResult _res5 = forwardStruct();
    main_File struc = _res5.val;
    err = _res5.err;
    (void)struc;
    so_R_ptr_err _res6 = forwardAny();
    void* a = _res6.val;
    err = _res6.err;
    (void)a;
    so_R_ptr_err _res7 = forwardPtr();
    main_File* ptr = _res7.val;
    err = _res7.err;
    (void)ptr;
}

static void testStructExported(void) {
    main_FileResult _res1 = makeFile(42);
    main_File f = _res1.val;
    so_Error err = _res1.err;
    if (f.size != 42 || err.self != NULL) {
        so_panic("Custom exported struct failed");
    }
}

static void testStructUnexported(void) {
    pointResult _res1 = makePoint(1, 2);
    point p = _res1.val;
    so_Error err = _res1.err;
    if (p.x != 1 || p.y != 2 || err.self != NULL) {
        so_panic("Custom unexported struct failed");
    }
}

static void testStructOtherPackage(void) {
    sub_PointResult _res1 = makeSubPoint(1, 2);
    sub_Point sp1 = _res1.val;
    so_Error err = _res1.err;
    if (sp1.X != 1 || sp1.Y != 2 || err.self != NULL) {
        so_panic("Custom struct from another package failed");
    }
    sub_PointResult _res2 = sub_MakePoint(3, 4);
    sub_Point sp2 = _res2.val;
    err = _res2.err;
    if (sp2.X != 3 || sp2.Y != 4 || err.self != NULL) {
        so_panic("Custom struct from another package failed");
    }
}

int main(void) {
    testBasic();
    testIf();
    testReturnTypes();
    testNilValues();
    testForwarding();
    testStructExported();
    testStructUnexported();
    testStructOtherPackage();
    return 0;
}
