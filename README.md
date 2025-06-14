# pcom-go

[![Go Report Card](https://goreportcard.com/badge/github.com/BlackBuck/pcom-go)](https://goreportcard.com/report/github.com/BlackBuck/pcom-go)

**pcom-go** is a modular, generic **parser combinator library** written in Go. Inspired by Haskell's `parsec` and Rust's `nom`, this library allows you to build complex parsers from smaller, composable units.

> Write expressive parsers for structured text, config files, DSLs, or even JSON — all in pure Go!

---

## Features

- Generic, type-safe parser combinators using Go 1.18+
- Precise error tracking with line, column, and snippet highlighting
- Backtracking and custom error messages (`Try`, `Or`)
- Primitives for character-level parsing (`Digit`, `Alpha`, etc.)
- Higher-order combinators like `Many`, `Between`, `Map`, `Lazy`
- Easy testing and benchmarking
- CLI-friendly colored error output

---

## Core Concepts

### `Parser[T]`
A parser is a function that consumes input and returns a `Result[T]` or an error. You can compose them using combinators like `Or`, `Many`, and `Map`.

```go
type Parser[T any] struct {
	Run   func(State) (Result[T], Error)
	Label string
}
````

---

## API Overview

### Primitives

| Function              | Description                      |
| --------------------- | -------------------------------- |
| `RuneParser('x')`     | Parses a single rune             |
| `StringParser("let")` | Parses exact string              |
| `Digit()`             | Parses one digit                 |
| `Alpha()`             | Parses one letter                |
| `AlphaNum()`          | Parses one letter or digit       |
| `CharWhere()`         | Parses a rune based on predicate |
| `StringCI()`          | Case-insensitive string parser   |
| `Whitespace()`        | Matches `' '` (space) character  |
| `OneOf("+-*/")`       | Parses one of the given runes    |
| `AnyChar()`           | Matches any single rune          |

### Combinators

| Function                  | Description                                   |
| ------------------------- | --------------------------------------------- |
| `Or(p1, p2, ...)`         | Tries alternatives in order                   |
| `And(p1, p2)`             | Parses a sequence (returns last by default)   |
| `Sequence([]Parser[T])`   | Chains parsers in order (returns last)        |
| `Many0(p)`                | Zero or more repetitions                      |
| `Many1(p)`                | One or more repetitions                       |
| `Optional(p)`             | Returns success even if `p` fails             |
| `Between(open, p, close)` | Matches `open`, `p`, `close` in sequence      |
| `Try(p)`                  | Allows backtracking on failure                |
| `Lazy(f)`                 | Allows recursion by deferring parser creation |
| `Map(p, f)`               | Transforms parser output                      |
| `Lexeme(p)`               | Skips trailing space after parsing `p`        |

---

## Example: Arithmetic Expression Parser

```go
num := Many1(Digit())
plus := RuneParser('+')
expr := Map(Sequence([]Parser[string]{
    num,
    Lexeme(plus),
    num,
}), func(v string) int {
    // Simplified: parse "3 + 4" → 7
    return ...
})
```

---

## Error Reporting Example

When a parser fails, you get colored output:

```text
Error: unexpected token at line 3, column 5, offset 42
  2 | let x = 5
  3 | let y = ?
               ^
Expected: digit
Got: ?
```

---

## Installation

```bash
go get github.com/BlackBuck/pcom-go
```

---

## Running Tests

```bash
go test ./...
```

---

## Upcoming Features

* Struct decoding and mapping
* JSON and DSL grammar examples
* CLI tool for file parsing and AST output
* Memoization and packrat parsing
* Performance benchmarks(WIP)

---

## Examples

Examples will be included in `/examples`:

* JSON parser
* Arithmetic evaluator
* Config file grammar

---

## 🛡️ License

MIT License © 2025 [Anil Bishnoi](https://github.com/BlackBuck)
