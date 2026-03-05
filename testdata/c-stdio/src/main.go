package main

import "github.com/nalgeon/solod/so/c/stdio"

func main() {
	{
		// Output stream.
		f := stdio.Fopen("/tmp/test.txt", "w")
		if f == nil {
			panic("failed to open file")
		}

		stdio.Fputs("hello", f)
		stdio.Fputc(10, f)
		stdio.Fflush(f)

		var buf [64]byte
		stdio.Fwrite(&buf[0], 1, 64, f)

		stdio.Fclose(f)
	}
	{
		// Input stream.
		f := stdio.Fopen("/tmp/test.txt", "r")
		if f == nil {
			panic("failed to open file")
		}

		ch := stdio.Fgetc(f)
		if ch == stdio.EOF {
			panic("unexpected EOF")
		}

		var buf [64]byte
		stdio.Fseek(f, 0, 0)
		stdio.Fgets(&buf[0], 64, f)
		stdio.Fread(&buf[0], 1, 64, f)

		if stdio.Feof(f) {
			panic("unexpected EOF")
		}
		if stdio.Ferror(f) {
			panic("stream error")
		}

		stdio.Fclose(f)
	}
	{
		// Formatted output.
		stdio.Printf("hello %d\n", 42)
		stdio.Fprintf(stdio.Stdout, "value: %d\n", 100)

		buf := make([]byte, 64)
		stdio.Snprintf(buf, 64, "count: %d", 10)
	}
	{
		// Formatted input.
		var n int32
		stdio.Sscanf("42", "%d", &n)
	}
}
