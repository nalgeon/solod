# bytes benchmarks

Requires GCC/Clang and mimalloc (for heap allocations in So). If mimalloc isn't available, the benchmarks will use the default libc allocator, which is much slower.

## Functions

Run the benchmark:

```text
make bench name=bytes
```

Go 1.26.1:

```text
goos: darwin
goarch: arm64
pkg: solod.dev/bench/bytes
cpu: Apple M1
Benchmark_Clone-8          11691513     102.1 ns/op     1024 B/op    1 allocs/op
Benchmark_Compare-8        35754898      34.11 ns/op       0 B/op    0 allocs/op
Benchmark_Index-8          57072535      21.00 ns/op       0 B/op    0 allocs/op
Benchmark_IndexByte-8      77301531      15.66 ns/op       0 B/op    0 allocs/op
Benchmark_Repeat-8         11228131     106.0 ns/op     1024 B/op    1 allocs/op
Benchmark_ReplaceAll-8      4901722     246.9 ns/op      288 B/op    1 allocs/op
Benchmark_Split-8           2355016     509.6 ns/op      416 B/op    1 allocs/op
Benchmark_ToUpper-8         3729416     321.6 ns/op      288 B/op    1 allocs/op
Benchmark_Trim-8           25322198     47.16 ns/op        0 B/op    0 allocs/op
Benchmark_TrimSuffix-8    302613986      3.969 ns/op       0 B/op    0 allocs/op
```

So (mimalloc):

```text
Benchmark_Clone            28935634     40.64 ns/op     1024 B/op    1 allocs/op
Benchmark_Compare          47212495     25.28 ns/op        0 B/op    0 allocs/op
Benchmark_Index            36508563     32.41 ns/op        0 B/op    0 allocs/op
Benchmark_IndexByte        47683381     25.09 ns/op        0 B/op    0 allocs/op
Benchmark_Repeat           21348134     56.05 ns/op     1024 B/op    1 allocs/op
Benchmark_ReplaceAll        4014303    257.9 ns/op       272 B/op    1 allocs/op
Benchmark_Split             2841338    421.8 ns/op       408 B/op    1 allocs/op
Benchmark_ToUpper           6814851    176.1 ns/op       272 B/op    1 allocs/op
Benchmark_Trim             27164685     44.07 ns/op        0 B/op    0 allocs/op
Benchmark_TrimSuffix      537721137      2.206 ns/op       0 B/op    0 allocs/op
```

So (arena):

```text
Benchmark_Clone            37411148     32.25 ns/op     1024 B/op    1 allocs/op
Benchmark_Compare          46719874     25.19 ns/op        0 B/op    0 allocs/op
Benchmark_Index            36524120     32.37 ns/op        0 B/op    0 allocs/op
Benchmark_IndexByte        47020100     25.33 ns/op        0 B/op    0 allocs/op
Benchmark_Repeat           25454467     47.64 ns/op     1024 B/op    1 allocs/op
Benchmark_ReplaceAll        4765062    241.9 ns/op       272 B/op    1 allocs/op
Benchmark_Split             2870538    421.0 ns/op       408 B/op    1 allocs/op
Benchmark_ToUpper           7046222    170.9 ns/op       272 B/op    1 allocs/op
Benchmark_Trim             27020941     44.31 ns/op        0 B/op    0 allocs/op
Benchmark_TrimSuffix      547245530      2.192 ns/op       0 B/op    0 allocs/op
```

## Buffer

Go 1.26.1:

```text
goos: darwin
goarch: arm64
pkg: solod.dev/bench/bytes/buffer
cpu: Apple M1
BenchmarkReadString-8    504800     2329 ns/op    14070.93 MB/s
BenchmarkWriteByte-8     135392     8858 ns/op      462.41 MB/s
BenchmarkWriteRune-8      79524    15110 ns/op      813.23 MB/s
BenchmarkWriteBlock-8     69115    17238 ns/op      261121 B/op  8 allocs/op
```

So (mimalloc):

```text
Benchmark_ReadString     683119     1757 ns/op    18652.40 MB/s
Benchmark_WriteByte      463195     2608 ns/op     1570.44 MB/s
Benchmark_WriteRune      304737     3902 ns/op     3149.49 MB/s
Benchmark_WriteBlock     142560     7830 ns/op      261120 B/op  8 allocs/op
```

So (arena):

```text
Benchmark_ReadString     895989     1719 ns/op    19061.32 MB/s
Benchmark_WriteByte      455563     2643 ns/op     1549.64 MB/s
Benchmark_WriteRune      297169     3956 ns/op     3106.25 MB/s
Benchmark_WriteBlock     156660     7510 ns/op      261120 B/op  8 allocs/op
```
