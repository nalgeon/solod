# So by example

Solod (So) is a strict subset of Go that transpiles to regular C. "So by example" is a hands-on introduction to So using annotated example programs:

[Hello world](./01-hello-world/main.go) •
[Values](./02-values/main.go) •
[Variables](./03-variables/main.go) •
[Constants](./04-constants/main.go) •
[For](./05-for/main.go) •
[If/else](./06-if-else/main.go) •
[Arrays](./08-arrays/main.go) •
[Slices](./09-slices/main.go) •
[Maps](./10-maps/main.go) •
[Functions](./11-functions/main.go) •
[Multiple returns](./12-returns/main.go) •
[Variadic functions](./13-variadics/main.go) •
[Recursion](./15-recursion/main.go) •
[For-range](./16-range/main.go) •
[Pointers](./17-pointers/main.go) •
[Strings and runes](./18-strings/main.go) •
[Structs](./19-structs/main.go) •
[Methods](./20-methods/main.go) •
[Interfaces](./21-interfaces/main.go) •
[Enums](./22-enums/main.go) •
[Errors](./26-errors/main.go) •
[Panic](./27-panic/main.go) •
[Defer](./28-defer/main.go) •
[Memory](./51-memory/main.go) •
[C interop](./52-interop/main.go) •
[Strings](./53-strings/main.go) •
[Time](./54-time/main.go) •
[Files](./55-files/main.go)

To run a specific example locally, use the `so run` command. For example:

```text
so run example/05-for
```

You'll need to have a C compiler installed and available as `cc`, or you can set a custom compiler by using the `CC` environment variable.

To see the generated C code, use the `so translate` command. For example:

```text
so translate -o example/05-for/generated example/05-for
```

Based on [Go by Example](https://gobyexample.com) by Mark McGranaghan, licensed under [CC BY 3.0](https://creativecommons.org/licenses/by/3.0/).
