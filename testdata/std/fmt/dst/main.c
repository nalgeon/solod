#include "main.h"

// -- Implementation --

int main(void) {
    {
        // Print.
        so_R_int_err _res1 = fmt_Print("hello", "world");
        so_int n = _res1.val;
        so_Error err = _res1.err;
        if (err != NULL) {
            so_panic("Print failed");
        }
        if (n != 11) {
            so_panic("Print: wrong count");
        }
        fmt_Print("\n");
    }
    {
        // Println.
        so_R_int_err _res2 = fmt_Println("hello", "world");
        so_int n = _res2.val;
        so_Error err = _res2.err;
        if (err != NULL) {
            so_panic("Println failed");
        }
        if (n != 12) {
            so_panic("Println: wrong count");
        }
    }
    {
        // Printf.
        so_String s = so_str("world");
        so_int d = 42;
        so_R_int_err _res3 = fmt_Printf("s = %s, d = %d\n", so_cstr(s), d);
        so_int n = _res3.val;
        so_Error err = _res3.err;
        if (err != NULL) {
            so_panic("Printf failed");
        }
        if (n != 18) {
            so_panic("Printf: wrong count");
        }
    }
    {
        // Sprintf.
        fmt_Buffer buf = fmt_NewBuffer(32);
        so_String s = so_str("world");
        so_int d = 42;
        so_String out = fmt_Sprintf(buf, "s = %s, d = %d", so_cstr(s), d);
        if (so_string_ne(out, so_str("s = world, d = 42"))) {
            so_panic("Sprintf: wrong output");
        }
    }
    {
        // Fprintf.
        strings_Builder sb = {0};
        int32_t i = 42;
        so_String s = so_str("world");
        so_R_int_err _res4 = fmt_Fprintf((io_Writer){.self = &sb, .Write = strings_Builder_Write}, "hello %d %s", i, so_cstr(s));
        so_int n = _res4.val;
        so_Error err = _res4.err;
        if (err != NULL) {
            so_panic("Fprintf failed");
        }
        if (n != 14) {
            so_panic("Fprintf: wrong count");
        }
        if (so_string_ne(strings_Builder_String(&sb), so_str("hello 42 world"))) {
            so_panic("Fprintf: wrong output");
        }
        strings_Builder_Free(&sb);
    }
    {
        // Sscanf.
        int32_t n1 = 0, n2 = 0;
        fmt_Buffer buf = fmt_NewBuffer(32);
        so_R_int_err _res5 = fmt_Sscanf("5 1 gophers", "%d %d %s", &n1, &n2, buf.Ptr);
        so_int n = _res5.val;
        so_Error err = _res5.err;
        if (err != NULL) {
            so_panic("Sscanf failed");
        }
        so_String s = fmt_Buffer_String(buf);
        if (n != 3) {
            so_panic("Sscanf: wrong count");
        }
        if (n1 != 5 || n2 != 1 || so_string_ne(s, so_str("gophers"))) {
            so_panic("Sscanf: wrong values");
        }
    }
    {
        // Fscanf.
        int32_t n1 = 0, n2 = 0;
        fmt_Buffer buf = fmt_NewBuffer(32);
        strings_Reader r = strings_NewReader(so_str("5 1 gophers"));
        so_R_int_err _res6 = fmt_Fscanf((io_Reader){.self = &r, .Read = strings_Reader_Read}, "%d %d %s", &n1, &n2, buf.Ptr);
        so_int n = _res6.val;
        so_Error err = _res6.err;
        if (err != NULL) {
            so_panic("Fscanf failed");
        }
        so_String s = fmt_Buffer_String(buf);
        if (n != 3) {
            so_panic("Fscanf: wrong count");
        }
        if (n1 != 5 || n2 != 1 || so_string_ne(s, so_str("gophers"))) {
            so_panic("Fscanf: wrong values");
        }
    }
}
