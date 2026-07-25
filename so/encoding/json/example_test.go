package json_test

import (
	"solod.dev/so/encoding/json"
	"solod.dev/so/fmt"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
)

type Money struct {
	Currency string
	Amount   int64
}

func (m *Money) EncodeJSON(enc *json.Encoder) {
	enc.BeginObject()
	enc.Str("currency")
	enc.Str(m.Currency)
	enc.Str("amount")
	enc.Int(m.Amount)
	enc.EndObject()
}

func (m *Money) DecodeJSON(alloc mem.Allocator, dec *json.Decoder) error {
	if dec.Kind() != json.KindObjBeg {
		dec.Skip()
		return json.ErrKind
	}
	for dec.Next() && dec.Kind() == json.KindString {
		switch dec.Str() {
		case "currency":
			dec.Next()
			m.Currency = strings.Clone(alloc, dec.Str())
		case "amount":
			dec.Next()
			m.Amount = dec.Int()
		default:
			dec.Next()
			dec.Skip()
		}
	}
	return dec.Err()
}

func (m *Money) Free(alloc mem.Allocator) {
	mem.FreeString(alloc, m.Currency)
	*m = Money{}
}

type Person struct {
	Name    string
	Balance Money
}

// EncodeJSON writes p as a JSON object.
func (p *Person) EncodeJSON(enc *json.Encoder) {
	// Encoding is done through a series of calls
	// that match the structure of the document.
	enc.BeginObject()
	enc.Str("name")
	enc.Str(p.Name)
	enc.Str("balance")
	// A nested value encodes itself into the same encoder.
	p.Balance.EncodeJSON(enc)
	enc.EndObject()
}

// DecodeJSON decodes p from dec, allocating the fields p owns with alloc.
// It starts on the opening { and stops on the matching }.
// Release the fields with [Person.Free] and the same allocator.
func (p *Person) DecodeJSON(alloc mem.Allocator, dec *json.Decoder) error {
	if dec.Kind() != json.KindObjBeg {
		// Not an object: consume the whole value so the caller
		// stays in sync, then report that nothing was decoded.
		dec.Skip()
		return json.ErrKind
	}
	// Decoding pulls one validated token at a time with Next.
	// Check each token with Kind and read its value with a typed
	// getter like Str, Int, or Bool.
	for dec.Next() && dec.Kind() == json.KindString {
		switch dec.Str() {
		case "name":
			dec.Next()
			// Str is a view into the decoder's buffer, valid only until
			// the next token, so an owned field must copy it.
			p.Name = strings.Clone(alloc, dec.Str())
		case "balance":
			dec.Next()
			// The nested value decodes itself, from the same allocator.
			p.Balance.DecodeJSON(alloc, dec)
		default:
			dec.Next()
			dec.Skip()
		}
	}
	return dec.Err()
}

// Free releases the memory p owns. It does not free p itself.
func (p *Person) Free(alloc mem.Allocator) {
	mem.FreeString(alloc, p.Name)
	p.Balance.Free(alloc)
	// Zeroing makes Free safe to call more than once.
	*p = Person{}
}

func ExampleEncoder() {
	out := make([]byte, 256)
	sb := strings.FixedBuilder(out)
	enc := json.NewEncoder(&sb)

	p := &Person{Name: "Alice", Balance: Money{Currency: "USD", Amount: 100}}
	p.EncodeJSON(&enc)
	enc.Flush()

	fmt.Println(sb.String())
	// Output: {"name":"Alice","balance":{"currency":"USD","amount":100}}
}

func ExampleDecoder() {
	src := `{"name":"Alice","balance":{"currency":"USD","amount":100}}`
	dec := json.NewDecoder(mem.System, []byte(src))
	defer dec.Free()

	var p Person
	// A failed decode may still have allocated some fields,
	// so p is freed either way.
	defer p.Free(mem.System)

	dec.Next() // the opening {
	if err := p.DecodeJSON(mem.System, &dec); err != nil {
		fmt.Println("decode failed")
		return
	}

	fmt.Printf("%s %s %d\n", p.Name, p.Balance.Currency, int(p.Balance.Amount))
	// Output: Alice USD 100
}
