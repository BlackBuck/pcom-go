# pcom-go

[![Go Report Card](https://goreportcard.com/badge/github.com/BlackBuck/pcom-go)](https://goreportcard.com/report/github.com/BlackBuck/pcom-go)

**pcom-go** is a modular, generic **parser combinator library** written in Go.  
Inspired by Haskell's `parsec` and Rust's `nom`, this library allows you to build complex parsers from smaller, composable units — now with **detailed error traces.**

> Write expressive parsers for structured text, config files, DSLs, or even JSON — all in pure Go!

---

## Features

- Generic, type-safe parser combinators using Go 1.18+
- Precise error tracking with **full trace stack, snippets, and positions**
- Backtracking and custom error messages (`Try`, `Or`)
- Primitives for character-level parsing (`Digit`, `Alpha`, etc.)
- Higher-order combinators like `Many`, `Between`, `Map`, `Lazy`
- CLI-friendly colored error output
- Easy testing and benchmarking

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

    digit := parser.Digit()
    alpha := parser.Alpha()

    parserSequence := parser.Many1("letters", alpha)

    res, err := parserSequence.Run(s)
    if err.HasError() {
        fmt.Println(err.FullTrace())
        return
    }

    fmt.Printf("Parsed: %v\n", res.Value)
}
````

---

## Core Concepts

### `Parser[T]`

A parser is a function that consumes input and returns a `Result[T]` or an `Error`. Parsers can be composed using combinators like `Or`, `Many`, `Map`, and more.

```go
type Parser[T any] struct {
	Run   func(State) (Result[T], Error)
	Label string
}
```

---

## 🛠️ API Overview

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
| `Then(p1, p2)`            | Chains parsers and returns both results       |
| `KeepLeft(p)`             | Keeps left parser result                      |
| `KeepRight(p)`            | Keeps right parser result                     |
| `Debug(p, name)`          | Logs parser execution                         |

---

## Example: Arithmetic Expression Parser

```go
num := parser.Many1("number", parser.Digit())
plus := parser.RuneParser("+", '+')

expr := parser.Map(parser.Sequence([]parser.Parser[string]{
    num,
    parser.Lexeme(plus),
    num,
}), func(values string) int {
    // Simplified: parse "3 + 4" → 7
    return ...
})
```

---

## Error Reporting Example

When a parser fails, you now get **full trace errors:**

```text
Parser: Or
Position: 1:1
Message: Or combinator failed
Expected: Digit
Got: 'a'
Snippet: abc123

Parser: DigitParser
Position: 1:1
Message: Failed to parse digit
Expected: Digit
Got: 'a'
Snippet: abc123
```

The trace shows:

* Which parsers failed
* The full stack of parser attempts
* Expected vs. actual values
* Input snippets with precise line and column positions

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
* Performance benchmarks

---

## Examples

Example parsers (coming soon in `/examples`):

* JSON parser
* Arithmetic evaluator
* Config file grammar

---

## Contributing

Pull requests, issues, and suggestions are welcome!

### To contribute:

* Fork the repo
* Create a new branch: `feature/my-feature`
* Submit a pull request

For major changes, please open an issue first to discuss your ideas.

---

## 🛡️ License

MIT License © 2025 [Anil Bishnoi](https://github.com/BlackBuck)