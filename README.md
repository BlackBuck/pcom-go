# pcom-go

[![Go Report Card](https://goreportcard.com/badge/github.com/BlackBuck/pcom-go)](https://goreportcard.com/report/github.com/BlackBuck/pcom-go)

**pcom-go** is a composable, generic **parser combinator library** written in Go, inspired by Haskell's `parsec` and Rust's `nom`.  
It enables you to build powerful parsers in a modular way with **comprehensive error reporting**, **type-safe generics**, and **support for recursive grammars**.

> Perfect for building parsers for arithmetic expressions, configuration files, domain-specific languages, or structured data formats in pure Go.

---

## Features

- **Type-safe parser combinators** using Go 1.18+ generics
- **Detailed error reporting** with position information, context snippets, and error traces
- **Recursive grammar support** with `Lazy` combinator for forward references
- **Rich set of primitives** for common parsing patterns (digits, letters, strings, etc.)
- **Powerful combinators** for sequencing, choice, repetition, mapping, and more
- **Precise state tracking** with line, column, and offset information
- **Memory-efficient** with proper backtracking and state management
- **Thoroughly tested** with comprehensive benchmarks

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

    res, err := letters.Run(&s)
    if err.HasError() {
        fmt.Println(err.FullTrace())
        return
    }

    fmt.Printf("Parsed: %v\n", res.Value) // Output: [a b c]
}
```

---

## Core Concepts

A parser transforms input text into structured data. Each parser is a function that takes a parsing state and returns either a successful result or an error with detailed diagnostics.

```go
type Parser[T any] struct {
    Run   func(*state.State) (Result[T], Error)
    Label string
}

type Result[T any] struct {
    Value     T            // The parsed value
    NextState *state.State // Updated parser state
    Span      state.Span   // Source location span
}
```

---

## 🛠️ API Overview

### Primitives

| Function                      | Description                                  |
| ----------------------------- | -------------------------------------------- |
| `RuneParser("label", 'x')`    | Parses a specific rune                       |
| `StringParser("label", "hi")` | Parses an exact string                       |
| `Digit()`                     | Parses a single digit (0-9)                  |
| `Alpha()`                     | Parses a single letter (a-z, A-Z)            |
| `AlphaNum()`                  | Parses a letter or digit                     |
| `Whitespace()`                | Parses a single space character              |
| `AnyChar()`                   | Parses any single character                  |
| `CharWhere(label, predicate)` | Parses a character matching custom condition |
| `StringCI("hello")`           | Case-insensitive string matching             |
| `OneOf("+-*/")`               | Parses one character from the given set      |
| `TakeWhile(label, predicate)` | Consumes characters while predicate is true  |

### Combinators

| Function                         | Description                                 |
| -------------------------------- | ------------------------------------------- |
| `Or(label, p1, p2, ...)`         | Try parsers in order, return first success  |
| `And(label, p1, p2, ...)`        | All parsers must succeed at same position   |
| `Sequence(label, []p)`           | Run parsers in sequence, return last result |
| `Then(label, p1, p2)`            | Combine two parsers into a `Pair[A, B]`     |
| `KeepLeft(label, p)`             | Keep only the left value from a pair        |
| `KeepRight(label, p)`            | Keep only the right value from a pair       |
| `Map(label, p, func)`            | Transform parser result with a function     |
| `Optional(label, p)`             | Zero-or-one occurrence, never fails         |
| `Many0(label, p)`                | Zero or more repetitions                    |
| `Many1(label, p)`                | One or more repetitions                     |
| `Between(label, open, p, close)` | Parse content between delimiters            |
| `SeparatedBy(label, p, sep)`     | Parse values separated by delimiter         |
| `ManyTill(label, p, end)`        | Parse until end delimiter is found          |
| `Lazy(label, func)`              | Enable recursive/forward-reference parsers  |
| `Lexeme(p)`                      | Parse `p` then consume trailing whitespace  |
| `Chainl1(label, p, op)`          | Left-associative binary operations          |
| `Chainr1(label, p, op)`          | Right-associative binary operations         |
| `Not(label, p)`                  | Negative lookahead (succeed if `p` fails)   |

---

## Example: Parsing Comma-Separated Digits

```go
package main

import (
    "fmt"
    "github.com/BlackBuck/pcom-go/parser"
    "github.com/BlackBuck/pcom-go/state"
)

func main() {
    digit := parser.Digit()
    comma := parser.Lexeme(parser.RuneParser("comma", ','))

    list := parser.SeparatedBy("digit list", digit, comma)

    input := "1, 2, 3"
    s := state.NewState(input, state.Position{Offset: 0, Line: 1, Column: 1})

    res, err := list.Run(&s)
    if err.HasError() {
        fmt.Println(err.FullTrace())
        return
    }

    fmt.Printf("Parsed digits: %v\n", res.Value) // Output: [49 50 51] (rune values)
}
```

---

## Arithmetic Expression Parser

Here's a complete example of parsing arithmetic expressions with operator precedence:

```go
// Parse numbers
number := parser.Map("number", parser.Many1("digits", parser.Digit()),
    func(digits []rune) int {
        // Convert runes to integer
        num := 0
        for _, d := range digits {
            num = num*10 + int(d-'0')
        }
        return num
    })

// Parse operators
addOp := parser.Map("add", parser.RuneParser("plus", '+'),
    func(_ rune) func(int, int) int {
        return func(a, b int) int { return a + b }
    })

mulOp := parser.Map("mul", parser.RuneParser("times", '*'),
    func(_ rune) func(int, int) int {
        return func(a, b int) int { return a * b }
    })

// Build expression parser with precedence
term := parser.Chainl1("term", parser.Lexeme(number), parser.Lexeme(mulOp))
expr := parser.Chainl1("expr", term, parser.Lexeme(addOp))

// Parse "2 + 3 * 4" => 14 (respects precedence)
```

---

## Error Reporting

When parsing fails, pcom-go provides detailed error information:

```go
input := "12a"
number := parser.Many1("digits", parser.Digit())
s := state.NewState(input, state.Position{Offset: 0, Line: 1, Column: 1})

_, err := number.Run(&s)
if err.HasError() {
    fmt.Println(err.FullTrace())
}
```

Output includes:

- **Error location**: Line, column, and character offset
- **Context snippet**: The surrounding source code
- **Expected vs. actual**: What the parser expected vs. what it found
- **Error chain**: Full trace of nested parser failures

---

## Installation

```bash
go get github.com/BlackBuck/pcom-go
```

**Requirements**: Go 1.18+ (for generics support)

---

## Testing and Benchmarking

Run the test suite:

```bash
go test ./...
```

Run benchmarks:

```bash
go test -bench=. ./benchmark/
```

---

## Project Status

**Current Version**: v0.2.0

### Completed Features ✅

- Core parser combinators (`Or`, `And`, `Then`, `Map`, etc.)
- Comprehensive primitive parsers
- Recursive parsing with `Lazy`
- Rich error reporting with context
- Performance optimizations and benchmarks
- Expression parsing with operator precedence

### Roadmap 🚧

- [ ] JSON/XML parser examples
- [ ] Stream processing capabilities
- [ ] Parser debugging utilities
- [ ] Advanced error recovery
- [ ] Custom error types and formatting

---

## Examples

The [`/examples`](./examples) directory contains complete, working examples:

### Arithmetic Expression Parser

```bash
cd examples/expressions
go run expression_parser.go
```

Demonstrates:

- Operator precedence (`*`, `/` before `+`, `-`)
- Parentheses for grouping
- Recursive grammar with `Lazy`
- AST construction and evaluation

### Quick Start Example

```bash
cd examples/quickstart
go run quickstart.go
```

Basic string parsing demonstration.

---

## Contributing

We welcome contributions! Here's how to get started:

```bash
# Clone the repository
git clone https://github.com/BlackBuck/pcom-go
cd pcom-go

# Create a feature branch
git checkout -b feature/my-improvement

# Make your changes and add tests
go test ./...

# Submit a pull request
git commit -m "Add: my improvement"
git push origin feature/my-improvement
```

### Contribution Guidelines

- Write tests for new features
- Follow Go conventions and `gofmt` formatting
- Add examples for complex features
- Update documentation as needed

---

## License

MIT License © 2025 [BlackBuck](https://github.com/BlackBuck)

See [LICENSE.md](./LICENSE.md) for full details.
