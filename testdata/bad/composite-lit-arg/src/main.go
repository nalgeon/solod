package main

func main() {
	rows := make([][]int, 0, 1)
	// append is a macro, so the slice literal would dangle.
	rows = append(rows, []int{1, 2})
	_ = rows
}
