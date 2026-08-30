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
}

func main() {
	testOffsetof()
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
