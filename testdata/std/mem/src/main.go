package main

import (
	"github.com/nalgeon/solod/so/mem"
)

type Point struct {
	x, y int
}

func withDefer() {
	p := mem.Alloc[Point](nil)
	defer mem.Free(nil, p)

	p.x = 11
	p.y = 22
	if p.x != 11 || p.y != 22 {
		panic("unexpected value")
	}
}

func main() {
	{
		// TryAlloc and Free.
		p, err := mem.TryAlloc[Point](mem.System)
		if err != nil {
			panic("Alloc: allocation failed")
		}
		p.x = 11
		p.y = 22
		if p.x != 11 || p.y != 22 {
			panic("Alloc: unexpected value")
		}
		mem.Free(mem.System, p)
	}
	{
		// TryAllocSlice and FreeSlice.
		slice, err := mem.TryAllocSlice[int](mem.System, 3, 3)
		if err != nil {
			panic("AllocSlice: allocation failed")
		}
		slice[0] = 11
		slice[1] = 22
		slice[2] = 33
		if slice[0] != 11 || slice[1] != 22 || slice[2] != 33 {
			panic("AllocSlice: unexpected value")
		}
		mem.FreeSlice(mem.System, slice)
	}
	{
		// Alloc/Free with default allocator.
		p := mem.Alloc[Point](nil)
		p.x = 11
		p.y = 22
		if p.x != 11 || p.y != 22 {
			panic("New: unexpected value")
		}
		mem.Free(nil, p)
	}
	{
		// AllocSlice/FreeSlice with default allocator.
		slice := mem.AllocSlice[int](nil, 3, 3)
		slice[0] = 11
		slice[1] = 22
		slice[2] = 33
		if slice[0] != 11 || slice[1] != 22 || slice[2] != 33 {
			panic("NewSlice: unexpected value")
		}
		mem.FreeSlice(nil, slice)
	}
	withDefer()
	{
		// Append within capacity.
		s := mem.AllocSlice[int](nil, 0, 8)
		s = mem.Append(nil, s, 10, 20, 30)
		if len(s) != 3 || s[0] != 10 || s[1] != 20 || s[2] != 30 {
			panic("Append: unexpected value")
		}
		mem.FreeSlice(nil, s)
	}
	{
		// Append that triggers growth.
		s := mem.AllocSlice[int](nil, 0, 2)
		s = mem.Append(nil, s, 1, 2)
		s = mem.Append(nil, s, 3, 4, 5)
		if len(s) != 5 || s[0] != 1 || s[4] != 5 {
			panic("Append grow: unexpected value")
		}
		mem.FreeSlice(nil, s)
	}
	{
		// Extend from another slice.
		s := mem.AllocSlice[int](nil, 0, 8)
		other := []int{100, 200, 300}
		s = mem.Extend(nil, s, other)
		if len(s) != 3 || s[0] != 100 || s[2] != 300 {
			panic("Extend: unexpected value")
		}
		mem.FreeSlice(nil, s)
	}
	{
		// TryAppend success.
		s := mem.AllocSlice[int](nil, 0, 4)
		s, err := mem.TryAppend(nil, s, 42)
		if err != nil {
			panic("TryAppend: unexpected error")
		}
		if len(s) != 1 || s[0] != 42 {
			panic("TryAppend: unexpected value")
		}
		mem.FreeSlice(nil, s)
	}
}
