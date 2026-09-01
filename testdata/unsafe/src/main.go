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

// Size of a struct literal, as a package-level constant.
const headSize = unsafe.Sizeof(header{})

func main() {
	testOffsetof()
	testSizeof()
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

func testSizeof() {
	var p packet
	if headSize != unsafe.Sizeof(p.head) {
		panic("unexpected header size")
	}

	// A struct literal with several fields.
	if unsafe.Sizeof(header{1, 2, 3}) != headSize {
		panic("unexpected header literal size")
	}
	if unsafe.Alignof(header{1, 2, 3}) != unsafe.Alignof(p.head) {
		panic("unexpected header literal alignment")
	}

	// An array literal.
	if unsafe.Sizeof([8]byte{}) != 8 {
		panic("unexpected byte array size")
	}
	if unsafe.Sizeof([2]int64{1, 2}) != 16 {
		panic("unexpected int array size")
	}
	if unsafe.Alignof([2]int64{1, 2}) != unsafe.Alignof(int64(0)) {
		panic("unexpected int array alignment")
	}

	// A two-dimensional array literal.
	if unsafe.Sizeof([2][3]int64{}) != 48 {
		panic("unexpected 2d array size")
	}

	// A slice literal has the size of a slice header.
	if unsafe.Sizeof([]int{1, 2}) != unsafe.Sizeof(p.nums) {
		panic("unexpected slice literal size")
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
