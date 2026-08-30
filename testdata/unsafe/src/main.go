package main

import "unsafe"

type header struct {
	kind  uint8
	size  int64
	flags uint32 `c:"attrs"`
}

type packet struct {
	head header
	body [8]byte
	name string
	nums []int
}

// Named pointer type.
type packetPtr *packet

func main() {
	testOffsetof()
	testPointerConv()
}

func testOffsetof() {
	// Offset of a field in a struct value.
	var p packet
	if unsafe.Offsetof(p.head) != 0 {
		panic("unexpected packet.head offset")
	}
	if unsafe.Offsetof(p.body) != unsafe.Sizeof(p.head) {
		panic("unexpected packet.body offset")
	}

	// Offset of a field in a nested struct.
	if unsafe.Offsetof(p.head.size) < unsafe.Sizeof(p.head.kind) {
		panic("unexpected header.size offset")
	}

	// The C name of the field comes from the c tag.
	if unsafe.Offsetof(p.head.flags) < unsafe.Offsetof(p.head.size) {
		panic("unexpected header.flags offset")
	}

	// Offset of a field through a pointer.
	pp := &p
	pp.head.kind = 1
	if unsafe.Offsetof(pp.body) != unsafe.Sizeof(p.head) {
		panic("unexpected packet.body offset")
	}

	// Offset as a constant.
	const off = unsafe.Offsetof(p.body)
	if off != unsafe.Sizeof(p.head) {
		panic("unexpected packet.body offset")
	}
}

func testPointerConv() {
	p := packet{name: "hello", nums: []int{1, 2, 3}}
	ptr := unsafe.Pointer(&p)

	// Field access through a converted pointer.
	(*packet)(ptr).head.kind = 7
	(*packet)(ptr).head.size++
	if (*packet)(ptr).head.kind != 7 || p.head.size != 1 {
		panic("unexpected header fields")
	}

	// A string field keeps the parentheses for .len and .ptr.
	println((*packet)(ptr).name)
	if (*packet)(ptr).name[1:] != "ello" {
		panic("unexpected name suffix")
	}

	// A slice field does the same.
	if len((*packet)(ptr).nums[1:]) != 2 {
		panic("unexpected nums length")
	}

	// Conversion to a named pointer type.
	if packetPtr(&p).head.kind != 7 {
		panic("unexpected header.kind")
	}
}
