package main

type stream struct {
	stderr int
}

func main() {
	s := stream{stderr: 1}
	_ = s
}
