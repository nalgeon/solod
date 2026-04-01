# Solod vs. Go benchmarks

Here are some benchmarks that show how So performs on common tasks compared to Go.

## Byte functions

So is generally ~1.5x faster than Go, except for Index operations.
Memory usage is the same for both.

| Benchmark  |    Go | So (mimalloc) | So (arena) | Winner        |
| ---------- | ----: | ------------: | ---------: | ------------- |
| Clone      | 102ns |          41ns |       32ns | **So** - 2.5x |
| Compare    |  34ns |          25ns |       25ns | **So** - 1.4x |
| Index      |  21ns |          32ns |       32ns | Go - 0.7x     |
| IndexByte  |  16ns |          25ns |       25ns | Go - 0.6x     |
| Repeat     | 106ns |          56ns |       48ns | **So** - 1.9x |
| ReplaceAll | 247ns |         258ns |      242ns | ~same         |
| Split      | 510ns |         422ns |      421ns | **So** - 1.2x |
| ToUpper    | 322ns |         176ns |      171ns | **So** - 1.8x |
| Trim       |  47ns |          44ns |       44ns | **So** - 1.1x |
| TrimSuffix |   4ns |           2ns |        2ns | **So** - 1.8x |

Apple M1 • Go 1.26.1 • [details](./bytes/README.md#functions)

## Byte buffer

So reads 1.3x faster and writes 2-4x faster than Go.
Memory usage is the same for both.

| Benchmark  |      Go | So (mimalloc) | So (arena) | Winner        |
| ---------- | ------: | ------------: | ---------: | ------------- |
| ReadString |  2329ns |        1757ns |     1719ns | **So** - 1.3x |
| WriteByte  |  8858ns |        2608ns |     2643ns | **So** - 3.4x |
| WriteRune  | 15110ns |        3902ns |     3956ns | **So** - 3.8x |
| WriteBlock | 17238ns |        7830ns |     7510ns | **So** - 2.2x |

Apple M1 • Go 1.26.1 • [details](./bytes/README.md#buffer)

## String functions

So is generally ~1.3x faster than Go, except for Index operations.
Memory usage is the same for both.

| Benchmark  |     Go | So (mimalloc) | So (arena) | Winner        |
| ---------- | -----: | ------------: | ---------: | ------------- |
| Clone      |   99ns |          42ns |       34ns | **So** - 2.4x |
| Compare    |   47ns |          36ns |       36ns | **So** - 1.3x |
| Fields     | 1524ns |         908ns |      912ns | **So** - 1.7x |
| Index      |   25ns |          35ns |       34ns | Go - 0.7x     |
| IndexByte  |   22ns |          33ns |       33ns | Go - 0.7x     |
| Repeat     |  127ns |          64ns |       67ns | **So** - 1.9x |
| ReplaceAll |  243ns |         200ns |      203ns | **So** - 1.2x |
| Split      | 1899ns |        1399ns |     1423ns | **So** - 1.3x |
| ToUpper    | 2066ns |        1602ns |     1622ns | **So** - 1.3x |
| Trim       |  501ns |         373ns |      375ns | **So** - 1.3x |

Apple M1 • Go 1.26.1 • [details](./strings/README.md#functions)

## String builder

So is 2-4x faster than Go and uses 10%-20% less memory.

| Benchmark                |    Go | So (mimalloc) | So (arena) | Winner        |
| ------------------------ | ----: | ------------: | ---------: | ------------- |
| Write bytes (auto-grow)  | 245ns |         118ns |       59ns | **So** - 2.1x |
| Write bytes (pre-grow)   | 109ns |          29ns |       25ns | **So** - 3.8x |
| Write string (auto-grow) | 224ns |         116ns |       57ns | **So** - 1.9x |
| Write string (pre-grow)  | 113ns |          29ns |       26ns | **So** - 3.9x |

Apple M1 • Go 1.26.1 • [details](./strings/README.md#builder)

## Methodology

So is compiled with `-Ofast -march=native -flto -funroll-loops` and uses mimalloc as the system allocator. Go is run with default `go test -bench=.` settings.

The Winner column shows the worse result between mimalloc and arena for each So benchmark.
