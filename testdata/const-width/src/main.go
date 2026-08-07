package main

import "unsafe"

// Constants whose value depends on the width of int. Go computes them at the
// host's width, so folding them into the C would freeze that width into the
// output. C reproduces them on its own, at the target's width.
const maxInt = int(uint64(^uint(0)) >> 1)
const maxUint = uint(^uint(0))
const uintSize = 32 << (uint64(^uint(0)) >> 63)
const ptrSize = 4 << (uint64(^uintptr(0)) >> 63)
const hiBits = 0x8080808080808080 >> (64 - 8*ptrSize)

// Constants C cannot reproduce, so it gets the computed value instead: an
// untyped intermediate above int64, and a shift count that reaches the width
// of the type it evaluates in.
const untypedBig = (1 << 100) >> 90
const wideShift uint32 = 0 << 40

// Here only the intermediate is beyond C's reach. It folds on its own, and the
// width-dependent shift around it stays.
const wideMask = (1<<64 - 1) >> (64 - 8*ptrSize)

func main() {
	// unsafe.Sizeof is also constant in Go, but it is emitted as C sizeof,
	// so these are the widths the C compiler actually chose.
	intSize := 8 * int(unsafe.Sizeof(int(0)))
	realPtrSize := int(unsafe.Sizeof(uintptr(0)))

	if uintSize != intSize {
		panic("uintSize")
	}
	if ptrSize != realPtrSize {
		panic("ptrSize")
	}
	if maxUint != uint(maxInt)*2+1 {
		panic("maxInt")
	}
	if maxUint>>(uintSize-1) != 1 {
		panic("maxUint")
	}

	// hiBits is the high bit of every byte of a word.
	var want uintptr
	for i := 0; i < ptrSize; i++ {
		want |= 0x80 << (8 * i)
	}
	if uintptr(hiBits) != want {
		panic("hiBits")
	}

	// wideMask is every bit of a word set, and no more. Comparing as uint64
	// keeps a value too wide for the target from being truncated into agreeing.
	if uint64(wideMask) != uint64(^uintptr(0)) {
		panic("wideMask")
	}

	if untypedBig != 1024 || wideShift != 0 {
		panic("folded")
	}
}
