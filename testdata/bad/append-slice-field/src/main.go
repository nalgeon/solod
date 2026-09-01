package main

type box struct {
	s []int
}

func main() {
	rows := make([]box, 0, 1)
	rows = append(rows, box{s: []int{1, 2}})
	_ = rows
}
