package main

func main() {
	var x struct{ a int }
	var y struct{ a int }
	y = x
	_ = y
}
