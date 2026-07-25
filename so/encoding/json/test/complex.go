package main

import (
	"solod.dev/so/encoding/json"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

type money struct {
	currency string
	amount   int64
}

type person struct {
	name    string
	age     int64
	balance money
}

func (p *person) decodeJSON(alloc mem.Allocator, dec *json.Decoder) error {
	if dec.Kind() != json.KindObjBeg {
		// Not an object: consume the whole value so the caller stays in sync,
		// then report that nothing was decoded.
		dec.Skip()
		return json.ErrKind
	}
	for dec.Next() && dec.Kind() == json.KindString {
		switch dec.Str() {
		case "name":
			dec.Next()
			// A Decoder.Str is a view into its buffer, valid only until
			// the next token, so we must allocate a copy of the string
			// for the person struct to own.
			p.name = strings.Clone(alloc, dec.Str())
		case "age":
			dec.Next()
			p.age = dec.Int()
		case "balance":
			dec.Next()
			var m money
			if err := m.decodeJSON(alloc, dec); err == nil {
				p.balance = m
			}
		default:
			dec.Next()
			dec.Skip()
		}
	}
	return dec.Err()
}

func (m *money) decodeJSON(alloc mem.Allocator, dec *json.Decoder) error {
	if dec.Kind() != json.KindObjBeg {
		dec.Skip()
		return json.ErrKind
	}
	for dec.Next() && dec.Kind() == json.KindString {
		switch dec.Str() {
		case "currency":
			dec.Next()
			m.currency = strings.Clone(alloc, dec.Str())
		case "amount":
			dec.Next()
			m.amount = dec.Int()
		default:
			dec.Next()
			dec.Skip()
		}
	}
	return dec.Err()
}

func TestDecodeCollection(t *testing.T) {
	// An easy way to decode a collection of objects is to use an arena,
	// so the entire memory allocated during decoding is released in one step,
	// with no per-object teardown.
	src := `[{"name":"Alice","age":25,"balance":{"currency":"USD","amount":100}},{"name":"Bob","age":42}]`
	dec := json.NewDecoder(mem.System, []byte(src))
	defer dec.Free()

	// The arena owns all allocations; freeing its buffer frees them all.
	buf := mem.AllocSlice[byte](mem.System, 1024, 1024)
	defer mem.FreeSlice(mem.System, buf)
	arena := mem.NewArena(buf)

	people := mem.AllocSlice[person](&arena, 0, 8)
	if !dec.Next() || dec.Kind() != json.KindArrBeg {
		t.Fatal("expected array")
		return
	}
	for dec.Next() && dec.Kind() != json.KindArrEnd {
		var p person
		err := p.decodeJSON(&arena, &dec)
		if err == json.ErrKind {
			continue // not an object; skipped, keep going
		}
		if err != nil {
			break // real decode error
		}
		people = append(people, p)
	}

	if dec.Err() != nil {
		t.Fatal("unexpected decode error")
		return
	}

	if len(people) != 2 {
		t.Fatal("expected 2 people")
		return
	}
	if people[0].name != "Alice" || people[0].age != 25 {
		t.Error("Alice mismatch")
	}
	if people[0].balance.currency != "USD" || people[0].balance.amount != 100 {
		t.Error("Alice balance mismatch")
	}
	if people[1].name != "Bob" || people[1].age != 42 {
		t.Error("Bob mismatch")
	}
}
