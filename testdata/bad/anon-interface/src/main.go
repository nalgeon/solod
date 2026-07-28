package main

func main() {
	var r interface{ Read() int }
	_ = r
}
