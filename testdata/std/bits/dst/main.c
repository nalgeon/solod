#include "main.h"

// -- Implementation --

int main(void) {
    {
        // Add32.
        uint32_t n1 = (uint32_t)(0b0101);
        uint32_t n2 = (uint32_t)(0b0011);
        so_R_u32_u32 _res1 = bits_Add32(n1, n2, 0);
        uint32_t d = _res1.val;
        uint32_t carry = _res1.val2;
        if (d != 0b1000 || carry != 0) {
            so_panic("Add32 failed");
        }
    }
    {
        // Sub32.
        uint32_t n1 = (uint32_t)(0b0101);
        uint32_t n2 = (uint32_t)(0b0011);
        so_R_u32_u32 _res2 = bits_Sub32(n1, n2, 0);
        uint32_t d = _res2.val;
        uint32_t borrow = _res2.val2;
        if (d != 0b0010 || borrow != 0) {
            so_panic("Sub32 failed");
        }
    }
    {
        // Mul32.
        uint32_t n1 = (uint32_t)(0b0101);
        uint32_t n2 = (uint32_t)(0b0011);
        so_R_u32_u32 _res3 = bits_Mul32(n1, n2);
        uint32_t dh = _res3.val;
        uint32_t dl = _res3.val2;
        if (dh != 0 || dl != 0b1111) {
            so_panic("Mul32 failed");
        }
    }
    {
        // LeadingZeros8.
        uint8_t n = (uint8_t)(0b00010000);
        if (bits_LeadingZeros8(n) != 3) {
            so_panic("LeadingZeros8 failed");
        }
    }
    {
        // TrailingZeros8.
        uint8_t n = (uint8_t)(0b00010000);
        if (bits_TrailingZeros8(n) != 4) {
            so_panic("TrailingZeros8 failed");
        }
    }
    {
        // OnesCount.
        uint64_t n = (uint64_t)(0b101010);
        if (bits_OnesCount(n) != 3) {
            so_panic("OnesCount failed");
        }
    }
    {
        // RotateLeft8.
        uint8_t n = (uint8_t)(0b00001111);
        if (bits_RotateLeft8(n, 2) != 0b00111100) {
            so_panic("RotateLeft8 failed");
        }
    }
    {
        // Reverse8.
        uint8_t n = (uint8_t)(0b00001111);
        if (bits_Reverse8(n) != 0b11110000) {
            so_panic("Reverse8 failed");
        }
    }
    {
        // ReverseBytes16.
        uint16_t n = (uint16_t)(0x1234);
        if (bits_ReverseBytes16(n) != 0x3412) {
            so_panic("ReverseBytes16 failed");
        }
    }
    {
        // Len8.
        uint8_t n = (uint8_t)(0b00001111);
        if (bits_Len8(n) != 4) {
            so_panic("Len8 failed");
        }
    }
}
