package main

func send(c chan int, x int) {
	_ = c
	_ = x
}

func main() {
	send(nil, 1)
}
