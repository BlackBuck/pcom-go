# PCOM-GO: A Parser Combinator Library in Go

The `pcom-go` package is a lightweight, flexible library for building parsers in Go using the parser combinator approach. It enables you to construct complex parsers by combining simple ones, making it suitable for tasks like parsing custom data formats, configuration files, or domain-specific languages.

## Table of Contents
1. [Installation](#installation)
2. [Core Concepts](#core-concepts)
3. [API Overview](#api-overview)
4. [Usage Examples](#usage-examples)
   - [State Helper Methods](#state-helper-methods)
   - [CharParser](#charparser)
   - [String](#string)
   - [Or](#or)
   - [And](#and)
   - [Map](#map)
   - [Many0 and Many1](#many0-and-many1)
   - [Seq](#seq)
   - [Optional](#optional)
   - [Between](#between)
   - [Lazy](#lazy)
   - [Parsing a CSV-like Format](#parsing-a-csv-like-format)
5. [Error Handling](#error-handling)
6. [Contributing](#contributing)

---

## Installation

To use the `parser` package, include it in your Go project. Assuming it’s hosted at `github.com/BlackBuck/pcom-go`, install it with:

```bash
go get github.com/BlackBuck/pcom-go
```

Import it in your code:

```go
import "github.com/BlackBuck/pcom-go"
```

---

## Core Concepts

The library revolves around two key types: `State` and `Result`.

- **`State`**: Represents the current parsing position with an input string and an offset. It includes helper methods to manipulate and inspect the input:
  - `HasAvailableChars(n int) bool`: Checks if there are at least `n` characters left to parse.
  - `Consume(n int) (string, error)`: Consumes and returns `n` characters, advancing the offset.
  - `PeekChar() (byte, error)`: Returns the next character without advancing the offset (note: current implementation advances; this might be a bug—see below).
  - `Advance(n int) State`: Returns a new `State` with the offset moved forward by `n`.
- **`Result`**: Holds the parsed value and the next `State` after parsing.
- **`Parser`**: A function type that takes a `State` and returns a `Result` and an `error`.

Parsers are combined using combinators like `Or`, `And`, `Seq`, and others to build more complex parsing logic. The parsed result is returned as an `interface{}`, allowing flexibility in the types of data parsed.

**Note**: The `PeekChar` method currently advances the offset, which might not be the intended behavior for a "peek" operation. Typically, peeking should not modify the state. I am planning on changing this implementation.

---

## API Overview

| Function            | Description                                      | Signature                                      |
|---------------------|--------------------------------------------------|------------------------------------------------|
| `CharParser(c)`     | Matches a specific character                    | `func CharParser(c byte) Parser`              |
| `String(s)`         | Matches an exact string                         | `func String(s string) Parser`                |
| `Or(left, right)`   | Tries `left`, then `right` if `left` fails      | `func Or(left, right Parser) Parser`          |
| `And(left, right)`  | Runs `left` then `right`, returns both results  | `func And(left, right Parser) Parser`         |
| `Map(p, f)`         | Transforms the result of `p` with `f`           | `func Map[A, B any](p Parser, f func(A) B) Parser` |
| `Many0(p)`          | Matches zero or more occurrences of `p`         | `func Many0(p Parser) Parser`                 |
| `Many1(p)`          | Matches one or more occurrences of `p`          | `func Many1(p Parser) Parser`                 |
| `Seq(parsers...)`   | Runs parsers sequentially                       | `func Seq(parsers ...Parser) Parser`          |
| `Optional(p)`       | Matches `p` zero or one time                    | `func Optional(p Parser) Parser`              |
| `Between(o, c, cl)` | Parses content between `open` and `close`       | `func Between(open, content, close Parser) Parser` |
| `Lazy(f)`           | Delays parser creation until needed             | `func Lazy(f func() Parser) Parser`           |

---

## Usage Examples

### State Helper Methods

These methods assist in inspecting and manipulating the parsing state directly.

```go
package main

import (
	"fmt"
	"github.com/BlackBuck/pcom-go"
)

func main() {
	s := parser.NewState("hello", 0)

	// Check available characters
	fmt.Println("Has 2 chars?", s.HasAvailableChars(2)) // Has 2 chars? true
	fmt.Println("Has 6 chars?", s.HasAvailableChars(6)) // Has 6 chars? false

	// Consume characters
	chunk, err := s.Consume(2)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Consumed:", chunk) // Consumed: he
	fmt.Println("New offset:", s.Advance(2).offset) // New offset: 2

	// Peek at a character (note: current implementation advances)
	// This might be a bug; ideally, it shouldn't advance
	ch, err := s.PeekChar()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Peeked:", string(ch)) // Peeked: h
	fmt.Println("Offset after peek:", s.Advance(1).offset) // Offset after peek: 1

	// Advance manually
	s = s.Advance(1)
	fmt.Println("Advanced offset:", s.offset) // Advanced offset: 1
}
```

### CharParser

Matches a single character in the input.

```go
package main

import (
	"fmt"
	"github.com/BlackBuck/pcom-go"
)

func main() {
	p := parser.CharParser('a')
	result, err := p(parser.NewState("abc", 0))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Parsed:", result.parsedResult)    // Parsed: a
	fmt.Println("Offset:", result.nextState.offset) // Offset: 1
}
```

### String

Matches an exact string.

```go
package main

import (
	"fmt"
	"github.com/BlackBuck/pcom-go"
)

func main() {
	p := parser.String("hello")
	result, err := p(parser.NewState("hello world", 0))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Parsed:", result.parsedResult)    // Parsed: hello
	fmt.Println("Offset:", result.nextState.offset) // Offset: 5
}
```

### Or

Tries one parser, falling back to another if it fails.

```go
package main

import (
	"fmt"
	"github.com/BlackBuck/pcom-go"
)

func main() {
	p := parser.Or(parser.String("yes"), parser.String("no"))
	result, err := p(parser.NewState("no thanks", 0))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Parsed:", result.parsedResult)    // Parsed: no
	fmt.Println("Offset:", result.nextState.offset) // Offset: 2
}
```

### And

Runs two parsers in sequence, returning both results.

```go
package main

import (
	"fmt"
	"github.com/BlackBuck/pcom-go"
)

func main() {
	p := parser.And(parser.String("key"), parser.String("="))
	result, err := p(parser.NewState("key=value", 0))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Parsed:", result.parsedResult)    // Parsed: [key =]
	fmt.Println("Offset:", result.nextState.offset) // Offset: 4
}
```

### Map

Transforms the result of a parser.

```go
package main

import (
	"fmt"
	"github.com/BlackBuck/pcom-go"
)

func main() {
	p := parser.Map(parser.String("true"), func(s string) bool { return s == "true" })
	result, err := p(parser.NewState("true end", 0))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Parsed:", result.parsedResult)    // Parsed: true
	fmt.Println("Offset:", result.nextState.offset) // Offset: 4
}
```

### Many0 and Many1

`Many0` matches zero or more occurrences; `Many1` requires at least one.

```go
package main

import (
	"fmt"
	"github.com/BlackBuck/pcom-go"
)

func main() {
	p0 := parser.Many0(parser.CharParser('a'))
	result0, err := p0(parser.NewState("aaab", 0))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Many0 Parsed:", result0.parsedResult) // Many0 Parsed: [a a a]

	p1 := parser.Many1(parser.CharParser('a'))
	result1, err := p1(parser.NewState("aaab", 0))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Many1 Parsed:", result1.parsedResult) // Many1 Parsed: [a a a]
}
```

### Seq

Runs multiple parsers sequentially.

```go
package main

import (
	"fmt"
	"github.com/BlackBuck/pcom-go"
)

func main() {
	p := parser.Seq(parser.String("x"), parser.String(","), parser.String("y"))
	result, err := p(parser.NewState("x,y,z", 0))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Parsed:", result.parsedResult)    // Parsed: [x , y]
	fmt.Println("Offset:", result.nextState.offset) // Offset: 5
}
```

### Optional

Matches a parser zero or one time.

```go
package main

import (
	"fmt"
	"github.com/BlackBuck/pcom-go"
)

func main() {
	p := parser.Optional(parser.String("maybe"))
	result, err := p(parser.NewState("nope", 0))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Parsed:", result.parsedResult)    // Parsed: <nil>
	fmt.Println("Offset:", result.nextState.offset) // Offset: 0
}
```

### Between

Parses content between two delimiters.

```go
package main

import (
	"fmt"
	"github.com/BlackBuck/pcom-go"
)

func main() {
	p := parser.Between(parser.CharParser('('), parser.String("data"), parser.CharParser(')'))
	result, err := p(parser.NewState("(data)extra", 0))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Parsed:", result.parsedResult)    // Parsed: data
	fmt.Println("Offset:", result.nextState.offset) // Offset: 6
}
```

### Lazy

Delays parser creation until it’s needed (useful for recursive parsers).

```go
package main

import (
	"fmt"
	"github.com/BlackBuck/pcom-go"
)

func main() {
	var p parser.Parser
	p = parser.Lazy(func() parser.Parser {
		return parser.Or(parser.String("end"), parser.Seq(parser.CharParser('x'), p))
	})
	result, err := p(parser.NewState("xxend", 0))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Parsed:", result.parsedResult)    // Parsed: [x [x end]]
	fmt.Println("Offset:", result.nextState.offset) // Offset: 5
}
```

### Parsing a CSV-like Format

Parse a row like `"name,age,42"`.

```go
package main

import (
	"fmt"
	"github.com/BlackBuck/pcom-go"
)

func main() {
	// Simple word parser using State helpers
	word := func(curState parser.State) (parser.Result, error) {
		input := curState.input[curState.offset:]
		for i, r := range input {
			if r == ',' {
				return parser.NewResult(input[:i], curState.Advance(i)), nil
			}
		}
		return parser.NewResult(input, curState.Advance(len(input))), nil
	}

	comma := parser.CharParser(',')
	p := parser.Seq(word, comma, word, comma, word)
	result, err := p(parser.NewState("name,age,42 extra", 0))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Parsed:", result.parsedResult)    // Parsed: [name , age , 42]
	fmt.Println("Offset:", result.nextState.offset) // Offset: 11
}
```

---

## Error Handling

Parsers return an `error` when parsing fails, often with a descriptive message. Handle errors like this:

```go
result, err := p(parser.NewState("invalid", 0))
if err != nil {
	fmt.Printf("Parsing failed: %v\n", err)
	return
}
```

The `nextState` in the `Result` can be used to determine where parsing stopped.

---

## Contributing

Contributions are welcome! Please submit a pull request or open an issue on the [GitHub repository](https://github.com/BlackBuck/pcom-go).

---