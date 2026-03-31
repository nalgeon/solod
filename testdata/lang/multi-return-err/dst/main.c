#include "main.h"

// -- Variables and constants --
static main_File file = {0};

// -- Forward declarations --
static so_R_int_err divide(so_int a, so_int b);
static so_R_rune_err returnRune(void);
static so_R_str_err returnString(void);
static so_R_slice_err returnSlice(void);
static so_R_ptr_err returnPtr(void);
static so_R_int_err forwardCall(void);
static main_FileResult create(so_int size);

// -- Implementation --

so_R_int_err main_File_Read(void* self, so_int buf) {
    main_File* f = (main_File*)self;
    (void)buf;
    return (so_R_int_err){.val = f->size, .err = NULL};
}

static so_R_int_err divide(so_int a, so_int b) {
    return (so_R_int_err){.val = a / b, .err = NULL};
}

static so_R_rune_err returnRune(void) {
    return (so_R_rune_err){.val = U'x', .err = NULL};
}

static so_R_str_err returnString(void) {
    return (so_R_str_err){.val = so_str("hello"), .err = NULL};
}

static so_R_slice_err returnSlice(void) {
    return (so_R_slice_err){.val = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3}, .err = NULL};
}

static so_R_ptr_err returnPtr(void) {
    return (so_R_ptr_err){.val = &file, .err = NULL};
}

static so_R_int_err forwardCall(void) {
    return divide(10, 3);
}

static main_FileResult create(so_int size) {
    return (main_FileResult){.val = (main_File){.size = size}, .err = NULL};
}

int main(void) {
    {
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
        err = NULL;
        so_R_int_err _res5 = divide(20, 7);
        q = _res5.val;
        err = _res5.err;
    }
    {
        // If-init with multi-return.
        main_File f = (main_File){.size = 42};
        {
            so_R_int_err _res6 = main_File_Read(&f, 64);
            so_int n = _res6.val;
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
        so_R_rune_err _res7 = returnRune();
        so_rune run = _res7.val;
        err = _res7.err;
        (void)run;
        so_R_str_err _res8 = returnString();
        so_String str = _res8.val;
        err = _res8.err;
        (void)str;
        so_R_slice_err _res9 = returnSlice();
        so_Slice slice = _res9.val;
        err = _res9.err;
        (void)slice;
        // struc, err := returnStruct()
        // _ = struc
        so_R_ptr_err _res10 = returnPtr();
        main_File* ptr = _res10.val;
        err = _res10.err;
        // iface, err := returnIface()
        // _ = iface
        (void)ptr;
    }
    {
        // Forward call.
        so_R_int_err _res11 = forwardCall();
        so_int q = _res11.val;
        so_Error err = _res11.err;
        (void)q;
        (void)err;
    }
    {
        // Custom struct + error.
        main_FileResult _res12 = create(42);
        main_File f = _res12.val;
        so_Error err = _res12.err;
        if (f.size != 42 || err != NULL) {
            so_panic("Custom struct failed");
        }
    }
}
