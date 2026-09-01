package main

type box struct {
	v any
}

func main() {
	rows := make([]box, 0, 1)
	rows = append(rows, box{v: 42})
	_ = rows
}
