#include "main.h"

// -- Forward declarations (functions and methods) --
static so_Result divide(so_int a, so_int b);
static so_Result returnRune(void);
static so_Result returnString(void);
static so_Result returnSlice(void);
static so_Result returnPtr(void);
static so_Result forwardCall(void);

// -- Implementation --
static main_File file = {0};

so_Result main_File_Read(void* self, so_int buf) {
    main_File* f = (main_File*)self;
    (void)buf;
    return (so_Result){.val.as_int = f->size, .err = NULL};
}

static so_Result divide(so_int a, so_int b) {
    return (so_Result){.val.as_int = a / b, .err = NULL};
}

static so_Result returnRune(void) {
    return (so_Result){.val.as_rune = U'x', .err = NULL};
}

static so_Result returnString(void) {
    return (so_Result){.val.as_string = so_strlit("hello"), .err = NULL};
}

static so_Result returnSlice(void) {
    return (so_Result){.val.as_slice = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3}, .err = NULL};
}

// Returning struct values is not supported.
// func returnStruct() (File, error) {
// 	return File{size: 42}, nil
// }
static so_Result returnPtr(void) {
    return (so_Result){.val.as_ptr = &file, .err = NULL};
}

// Returning interface values is not supported.
// func returnIface() (Reader, error) {
// 	return &file, nil
// }
static so_Result forwardCall(void) {
    return divide(10, 3);
}

int main(void) {
    {
        // Destructure into new variables.
        so_Result _res1 = divide(10, 3);
        so_int q = _res1.val.as_int;
        so_Error err = _res1.err;
        (void)q;
        (void)err;
        // Blank identifier.
        so_Result _res2 = divide(10, 3);
        so_Error err2 = _res2.err;
        (void)err2;
        so_Result _res3 = divide(10, 3);
        so_int r3 = _res3.val.as_int;
        (void)r3;
        // Partial reassignment.
        so_Result _res4 = divide(10, 3);
        so_int r4 = _res4.val.as_int;
        err2 = _res4.err;
        (void)r4;
        // Assign to existing variables.
        q = 0;
        err = NULL;
        so_Result _res5 = divide(20, 7);
        q = _res5.val.as_int;
        err = _res5.err;
    }
    {
        // If-init with multi-return.
        main_File f = (main_File){.size = 42};
        {
            so_Result _res6 = main_File_Read(&f, 64);
            so_int n = _res6.val.as_int;
            so_Error err = _res6.err;
            if (err != NULL) {
                (void)n;
            }
        }
    }
    {
        // Various return types.
        so_Error err = NULL;
        (void)err;
        so_Result _res7 = returnRune();
        int32_t run = _res7.val.as_rune;
        err = _res7.err;
        (void)run;
        so_Result _res8 = returnString();
        so_String str = _res8.val.as_string;
        err = _res8.err;
        (void)str;
        so_Result _res9 = returnSlice();
        so_Slice slice = _res9.val.as_slice;
        err = _res9.err;
        (void)slice;
        // struc, err := returnStruct()
        // _ = struc
        so_Result _res10 = returnPtr();
        main_File* ptr = _res10.val.as_ptr;
        err = _res10.err;
        // iface, err := returnIface()
        // _ = iface
        (void)ptr;
    }
    {
        // Forward call.
        so_Result _res11 = forwardCall();
        so_int q = _res11.val.as_int;
        so_Error err = _res11.err;
        (void)q;
        (void)err;
    }
}
