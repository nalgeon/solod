#include "main.h"

// -- Types --

typedef struct header header;
typedef struct packet packet;

typedef struct header {
    uint8_t kind;
    int64_t size;
    uint32_t attrs;
} header;

typedef struct packet {
    header head;
    so_byte body[8];
} packet;

// -- Forward declarations --
static void testOffsetof(void);

// -- Implementation --

int main(void) {
    testOffsetof();
    return 0;
}

static void testOffsetof(void) {
    // Offset of a field in a struct value.
    packet p = {};
    if (unsafe_Offsetof(packet, head) != 0) {
        so_panic("unexpected packet.head offset");
    }
    if (unsafe_Offsetof(packet, body) != unsafe_Sizeof(p.head)) {
        so_panic("unexpected packet.body offset");
    }
    // Offset of a field in a nested struct.
    if (unsafe_Offsetof(header, size) < unsafe_Sizeof(p.head.kind)) {
        so_panic("unexpected header.size offset");
    }
    // The C name of the field comes from the c tag.
    if (unsafe_Offsetof(header, attrs) < unsafe_Offsetof(header, size)) {
        so_panic("unexpected header.flags offset");
    }
    // Offset of a field through a pointer.
    packet* pp = &p;
    pp->head.kind = 1;
    if (unsafe_Offsetof(packet, body) != unsafe_Sizeof(p.head)) {
        so_panic("unexpected packet.body offset");
    }
    // Offset as a constant.
    const so_unused uintptr_t off = unsafe_Offsetof(packet, body);
    if (off != unsafe_Sizeof(p.head)) {
        so_panic("unexpected packet.body offset");
    }
}
