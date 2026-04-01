# strings benchmarks

Requires GCC/Clang and mimalloc (for heap allocations in So). If mimalloc isn't available, the benchmarks will use the default libc allocator, which is much slower.

## Functions

Run the benchmark:

```text
make bench name=strings
```

Go 1.26.1:

```text
goos: darwin
goarch: arm64
pkg: solod.dev/bench/strings
cpu: Apple M1
Benchmark_Clone-8          12143073      98.50 ns/op    1024 B/op    1 allocs/op
Benchmark_Compare-8        25180718      47.03 ns/op       0 B/op    0 allocs/op
Benchmark_Fields-8           791077    1524 ns/op        288 B/op    1 allocs/op
Benchmark_Index-8          47874408      25.14 ns/op       0 B/op    0 allocs/op
Benchmark_IndexByte-8      54982188      21.98 ns/op       0 B/op    0 allocs/op
Benchmark_Repeat-8          9197040     127.3 ns/op     1024 B/op    1 allocs/op
Benchmark_ReplaceAll-8      5164438     242.8 ns/op      256 B/op    1 allocs/op
Benchmark_Split-8            631638    1899 ns/op        288 B/op    1 allocs/op
Benchmark_ToUpper-8          581395    2066 ns/op        384 B/op    1 allocs/op
Benchmark_Trim-8            2519773     501.0 ns/op        0 B/op    0 allocs/op
Benchmark_TrimSuffix-8    313593835       3.800 ns/op      0 B/op    0 allocs/op
```

So (mimalloc):

```text
Benchmark_Clone            27935466      41.84 ns/op    1024 B/op    1 allocs/op
Benchmark_Compare         473186119      35.76 ns/op       0 B/op    0 allocs/op
Benchmark_Fields            1319384     907.7 ns/op      272 B/op    1 allocs/op
Benchmark_Index            33552540      35.21 ns/op       0 B/op    0 allocs/op
Benchmark_IndexByte        36868624      32.81 ns/op       0 B/op    0 allocs/op
Benchmark_Repeat           18445929      64.11 ns/op    1024 B/op    1 allocs/op
Benchmark_ReplaceAll        5952439     200.2 ns/op      256 B/op    1 allocs/op
Benchmark_Split              852272    1399 ns/op        272 B/op    1 allocs/op
Benchmark_ToUpper            738597    1602 ns/op        372 B/op    1 allocs/op
Benchmark_Trim              3214159     373.0 ns/op        0 B/op    0 allocs/op
Benchmark_TrimSuffix      572874397       2.126 ns/op      0 B/op    0 allocs/op
```

So (arena):

```text
Benchmark_Clone            35269221      33.92 ns/op    1024 B/op    1 allocs/op
Benchmark_Compare          29562475      36.16 ns/op       0 B/op    0 allocs/op
Benchmark_Fields            1320866     911.5 ns/op      272 B/op    1 allocs/op
Benchmark_Index            35494556      33.85 ns/op       0 B/op    0 allocs/op
Benchmark_IndexByte        36586480      32.57 ns/op       0 B/op    0 allocs/op
Benchmark_Repeat           18169980      66.64 ns/op    1024 B/op    1 allocs/op
Benchmark_ReplaceAll        6082384     202.5 ns/op      256 B/op    1 allocs/op
Benchmark_Split              839042    1423 ns/op        272 B/op    1 allocs/op
Benchmark_ToUpper            739098    1622 ns/op        372 B/op    1 allocs/op
Benchmark_Trim              3222462     375.1 ns/op        0 B/op    0 allocs/op
Benchmark_TrimSuffix      586699521       2.043 ns/op      0 B/op    0 allocs/op
```

## Builder

Run the benchmark:

```text
make bench name=strings/builder
```

Go 1.26.1:

```text
goos: darwin
goarch: arm64
pkg: solod.dev/bench/strings/builder
cpu: Apple M1

Benchmark_WriteB_AutoGrow-8    4791342       245.3 ns/op     1424 B/op    5 allocs/op
Benchmark_WriteB_PreGrow-8    10950619       108.5 ns/op      640 B/op    1 allocs/op
Benchmark_WriteS_AutoGrow-8    5385492       224.0 ns/op     1424 B/op    5 allocs/op
Benchmark_WriteS_PreGrow-8    10692721       112.9 ns/op      640 B/op    1 allocs/op
```

So (mimalloc):

```text
Benchmark_WriteB_AutoGrow     10068822       118.3 ns/op     1147 B/op    5 allocs/op
Benchmark_WriteB_PreGrow      41877507        28.69 ns/op     592 B/op    1 allocs/op
Benchmark_WriteS_AutoGrow     10344024       115.9 ns/op     1147 B/op    5 allocs/op
Benchmark_WriteS_PreGrow      41045286        28.74 ns/op     592 B/op    1 allocs/op
```

So (arena):

```text
Benchmark_WriteB_AutoGrow     19914367        58.62 ns/op    1147 B/op    5 allocs/op
Benchmark_WriteB_PreGrow      47978888        25.28 ns/op     592 B/op    1 allocs/op
Benchmark_WriteS_AutoGrow     21232549        56.69 ns/op    1147 B/op    5 allocs/op
Benchmark_WriteS_PreGrow      46214280        25.84 ns/op     592 B/op    1 allocs/op
```
