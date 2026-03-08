package main

import (
	"github.com/nalgeon/solod/so/mem"
)

type Point struct {
	x, y int
}

func withDefer() {
	p := mem.New[Point]()
	defer mem.Free(p)

	p.x = 11
	p.y = 22
	if p.x != 11 || p.y != 22 {
		panic("unexpected value")
	}
}

func main() {
	{
		// mem.Alloc and mem.Dealloc
		p, err := mem.Alloc[Point](mem.System)
		if err != nil {
			panic("Alloc: allocation failed")
		}
		p.x = 11
		p.y = 22
		if p.x != 11 || p.y != 22 {
			panic("Alloc: unexpected value")
		}
		mem.Dealloc[Point](mem.System, p)
	}
	{
		// mem.AllocSlice and mem.DeallocSlice
		slice, err := mem.AllocSlice[int](mem.System, 3, 3)
		if err != nil {
			panic("AllocSlice: allocation failed")
		}
		slice[0] = 11
		slice[1] = 22
		slice[2] = 33
		if slice[0] != 11 || slice[1] != 22 || slice[2] != 33 {
			panic("AllocSlice: unexpected value")
		}
		mem.DeallocSlice[int](mem.System, slice)
	}
	{
		// mem.New and mem.Free
		p := mem.New[Point]()
		p.x = 11
		p.y = 22
		if p.x != 11 || p.y != 22 {
			panic("New: unexpected value")
		}
		mem.Free[Point](p)
	}
	{
		// mem.NewSlice and mem.FreeSlice
		slice := mem.NewSlice[int](3, 3)
		slice[0] = 11
		slice[1] = 22
		slice[2] = 33
		if slice[0] != 11 || slice[1] != 22 || slice[2] != 33 {
			panic("NewSlice: unexpected value")
		}
		mem.FreeSlice[int](slice)
	}
	withDefer()
}
