# pcom-go

[![Go Report Card](https://goreportcard.com/badge/github.com/BlackBuck/pcom-go)](https://goreportcard.com/report/github.com/BlackBuck/pcom-go)

**pcom-go** is a composable, generic **parser combinator library** written in Go, inspired by Haskell's `parsec` and Rust's `nom`.  
It allows you to write powerful parsers in a modular way with **traceable, color-coded error messages**, **generic combinators**, and **support for recursive grammars**.

> Perfect for building parsers for arithmetic expressions, config files, DSLs, or even JSON in pure Go.

---

## Features

- Type-safe parser combinators using Go 1.18+ generics
- Detailed, color-coded error messages with context and trace
- Recursion and backtracking support with `Lazy` and `Try`
- Built-in primitives for digits, letters, whitespace, etc.
- Combinators for sequences, repetition, mapping, and separation
- Streaming-ready state tracking (offset, line, column)
- Benchmarkable and fully testable

---

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/BlackBuck/pcom-go/parser"
    "github.com/BlackBuck/pcom-go/state"
)

func main() {
    input := "abc123"
    s := state.NewState(input, state.Position{Offset: 0, Line: 1, Column: 1})

    letters := parser.Many1("letters", parser.Alpha())

    res, err := letters.Run(s)
    if err.HasError() {
        fmt.Println(err.FullTrace())
        return
    }

    fmt.Printf("Parsed: %v\n", res.Value)
}
````

---

## Core Concepts

A parser is a function that transforms a `state.State` into a `Result[T]` or an `Error`.

```go
type Parser[T any] struct {
	Run   func(State) (Result[T], Error)
	Label string
}
```

---

## 🛠️ API Overview

### Primitives

| Function               | Description                           |
| ---------------------- | ------------------------------------- |
| `RuneParser("x", 'x')` | Parses a single rune                  |
| `StringParser("let")`  | Parses exact string                   |
| `Digit()`              | Parses one digit                      |
| `Alpha()`              | Parses one letter                     |
| `AlphaNum()`           | Parses letter or digit                |
| `Whitespace()`         | Matches a space character             |
| `OneOf("+-*/")`        | Matches one of the listed characters  |
| `StringCI("hello")`    | Case-insensitive string match         |
| `CharWhere(fn, lbl)`   | Parses rune based on custom predicate |
| `AnyChar()`            | Matches any rune                      |
| `TakeWhile(f)`         | Consumes while predicate holds        |

---

### Combinators

| Function                  | Description                                   |
| ------------------------- | --------------------------------------------- |
| `Or(p1, p2, ...)`         | Try multiple parsers, return first successful |
| `And(p1, p2)`             | Sequence parsers, return last by default      |
| `Sequence([]p)`           | General sequence, returns last result         |
| `Then(p1, p2)`            | Combine results into a `Pair[A, B]`           |
| `KeepLeft(Then(p1, p2))`  | Return left result only                       |
| `KeepRight(Then(p1, p2))` | Return right result only                      |
| `Map(p, f)`               | Transform result using function               |
| `Optional(p)`             | Zero-or-one occurrence, never errors          |
| `Many0(p)`                | Zero or more repetitions                      |
| `Many1(p)`                | One or more repetitions                       |
| `Between(open, p, close)` | Matches `open`, `p`, and `close`              |
| `SeparatedBy(p, sep)`     | Parses values separated by a delimiter        |
| `ManyTill(p, end)`        | Parses until end delimiter is found           |
| `Try(p)`                  | Backtracking support on failure               |
| `Lazy(f)`                 | Allows self-referencing parsers (recursion)   |
| `Lexeme(p)`               | Parses `p`, then skips trailing whitespace    |
| `Debug(p, name)`          | Prints trace for the parser                   |

---

## Example: Comma-Separated Digits

```go
digit := parser.Digit()
comma := parser.Lexeme(parser.RuneParser("comma", ','))

list := parser.SeparatedBy("digit list", digit, comma)

input := "1,2,3"
state := state.NewState(input, state.Position{0, 1, 1})
res, err := list.Run(state)
fmt.Println(res.Value) // Output: ['1', '2', '3']
```

---

## Error Tracing Example

If a parser fails, you get rich diagnostics:

```text
Or combinator failed
At: Line 1, Column 1, Offset 0
1| abc
   ^ 
Expected: Digit         Got: a
```

Trace includes:

* Position and snippet
* Expected vs. got values
* Full recursive error cause chain

---

## Installation

```bash
go get github.com/BlackBuck/pcom-go
```

---

## Testing and Benchmarking

```bash
go test ./...
go test -bench=. ./...
```

---

## Roadmap

* [x] Generic `Parser[T]`
* [x] Structured error reporting with position + snippet
* [x] Core combinators: `Or`, `Then`, `Map`, `Many0`, `Optional`
* [x] Support for recursion via `Lazy`
* [x] `SeparatedBy`, `ManyTill`, `TakeWhile`
* [ ] Struct decoding into ASTs
* [ ] JSON / INI parser examples
* [ ] Codegen or DSL grammar support (future)

---

## Examples

See [`/examples`](./examples) for:

* Arithmetic expression parser
* Identifier / keyword matcher
* JSON parser (WIP)

---

## Contributing

Pull requests, issues, and parser ideas are welcome!

To contribute:

```bash
git clone https://github.com/BlackBuck/pcom-go
git checkout -b feature/my-feature
# make changes...
git commit -m "Add my feature"
git push origin feature/my-feature
```

---

## License

MIT License © 2025 [Anil Bishnoi](https://github.com/BlackBuck)