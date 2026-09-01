package main

import "unsafe"

func main() {
	println(unsafe.Sizeof(map[int]int{1: 2}))
}
