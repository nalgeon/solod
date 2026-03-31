#include "main.h"

// -- Implementation --

int main(void) {
    so_Slice buf = so_make_slice(so_byte, 64, 64);
    {
        // AppendBool.
        so_Slice b = strconv_AppendBool(so_slice(so_byte, buf, 0, 0), true);
        if (so_string_ne(so_bytes_string(b), so_str("true"))) {
            so_panic("AppendBool");
        }
    }
    {
        // AppendFloat.
        so_Slice b = strconv_AppendFloat(so_slice(so_byte, buf, 0, 0), 3.1415926535, 'E', -1, 32);
        if (so_string_ne(so_bytes_string(b), so_str("3.1415927E+00"))) {
            so_panic("AppendFloat 32");
        }
        b = strconv_AppendFloat(so_slice(so_byte, buf, 0, 0), 3.1415926535, 'E', -1, 64);
        if (so_string_ne(so_bytes_string(b), so_str("3.1415926535E+00"))) {
            so_panic("AppendFloat 64");
        }
    }
    {
        // AppendInt.
        so_Slice b = strconv_AppendInt(so_slice(so_byte, buf, 0, 0), -42, 10);
        if (so_string_ne(so_bytes_string(b), so_str("-42"))) {
            so_panic("AppendInt base 10");
        }
        b = strconv_AppendInt(so_slice(so_byte, buf, 0, 0), -42, 16);
        if (so_string_ne(so_bytes_string(b), so_str("-2a"))) {
            so_panic("AppendInt base 16");
        }
    }
    {
        // AppendUint.
        so_Slice b = strconv_AppendUint(so_slice(so_byte, buf, 0, 0), 42, 10);
        if (so_string_ne(so_bytes_string(b), so_str("42"))) {
            so_panic("AppendUint base 10");
        }
        b = strconv_AppendUint(so_slice(so_byte, buf, 0, 0), 42, 16);
        if (so_string_ne(so_bytes_string(b), so_str("2a"))) {
            so_panic("AppendUint base 16");
        }
    }
    {
        // Atoi.
        so_R_int_err _res1 = strconv_Atoi(so_str("10"));
        so_int s = _res1.val;
        so_Error err = _res1.err;
        if (err != NULL) {
            so_panic("Atoi error");
        }
        if (s != 10) {
            so_panic("Atoi value");
        }
    }
    {
        // FormatBool.
        so_String s = strconv_FormatBool(true);
        if (so_string_ne(s, so_str("true"))) {
            so_panic("FormatBool");
        }
    }
    {
        // FormatFloat.
        so_String s = strconv_FormatFloat(buf, 3.1415926535, 'E', -1, 32);
        if (so_string_ne(s, so_str("3.1415927E+00"))) {
            so_panic("FormatFloat 32");
        }
        s = strconv_FormatFloat(buf, 3.1415926535, 'E', -1, 64);
        if (so_string_ne(s, so_str("3.1415926535E+00"))) {
            so_panic("FormatFloat 64");
        }
        s = strconv_FormatFloat(buf, 3.1415926535, 'g', -1, 64);
        if (so_string_ne(s, so_str("3.1415926535"))) {
            so_panic("FormatFloat g");
        }
    }
    {
        // FormatInt.
        so_String s = strconv_FormatInt(buf, -42, 10);
        if (so_string_ne(s, so_str("-42"))) {
            so_panic("FormatInt base 10");
        }
        s = strconv_FormatInt(buf, -42, 16);
        if (so_string_ne(s, so_str("-2a"))) {
            so_panic("FormatInt base 16");
        }
    }
    {
        // FormatUint.
        so_String s = strconv_FormatUint(buf, 42, 10);
        if (so_string_ne(s, so_str("42"))) {
            so_panic("FormatUint base 10");
        }
        s = strconv_FormatUint(buf, 42, 16);
        if (so_string_ne(s, so_str("2a"))) {
            so_panic("FormatUint base 16");
        }
    }
    {
        // Itoa.
        so_String s = strconv_Itoa(buf, 10);
        if (so_string_ne(s, so_str("10"))) {
            so_panic("Itoa");
        }
    }
    {
        // ParseBool.
        so_R_bool_err _res2 = strconv_ParseBool(so_str("true"));
        bool s = _res2.val;
        so_Error err = _res2.err;
        if (err != NULL) {
            so_panic("ParseBool error");
        }
        if (!s) {
            so_panic("ParseBool value");
        }
    }
    {
        // ParseFloat.
        so_R_f64_err _res3 = strconv_ParseFloat(so_str("3.1415926535"), 32);
        double s = _res3.val;
        so_Error err = _res3.err;
        if (err != NULL) {
            so_panic("ParseFloat 32 error");
        }
        so_String r = strconv_FormatFloat(buf, s, 'E', -1, 32);
        if (so_string_ne(r, so_str("3.1415927E+00"))) {
            so_panic("ParseFloat 32 value");
        }
        so_R_f64_err _res4 = strconv_ParseFloat(so_str("3.1415926535"), 64);
        s = _res4.val;
        err = _res4.err;
        if (err != NULL) {
            so_panic("ParseFloat 64 error");
        }
        if (s != 3.1415926535) {
            so_panic("ParseFloat 64 value");
        }
        // NaN.
        so_R_f64_err _res5 = strconv_ParseFloat(so_str("NaN"), 32);
        s = _res5.val;
        err = _res5.err;
        if (err != NULL) {
            so_panic("ParseFloat NaN error");
        }
        if (s == s) {
            so_panic("ParseFloat NaN value");
        }
        // Case insensitive.
        so_R_f64_err _res6 = strconv_ParseFloat(so_str("nan"), 32);
        s = _res6.val;
        err = _res6.err;
        if (err != NULL) {
            so_panic("ParseFloat nan error");
        }
        if (s == s) {
            so_panic("ParseFloat nan value");
        }
        // inf.
        so_R_f64_err _res7 = strconv_ParseFloat(so_str("inf"), 32);
        s = _res7.val;
        err = _res7.err;
        if (err != NULL) {
            so_panic("ParseFloat inf error");
        }
        r = strconv_FormatFloat(buf, s, 'g', -1, 64);
        if (so_string_ne(r, so_str("+Inf"))) {
            so_panic("ParseFloat inf value");
        }
        // +Inf.
        so_R_f64_err _res8 = strconv_ParseFloat(so_str("+Inf"), 32);
        s = _res8.val;
        err = _res8.err;
        if (err != NULL) {
            so_panic("ParseFloat +Inf error");
        }
        r = strconv_FormatFloat(buf, s, 'g', -1, 64);
        if (so_string_ne(r, so_str("+Inf"))) {
            so_panic("ParseFloat +Inf value");
        }
        // -Inf.
        so_R_f64_err _res9 = strconv_ParseFloat(so_str("-Inf"), 32);
        s = _res9.val;
        err = _res9.err;
        if (err != NULL) {
            so_panic("ParseFloat -Inf error");
        }
        r = strconv_FormatFloat(buf, s, 'g', -1, 64);
        if (so_string_ne(r, so_str("-Inf"))) {
            so_panic("ParseFloat -Inf value");
        }
        // -0.
        so_R_f64_err _res10 = strconv_ParseFloat(so_str("-0"), 32);
        s = _res10.val;
        err = _res10.err;
        if (err != NULL) {
            so_panic("ParseFloat -0 error");
        }
        r = strconv_FormatFloat(buf, s, 'g', -1, 64);
        if (so_string_ne(r, so_str("-0"))) {
            so_panic("ParseFloat -0 value");
        }
        // +0.
        so_R_f64_err _res11 = strconv_ParseFloat(so_str("+0"), 32);
        s = _res11.val;
        err = _res11.err;
        if (err != NULL) {
            so_panic("ParseFloat +0 error");
        }
        if (s != 0) {
            so_panic("ParseFloat +0 value");
        }
    }
    {
        // ParseInt.
        so_R_i64_err _res12 = strconv_ParseInt(so_str("-354634382"), 10, 32);
        int64_t s = _res12.val;
        so_Error err = _res12.err;
        if (err != NULL) {
            so_panic("ParseInt 32 error");
        }
        if (s != -354634382) {
            so_panic("ParseInt 32 value");
        }
        so_R_i64_err _res13 = strconv_ParseInt(so_str("-3546343826724305832"), 10, 64);
        s = _res13.val;
        err = _res13.err;
        if (err != NULL) {
            so_panic("ParseInt 64 error");
        }
        if (s != -3546343826724305832) {
            so_panic("ParseInt 64 value");
        }
    }
    {
        // ParseUint.
        so_R_u64_err _res14 = strconv_ParseUint(so_str("42"), 10, 32);
        uint64_t s = _res14.val;
        so_Error err = _res14.err;
        if (err != NULL) {
            so_panic("ParseUint 32 error");
        }
        if (s != 42) {
            so_panic("ParseUint 32 value");
        }
        so_R_u64_err _res15 = strconv_ParseUint(so_str("42"), 10, 64);
        s = _res15.val;
        err = _res15.err;
        if (err != NULL) {
            so_panic("ParseUint 64 error");
        }
        if (s != 42) {
            so_panic("ParseUint 64 value");
        }
    }
}
