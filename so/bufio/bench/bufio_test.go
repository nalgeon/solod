package bufio_bench

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

// An onlyReader only implements io.Reader, no matter what
// other methods the underlying implementation may have.
type onlyReader struct {
	io.Reader
}

// An onlyWriter only implements io.Writer, no matter what
// other methods the underlying implementation may have.
type onlyWriter struct {
	io.Writer
}

func BenchmarkReaderBuf_Go(b *testing.B) {
	data := make([]byte, 16<<10)
	r := bytes.NewReader(data)
	b.ReportAllocs()

	for b.Loop() {
		r.Reset(data)
		src := bufio.NewReader(onlyReader{r})
		dst := onlyWriter{new(bytes.Buffer)}
		sinkInt, _ = src.WriteTo(dst)
	}
}

func BenchmarkReaderUnbuf_Go(b *testing.B) {
	data := make([]byte, 16<<10)
	src := bytes.NewReader(data)
	b.ReportAllocs()

	for b.Loop() {
		src.Reset(data)
		dst := onlyWriter{new(bytes.Buffer)}
		sinkInt, _ = src.WriteTo(dst)
	}
}

func BenchmarkWriterBuf_Go(b *testing.B) {
	data := make([]byte, 16<<10)
	r := bytes.NewReader(data)
	b.ReportAllocs()

	for b.Loop() {
		r.Reset(data)
		src := onlyReader{r}
		dstBuf := new(bytes.Buffer)
		dst := bufio.NewWriter(onlyWriter{dstBuf})
		sinkInt, _ = dst.ReadFrom(src)
	}
}

func BenchmarkWriterUnbuf_Go(b *testing.B) {
	data := make([]byte, 16<<10)
	r := bytes.NewReader(data)
	b.ReportAllocs()
	for b.Loop() {
		r.Reset(data)
		src := onlyReader{r}
		dst := new(bytes.Buffer)
		sinkInt, _ = dst.ReadFrom(src)
	}
}

func BenchmarkScanner_Go(b *testing.B) {
	const text = "Lorem ipsum dolor sit amet,\nconsectetur adipiscing elit,\nsed do eiusmod tempor incididunt ut labore\net dolore magna aliqua."
	b.ReportAllocs()
	for b.Loop() {
		r := strings.NewReader(text)
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			sinkStr = sc.Text()
		}
		if err := sc.Err(); err != nil {
			b.Fatal(err)
		}
	}
}
