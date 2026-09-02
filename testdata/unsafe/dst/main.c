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
    so_String name;
    so_Slice nums;
} packet;

// Named pointer type.
typedef packet* packetPtr;

// -- Forward declarations --
static void testOffsetof(void);
static void testSizeof(void);
static void testPointerConv(void);

// -- Variables and constants --

// Size of a struct literal, as a package-level constant.
static const so_unused uintptr_t headSize = unsafe_Sizeof(header);

// -- Implementation --

int main(void) {
    testOffsetof();
    testSizeof();
    testPointerConv();
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

static void testSizeof(void) {
    packet p = {};
    if (headSize != unsafe_Sizeof(p.head)) {
        so_panic("unexpected header size");
    }
    // A struct literal with several fields.
    if (unsafe_Sizeof(header) != headSize) {
        so_panic("unexpected header literal size");
    }
    if (unsafe_Alignof(header) != unsafe_Alignof(p.head)) {
        so_panic("unexpected header literal alignment");
    }
    // An array literal.
    if (unsafe_Sizeof(so_byte[8]) != 8) {
        so_panic("unexpected byte array size");
    }
    if (unsafe_Sizeof(int64_t[2]) != 16) {
        so_panic("unexpected int array size");
    }
    if (unsafe_Alignof(int64_t[2]) != unsafe_Alignof((int64_t)(0))) {
        so_panic("unexpected int array alignment");
    }
    // A two-dimensional array literal.
    if (unsafe_Sizeof(int64_t[2][3]) != 48) {
        so_panic("unexpected 2d array size");
    }
    // A slice literal has the size of a slice header.
    if (unsafe_Sizeof(so_Slice) != unsafe_Sizeof(p.nums)) {
        so_panic("unexpected slice literal size");
    }
}

static void testPointerConv(void) {
    packet p = (packet){.name = so_str("hello"), .nums = (so_Slice){(so_int[3]){1, 2, 3}, 3, 3}};
    void* ptr = (void*)(&p);
    // Field access through a converted pointer.
    ((packet*)(ptr))->head.kind = 7;
    ((packet*)(ptr))->head.size++;
    if (((packet*)(ptr))->head.kind != 7 || p.head.size != 1) {
        so_panic("unexpected header fields");
    }
    // A string field keeps the parentheses for .len and .ptr.
    so_println("%.*s", ((packet*)(ptr))->name.len, ((packet*)(ptr))->name.ptr);
    if (so_string_ne(so_string_slice(((packet*)(ptr))->name, 1, ((packet*)(ptr))->name.len), so_str("ello"))) {
        so_panic("unexpected name suffix");
    }
    // A slice field does the same.
    if (so_len(so_slice(so_int, ((packet*)(ptr))->nums, 1, ((packet*)(ptr))->nums.len)) != 2) {
        so_panic("unexpected nums length");
    }
    // Conversion to a named pointer type.
    if (((packetPtr)(&p))->head.kind != 7) {
        so_panic("unexpected header.kind");
    }
}
