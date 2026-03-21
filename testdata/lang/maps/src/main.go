package main

type Pair struct {
	x int
	y int
}

type IntFunc func() int

type StrMap map[string]int

func takeMap(m map[string]int) int {
	return m["a"] + m["b"]
}

func modifyMap(m map[string]int) {
	m["a"] = 99
	m["b"] = 22
}

type MapHolder struct {
	m map[string]int
}

func (h MapHolder) get(key string) int {
	return h.m[key]
}

func (h MapHolder) sum() int {
	s := 0
	for _, v := range h.m {
		s += v
	}
	return s
}

func ten() int    { return 10 }
func twenty() int { return 20 }

func main() {
	{
		// Key type: string (eq_str)
		m := map[string]int{"a": 1, "b": 2}
		if m["a"] != 1 || m["b"] != 2 {
			panic("string key")
		}
	}
	{
		// Key type: int (eq_8)
		m := map[int]int{1: 10, 2: 20}
		if m[1] != 10 || m[2] != 20 {
			panic("int key")
		}
	}
	{
		// Key type: float32 (eq_4)
		m := map[float32]int{1.5: 10, 2.5: 20}
		if m[1.5] != 10 || m[2.5] != 20 {
			panic("float32 key")
		}
	}
	{
		// Key type: uint16 (eq_2)
		m := map[uint16]int{1: 10, 2: 20}
		if m[1] != 10 || m[2] != 20 {
			panic("uint16 key")
		}
	}
	{
		// Key type: bool (eq_1)
		m := map[bool]int{true: 1, false: 0}
		if m[true] != 1 || m[false] != 0 {
			panic("bool key")
		}
	}
	{
		// Key type: uint8 (eq_1)
		m := map[uint8]int{1: 10, 2: 20}
		if m[1] != 10 || m[2] != 20 {
			panic("uint8 key")
		}
	}
	{
		// Key type: *int (pointer, eq_8)
		a := 1
		b := 2
		m := map[*int]string{&a: "first", &b: "second"}
		if m[&a] != "first" || m[&b] != "second" {
			panic("pointer key")
		}
	}
	{
		// Value type: string
		m := map[int]string{1: "a", 2: "b"}
		if m[1] != "a" || m[2] != "b" {
			panic("string value")
		}
	}
	{
		// Value type: bool
		m := map[string]bool{"yes": true, "no": false}
		if !m["yes"] || m["no"] {
			panic("bool value")
		}
	}
	{
		// Value type: float64
		m := map[string]float64{"pi": 3.14, "e": 2.71}
		if m["pi"] != 3.14 || m["e"] != 2.71 {
			panic("float64 value")
		}
	}
	{
		// Key and value: string
		m := map[string]string{"a": "x", "b": "y"}
		if m["a"] != "x" || m["b"] != "y" {
			panic("string string")
		}
	}
	{
		// Value type: struct
		m := map[string]Pair{"origin": Pair{0, 0}, "point": Pair{3, 4}}
		if m["origin"].x != 0 || m["point"].x != 3 || m["point"].y != 4 {
			panic("struct value")
		}
	}
	{
		// Value type: slice
		s1 := make([]int, 2)
		s1[0] = 10
		s1[1] = 20
		s2 := make([]int, 2)
		s2[0] = 30
		s2[1] = 40
		m := map[string][]int{"a": s1, "b": s2}
		if m["a"][0] != 10 || m["b"][1] != 40 {
			panic("slice value")
		}
	}
	{
		// Value type: *int (pointer)
		a := 42
		b := 99
		m := map[string]*int{"a": &a, "b": &b}
		if *m["a"] != 42 || *m["b"] != 99 {
			panic("pointer value")
		}
	}
	{
		// Value type: map (nested)
		inner1 := map[string]int{"x": 1}
		inner2 := map[string]int{"y": 2}
		m := map[string]map[string]int{"a": inner1, "b": inner2}
		if m["a"]["x"] != 1 || m["b"]["y"] != 2 {
			panic("nested map value")
		}
	}
	{
		// Value type: function
		m := map[string]IntFunc{
			"ten":    ten,
			"twenty": twenty,
		}
		if m["ten"]() != 10 || m["twenty"]() != 20 {
			panic("func value")
		}
	}
	{
		// Single element literal
		m := map[string]int{"only": 42}
		if m["only"] != 42 {
			panic("single element")
		}
	}
	{
		// Empty literal
		m := map[string]int{}
		if len(m) != 0 {
			panic("empty literal")
		}
	}
	{
		// Make and populate
		m := make(map[string]int, 3)
		if len(m) != 0 {
			panic("make initial len")
		}
		m["a"] = 10
		m["b"] = 20
		m["c"] = 30
		if m["a"] != 10 || m["b"] != 20 || m["c"] != 30 {
			panic("make values")
		}
		if len(m) != 3 {
			panic("make final len")
		}
	}
	{
		// Make with int key
		m := make(map[int]string, 2)
		m[1] = "one"
		m[2] = "two"
		if m[1] != "one" || m[2] != "two" {
			panic("make int key")
		}
	}
	{
		// Missing key: zero value int
		m := map[string]int{"a": 1}
		if m["missing"] != 0 {
			panic("zero int")
		}
	}
	{
		// Missing key: zero value string
		m := map[int]string{1: "a"}
		if m[99] != "" {
			panic("zero string")
		}
	}
	{
		// Missing key: zero value bool
		m := map[string]bool{"a": true}
		if m["missing"] {
			panic("zero bool")
		}
	}
	{
		// Missing key: zero value struct
		m := map[string]Pair{"a": Pair{1, 2}}
		p := m["missing"]
		if p.x != 0 || p.y != 0 {
			panic("zero struct")
		}
	}
	{
		// Overwrite existing key
		m := map[string]int{"a": 1}
		m["a"] = 99
		if m["a"] != 99 {
			panic("overwrite value")
		}
		if len(m) != 1 {
			panic("overwrite len")
		}
	}
	{
		// Map value in arithmetic
		m := map[string]int{"a": 10, "b": 20}
		sum := m["a"] + m["b"]
		if sum != 30 {
			panic("arithmetic")
		}
	}
	{
		// Map value in nested expression
		m := map[string]int{"a": 2, "b": 3}
		result := m["a"]*m["b"] + m["a"]
		if result != 8 {
			panic("nested expr")
		}
	}
	{
		// Map bool value in condition
		m := map[string]bool{"flag": true}
		if !m["flag"] {
			panic("bool condition")
		}
	}
	{
		// Comma-ok: define then assign
		m := map[string]int{"a": 1, "b": 2}
		// Define, key exists.
		v, ok := m["a"]
		if !ok || v != 1 {
			panic("comma-ok define hit")
		}
		// Assign, key missing.
		v, ok = m["missing"]
		if ok || v != 0 {
			panic("comma-ok assign miss")
		}
		// Assign, key exists.
		v, ok = m["b"]
		if !ok || v != 2 {
			panic("comma-ok assign hit")
		}
	}
	{
		// Comma-ok: blank value
		m := map[string]int{"a": 1}
		_, ok := m["a"]
		if !ok {
			panic("comma-ok blank value hit")
		}
		_, ok = m["missing"]
		if ok {
			panic("comma-ok blank value miss")
		}
	}
	{
		// Comma-ok: blank ok
		m := map[string]int{"a": 1}
		v, _ := m["a"]
		if v != 1 {
			panic("comma-ok blank ok")
		}
	}
	{
		// Comma-ok: with string value
		m := map[int]string{1: "hello"}
		v, ok := m[1]
		if !ok || v != "hello" {
			panic("comma-ok string value")
		}
		v, ok = m[99]
		if ok || v != "" {
			panic("comma-ok string value miss")
		}
	}
	{
		// Range: key + value
		m := map[int]int{1: 10, 2: 20, 3: 30}
		sum := 0
		for k, v := range m {
			sum += k * v
		}
		// 1*10 + 2*20 + 3*30 = 10 + 40 + 90 = 140
		if sum != 140 {
			panic("range k,v define")
		}
	}
	{
		// Range: key only
		m := map[int]int{10: 100, 20: 200}
		sum := 0
		for k := range m {
			sum += k
		}
		if sum != 30 {
			panic("range k only")
		}
	}
	{
		// Range: value only
		m := map[int]int{10: 100, 20: 200}
		sum := 0
		for _, v := range m {
			sum += v
		}
		if sum != 300 {
			panic("range v only")
		}
	}
	{
		// Range: key + value (assign, not define)
		m := map[int]int{1: 10, 2: 20}
		k := 0
		v := 0
		sum := 0
		for k, v = range m {
			sum += k + v
		}
		// 1+10 + 2+20 = 33
		if sum != 33 {
			panic("range assign")
		}
	}
	{
		// Range: string keys and values
		m := map[string]string{"hello": "world", "foo": "bar"}
		keys := ""
		vals := ""
		for k, v := range m {
			keys += k
			vals += v
		}
		if keys != "hellofoo" && keys != "foohello" {
			panic("range string keys")
		}
		if vals != "worldbar" && vals != "barworld" {
			panic("range string values")
		}
	}
	{
		// Range: over struct values
		m := map[string]Pair{"a": Pair{1, 2}, "b": Pair{3, 4}}
		sum := 0
		for _, v := range m {
			sum += v.x + v.y
		}
		if sum != 10 {
			panic("range struct value")
		}
	}
	{
		// len: literal
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		if len(m) != 3 {
			panic("len literal")
		}
	}
	{
		// len: empty
		m := map[string]int{}
		if len(m) != 0 {
			panic("len empty")
		}
	}
	{
		// len: grows with set
		m := make(map[string]int, 3)
		if len(m) != 0 {
			panic("len make 0")
		}
		m["a"] = 1
		if len(m) != 1 {
			panic("len make 1")
		}
		m["b"] = 2
		if len(m) != 2 {
			panic("len make 2")
		}
	}
	{
		// len: overwrite does not change len
		m := map[string]int{"a": 1}
		m["a"] = 99
		if len(m) != 1 {
			panic("len overwrite")
		}
	}
	{
		// len: in expression
		m := map[string]int{"a": 1, "b": 2}
		n := len(m) + 1
		if n != 3 {
			panic("len expr")
		}
	}
	{
		// Nil: non-nil literal
		m := map[string]int{"a": 1}
		if m == nil {
			panic("non-nil")
		}
	}
	{
		// Nil: assign and check
		m := map[string]int{"a": 1}
		m = nil
		if m != nil {
			panic("nil after assign")
		}
	}
	{
		// Pass map to function
		m := map[string]int{"a": 10, "b": 20}
		if takeMap(m) != 30 {
			panic("pass to func")
		}
	}
	{
		// Function modifies map
		m := make(map[string]int, 2)
		m["a"] = 11
		modifyMap(m)
		if m["a"] != 99 {
			panic("func modify a")
		}
		if m["b"] != 22 {
			panic("func modify b")
		}
		if len(m) != 2 {
			panic("func modify len")
		}
	}
	{
		// Method on struct with map field
		h := MapHolder{m: map[string]int{"x": 10, "y": 20}}
		if h.get("x") != 10 {
			panic("method get")
		}
		if h.sum() != 30 {
			panic("method sum")
		}
	}
	{
		// Named map type
		m := StrMap{"a": 1, "b": 2}
		if m["a"] != 1 || m["b"] != 2 {
			panic("named type literal")
		}
	}
	{
		// Named map type: set and get
		m := make(StrMap, 2)
		m["x"] = 10
		m["y"] = 20
		if m["x"] != 10 || m["y"] != 20 {
			panic("named type make")
		}
	}
	{
		// Named map type: comma-ok
		m := StrMap{"a": 1}
		v, ok := m["a"]
		if !ok || v != 1 {
			panic("named type comma-ok")
		}
	}
	{
		// Named map type: range
		m := StrMap{"a": 1, "b": 2}
		sum := 0
		for _, v := range m {
			sum += v
		}
		if sum != 3 {
			panic("named type range")
		}
	}
	{
		// Map assigned to another variable
		m1 := map[string]int{"a": 1}
		m2 := m1
		m2["a"] = 99
		// In Go maps are references, m1 sees the change.
		if m1["a"] != 99 {
			panic("map alias")
		}
	}
	{
		// Map in if-init statement
		m := map[string]int{"a": 1}
		if v, ok := m["a"]; ok {
			if v != 1 {
				panic("if-init value")
			}
		}
	}
	{
		// Map in if-init statement miss
		m := map[string]int{"a": 1}
		if _, ok := m["missing"]; ok {
			panic("if-init miss")
		}
	}
	{
		// Map increment: m[key]++
		m := map[string]int{"a": 1}
		// m[key]++ and m[key] += 1 are not supported
		m["a"] = m["a"] + 1
		if m["a"] != 2 {
			panic("increment")
		}
	}
}
