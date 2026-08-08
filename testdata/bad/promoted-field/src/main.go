package main

type hidden struct {
	n int
}

//so:promote
type shown struct {
	h hidden
}

func main() {
	_ = shown{}
}
