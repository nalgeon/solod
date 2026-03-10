package main

func main() {
	{
		// Integer literals.
		const d1 = 123
		_ = d1
		const d2 = 100_000
		_ = d2
		const d3 = 0b1010
		_ = d3
		const d4 = 0o600
		_ = d4
		const d5 = 0xBadFace
		_ = d5
		const d6 = 0x_67_7a_2f_cc_40_c6
		_ = d6
	}

	{
		// Floating-point literals.
		const f1 = 3.14
		_ = f1
		const f2 = 0.25
		_ = f2
		const f3 = 1e-9
		_ = f3
		const f4 = 6.022e23
		_ = f4
		const f5 = 1e6
		_ = f5
	}

	// {
	// 	// Imaginary literals - not supported.
	// 	const i1 = 0i
	// 	_ = i1
	// 	const i2 = 0o123i // == 0o123 * 1i == 83i
	// 	_ = i2
	// 	const i3 = 0xabci // == 0xabc * 1i == 2748i
	// 	_ = i3
	// 	const i4 = 2.71828i
	// 	_ = i4
	// 	const i5 = 1.e+0i
	// }

	{
		// Rune literals.
		const r1 = 'a'
		_ = r1
		const r2 = 'ä'
		_ = r2
		const r3 = '本'
		_ = r3
		const r4 = '\xff'
		_ = r4
		const r5 = '\u12e4'
		_ = r5
	}

	{
		// String literals.
		const s1 = "abc"
		_ = s1
		const s2 = `abc
		def`
		_ = s2
		const s3 = "\n"
		_ = s3
		const s4 = "日本語"
		_ = s4
		const s5 = "\xff\u00FF"
		_ = s5
	}

	{
		// Conversions.
		const x uint = 123
		const n1 = int(x)
		_ = n1
		const n2 = int(x & 7)
		_ = n2

		const mask2 = 0b00011111
		var p0 byte = 'x'
		r := rune(p0 & mask2)
		_ = r
	}
}
