package main

// The slice element walk goes through named types and finds the array.
type Cell [2]int
type Row []Cell
type Grid []Row

func main() {
	var g Grid
	_ = g
}
