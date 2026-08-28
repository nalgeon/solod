package main

func main() {
	j := 3
	for i := 0; i < j; i, j = i+1, j-1 {
		println(i)
	}
}
