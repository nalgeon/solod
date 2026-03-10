#include "main.h"

// -- Implementation --

int main(void) {
    if (!ctype_IsAlnum(U'a')) {
        so_panic("want IsAlnum = true");
    }
    if (!ctype_IsAlpha(U'a')) {
        so_panic("want IsAlpha = true");
    }
    if (!ctype_IsBlank(U' ')) {
        so_panic("want IsBlank = true");
    }
    if (!ctype_IsCntrl(0x1F)) {
        so_panic("want IsCntrl = true");
    }
    if (!ctype_IsDigit(U'7')) {
        so_panic("want IsDigit = true");
    }
    if (!ctype_IsPunct(U',')) {
        so_panic("want IsPunct = true");
    }
    if (!ctype_IsSpace(U'\n')) {
        so_panic("want IsSpace = true");
    }
    if (!ctype_IsUpper(U'A')) {
        so_panic("want IsUpper = true");
    }
    if (!ctype_IsXDigit(U'B')) {
        so_panic("want IsXDigit = true");
    }
    if (ctype_ToLower(U'A') != U'a') {
        so_panic("want ToLower(A) = a");
    }
    if (ctype_ToUpper(U'a') != U'A') {
        so_panic("want ToUpper(a) = A");
    }
}
