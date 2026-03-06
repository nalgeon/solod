package main

type Person struct {
	Name string
	Age  int
	Nums []int
}

func (p *Person) Sleep() int {
	p.Age += 1
	return p.Age
}

func main() {
	p := Person{Name: "Alice", Age: 30}
	p.Sleep()
	println(p.Name, "is now", p.Age, "years old.")

	p.Nums = make([]int, 0, 4)
	p.Nums = append(p.Nums, 7, 42, 13)
	println("1st lucky number is", p.Nums[0])
}
