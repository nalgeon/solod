// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package uuid

import "testing"

func FuzzParse(f *testing.F) {
	f.Add("f81d4fae-7dec-11d0-a765-00a0c91e6bf6")
	f.Add("F81D4FAE-7DEC-11D0-A765-00A0C91E6BF6")
	f.Add("f81d4fae7dec11d0a76500a0c91e6bf6")
	f.Add("{f81d4fae-7dec-11d0-a765-00a0c91e6bf6}")
	f.Add("urn:uuid:f81d4fae-7dec-11d0-a765-00a0c91e6bf6")
	f.Add("00000000-0000-0000-0000-000000000000")
	f.Add("ffffffff-ffff-ffff-ffff-ffffffffffff")
	f.Add("")
	f.Add("0000-0000-0000-0000-0000-0000-0000-0000")

	f.Fuzz(func(t *testing.T, s string) {
		u, err := Parse(s)
		if err != nil {
			return
		}
		// Parse must accept the string that String writes for an accepted
		// UUID, and must return the same UUID for it.
		buf := make([]byte, UUIDLen)
		got, err := Parse(u.String(buf))
		if err != nil {
			t.Fatalf("Parse(%q) = %v; the round trip failed: %v", s, u, err)
		}
		if got != u {
			t.Fatalf("Parse(%q) = %v; the round trip gave %v", s, u, got)
		}
	})
}
