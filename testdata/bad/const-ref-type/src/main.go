package main

import "unsafe"

type point struct{ x, y int32 }

const Size = unsafe.Sizeof(point{})

func main() {
	println(Size)
}
