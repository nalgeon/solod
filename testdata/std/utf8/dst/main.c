#include "main.h"

// -- Implementation --

int main(void) {
    {
        // DecodeLastRune.
        so_Slice b = so_string_bytes(so_str("Hello, 世界"));
        so_R_rune_int _res1 = utf8_DecodeLastRune(b);
        so_rune r = _res1.val;
        so_int size = _res1.val2;
        if (r != U'界' || size != 3) {
            so_panic("DecodeLastRune failed");
        }
    }
    {
        // DecodeLastRuneInString.
        so_String str = so_str("Hello, 世界");
        so_R_rune_int _res2 = utf8_DecodeLastRuneInString(str);
        so_rune r = _res2.val;
        so_int size = _res2.val2;
        if (r != U'界' || size != 3) {
            so_panic("DecodeLastRuneInString failed");
        }
    }
    {
        // DecodeRune.
        so_Slice b = so_string_bytes(so_str("Hello, 世界"));
        so_R_rune_int _res3 = utf8_DecodeRune(b);
        so_rune r = _res3.val;
        so_int size = _res3.val2;
        if (r != U'H' || size != 1) {
            so_panic("DecodeRune failed");
        }
    }
    {
        // DecodeRuneInString.
        so_String str = so_str("Hello, 世界");
        so_R_rune_int _res4 = utf8_DecodeRuneInString(str);
        so_rune r = _res4.val;
        so_int size = _res4.val2;
        if (r != U'H' || size != 1) {
            so_panic("DecodeRuneInString failed");
        }
    }
    {
        // EncodeRune.
        so_Slice buf = so_make_slice(so_byte, 3, 3);
        so_int n = utf8_EncodeRune(buf, U'界');
        if (n != 3 || so_string_ne(so_bytes_string(buf), so_str("界"))) {
            so_panic("EncodeRune failed");
        }
    }
    {
        // RuneCount.
        so_int n = utf8_RuneCount(so_string_bytes(so_str("Hello, 世界")));
        if (n != 9) {
            so_panic("RuneCount failed");
        }
    }
    {
        // RuneCountInString.
        so_int n = utf8_RuneCountInString(so_str("Hello, 世界"));
        if (n != 9) {
            so_panic("RuneCountInString failed");
        }
    }
    {
        // RuneLen.
        so_int n = utf8_RuneLen(U'界');
        if (n != 3) {
            so_panic("RuneLen failed");
        }
    }
    {
        // ValidString.
        if (!utf8_ValidString(so_str("Hello, 世界"))) {
            so_panic("ValidString failed");
        }
    }
    {
        // AppendRune.
        so_Slice buf = so_make_slice(so_byte, 7, 10);
        so_copy(so_byte, buf, so_string_bytes(so_str("Hello, ")));
        buf = utf8_AppendRune(buf, U'界');
        if (so_string_ne(so_bytes_string(buf), so_str("Hello, 界"))) {
            so_panic("AppendRune failed");
        }
    }
}
