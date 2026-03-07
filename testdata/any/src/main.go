package main

type point struct {
	x, y int
}

func acceptAny(v any) {
	_ = v
}

func acceptByte(v *byte) {
	_ = v
}

func acceptPoint(v *point) {
	_ = v
}

func main() {
	{
		// Nil value.
		var n any
		acceptAny(n)
		acceptAny(any(n))
	}
	{
		// Integer value.
		n := 42
		acceptAny(n)
		acceptAny(any(n))
		acceptByte(any(n).(*byte))
	}
	{
		// Integer pointer.
		nval := 42
		n := &nval
		acceptAny(n)
		acceptAny(any(n))
		acceptByte(any(n).(*byte))
	}
	{
		// String value.
		s := "hello"
		acceptAny(s)
		acceptAny(any(s))
		acceptByte(any(s).(*byte))
	}
	{
		// String pointer.
		sval := "hello"
		s := &sval
		acceptAny(s)
		acceptAny(any(s))
		acceptByte(any(s).(*byte))
	}
	{
		// Slice value.
		s := []int{1, 2, 3}
		acceptAny(s)
		acceptAny(any(s))
		acceptByte(any(s).(*byte))
	}
	{
		// Slice pointer.
		sval := []int{1, 2, 3}
		s := &sval
		acceptAny(s)
		acceptAny(any(s))
		acceptByte(any(s).(*byte))
	}
	{
		// Struct value.
		p := point{1, 2}
		acceptAny(p)
		acceptAny(any(p))
		acceptPoint(any(p).(*point))
	}
	{
		// Struct pointer.
		pval := point{1, 2}
		p := &pval
		acceptAny(p)
		acceptAny(any(p))
		acceptPoint(any(p).(*point))
	}
	{
		// Any casts.
		n := 42
		var a any = n
		b := a.(*byte)
		if *b != 42 {
			panic("want *b == 42")
		}
		s1 := "hello"
		a = &s1
		s2 := a.(*string)
		if s2 != &s1 {
			panic("want s2 == s1")
		}
	}
}
