package main

// Primitive types.
type ID int         // same type in C
type AlsoID ID      // also int
type AliasInt = int // also int
type AliasID = ID   // same type in C
type Rune rune

// Complex types.
type Name string
type IntArray [3]int
type IntSlice []int
type IntPtr *int
type Any interface{}
type Empty struct{}

// Recursive slice types.
type Tree []Tree
type Deep [][]Deep
type Left []Right
type Right []Left

// Struct type.
type Person struct {
	name string
	age  int
}

func newPerson(name string) Person {
	p := Person{name: name}
	p.age = 42
	return p
}

// Unexported struct type.
type point struct {
	x, y int
}

// Alias for a struct type.
type Human = Person
type Employee Person

// Alias for a pointer type.
type HumanPtr = *Person

// Alias for a function type.
type Namer = func(Person) string

// Methods on aliases.
func (h *Human) Age() int {
	return h.age
}
func (aid AliasID) GetVal() int {
	return int(aid)
}
func (aid *AliasID) GetPtr() int {
	return int(*aid)
}

func personName(p Person) string {
	return p.name
}

// Inner struct.
type Benchmark struct {
	name string
	loop struct {
		n int
		i int
	}
}

type empty struct{}

// Struct with an empty first field.
type tagged struct {
	e empty
	n int
}

func main() {
	{
		// Primitive types.
		var id ID = 123
		_ = id

		var aid AliasID = 456
		_ = aid

		var alsoID AlsoID = 789
		_ = alsoID

		var r Rune = 'A'
		_ = r
	}
	{
		// Complex types.
		var n Name = "Alice"
		_ = n

		var arr IntArray = [3]int{1, 2, 3}
		_ = arr

		var slice IntSlice = []int{4, 5, 6}
		_ = slice
	}
	{
		// Recursive slice types.
		var leaf Tree
		tree := make(Tree, 0, 1)
		tree = append(tree, leaf)
		if len(tree) != 1 || len(tree[0]) != 0 {
			panic("tree is not a single empty node")
		}

		var deep Deep
		if len(deep) != 0 {
			panic("len(deep) != 0")
		}

		var right Right
		left := make(Left, 0, 1)
		left = append(left, right)
		if len(left) != 1 {
			panic("len(left) != 1")
		}
	}
	{
		// Struct types.
		bob := Person{"Bob", 20}
		_ = bob

		alice := Person{name: "Alice", age: 30}
		_ = alice

		fred := Person{name: "Fred"}
		_ = fred

		ann := &Person{name: "Ann", age: 40}
		*ann = newPerson("Jon")
		_ = ann

		var sean Person
		sean.name = "Sean"
		sean.age = 50
		sp := &sean
		sp.age = 51
		_ = sean
	}
	{
		// Empty struct types.
		var e empty
		_ = e

		ep := new(empty)
		_ = ep

		var earr [3]empty
		_ = earr

		var tag tagged
		tag.n = 60
		if tag.n != 60 {
			panic("tag.n != 60")
		}
	}
	{
		// Anonymous struct type.
		dog := struct {
			name   string
			isGood bool
		}{
			"Rex",
			true,
		}
		_ = dog

		// Anonymous struct type without an initializer.
		var cat struct {
			name   string
			isGood bool
		}
		cat.name = "Tom"
		if cat.isGood {
			panic("cat.isGood")
		}

		var a, b struct{ x int }
		a.x = 1
		b.x = 2
		if a.x+b.x != 3 {
			panic("a.x+b.x != 3")
		}

		// Anonymous struct literal without values.
		zero := struct {
			name   string
			isGood bool
		}{}
		if zero.name != "" || zero.isGood {
			panic("zero is not zeroed")
		}

		// Anonymous struct literal without fields.
		unit := struct{}{}
		_ = unit
	}
	{
		// Named struct type inside a function.
		type Point struct {
			x, y int
		}
		p := Point{1, 2}
		_ = p
	}
	{
		// Inner struct.
		b1 := Benchmark{name: "Test"}
		b1.loop.n = 100
		if b1.loop.n != 100 {
			panic("b1.loop.n != 100")
		}
		b2 := Benchmark{name: "Test2", loop: struct{ n, i int }{n: 200, i: 10}}
		if b2.loop.n != 200 {
			panic("b2.loop.n != 200")
		}
		b3 := Benchmark{name: "Test3", loop: struct{ n, i int }{300, 30}}
		if b3.loop.n != 300 {
			panic("b3.loop.n != 300")
		}
		var b4 Benchmark
		if b4.loop.n != 0 {
			panic("b4.loop.n != 0")
		}
	}
	{
		// Type aliases.
		h := Human{name: "Alice", age: 30}
		age := h.Age()
		if age != 30 {
			panic("h.Age() != 30")
		}
		aid := AliasID(123)
		if aid.GetVal() != 123 {
			panic("aid.GetVal() != 123")
		}
		if aid.GetPtr() != 123 {
			panic("aid.GetPtr() != 123")
		}
		var id ID = aid
		if id.GetVal() != 123 {
			panic("id.GetVal() != 123")
		}
	}
	{
		// Alias for a pointer type: the method call unaliases the receiver.
		var hp HumanPtr = &Person{name: "Zoe", age: 60}
		if hp.Age() != 60 {
			panic("hp.Age() != 60")
		}
	}
	{
		// Alias for a function type.
		var namer Namer = personName
		if namer(Person{name: "Ivy"}) != "Ivy" {
			panic("namer(Ivy) != Ivy")
		}
	}
	{
		// Conversion between structs with the same underlying type.
		p := Person{name: "Nina", age: 25}
		e := Employee(p)
		if e.name != "Nina" || e.age != 25 {
			panic("Employee(p) lost the fields")
		}
		back := Person(e)
		if back.name != "Nina" || back.age != 25 {
			panic("Person(e) lost the fields")
		}
	}
}
