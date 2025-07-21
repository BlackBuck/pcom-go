package parser

import (
	"fmt"
	"sync"

	state "github.com/BlackBuck/pcom-go/state"
)

type Pair[A, B any] struct {
	Left  A
	Right B
}

// Result represents the outcome of a parser.
// Value holds the parsed value of type T.
// NextState is the parser state after parsing is complete.
// Span indicates the range in the input that was consumed by the parser.
type Result[T any] struct {
	Value     T
	NextState *state.State
	Span      state.Span
}

type Parser[T any] struct {
	Run   func(curState *state.State) (result Result[T], error Error)
	Label string
}

func NewResult[T any](value T, nextState *state.State, span state.Span) Result[T] {
	return Result[T]{value, nextState, span}
}

// RuneParser parses a single rune from the input.
// If the end of input is reached, it returns an EOF error.
// If the next input rune matches the expected rune, it returns it in the Result.
// Otherwise, it returns an Error indicating the mismatch.
// Example: RuneParser("myRune", 'a') will parse 'a' from the input.
// If the input does not match 'a' at the current position, it returns an error.
func RuneParser(label string, c rune) Parser[rune] {
	return Parser[rune]{
		Run: func(curState *state.State) (Result[rune], Error) {
			if !curState.InBounds(curState.Offset) {
				return Result[rune]{}, Error{
					Message:  "Reached the end of file while parsing",
					Expected: string(c),
					Got:      "EOF",
					Snippet:  state.GetSnippetStringFromCurrentContext(curState),
					Position: state.NewPositionFromState(curState),
					Cause:    nil,
				}
			}
			if curState.Input[curState.Offset] == byte(c) {
				prev := state.NewPositionFromState(curState)
				curState.Consume(1)
				return NewResult(
					c,
					curState,
					state.Span{
						Start: prev,
						End:   state.NewPositionFromState(curState),
					}), Error{}
			}

			return Result[rune]{}, Error{
				Message:  fmt.Sprintf("Failed to parse %s", label),
				Expected: string(c),
				Got:      string(curState.Input[curState.Offset]),
				Snippet:  state.GetSnippetStringFromCurrentContext(curState),
				Position: state.NewPositionFromState(curState),
				Cause:    nil,
			}
		},
		Label: label,
	}
}

// StringParser parses the exact string s (case-sensitive) from the input.
// If the input does not match s at the current position, it returns an error.
// Example: StringParser("myString", "hello") will parse "hello" from the input.
// If the input does not match "hello" at the current position, it returns an error.
// If the input matches, it returns the parsed string and updates the state.
// If the end of input is reached before matching, it returns an EOF error.
func StringParser(label string, s string) Parser[string] {
	return Parser[string]{
		Run: func(curState *state.State) (Result[string], Error) {
			if !curState.InBounds(curState.Offset + len(s) - 1) {
				return Result[string]{}, Error{
					Message:  "Reached the end of file while parsing",
					Expected: s,
					Got:      "EOF",
					Snippet:  state.GetSnippetStringFromCurrentContext(curState),
					Position: state.NewPositionFromState(curState),
					Cause:    nil,
				}
			}

			if curState.Input[curState.Offset:curState.Offset+len(s)] != s {
				// TODO: Run which is better, passing the state (all state functions without pointer)
				// or updating the state in-place (all state functions with a pointer)

				return Result[string]{}, Error{
					Message:  "Strings do not match.",
					Expected: s,
					Snippet:  state.GetSnippetStringFromCurrentContext(curState),
					Got:      curState.Input[curState.Offset : curState.Offset+len(s)],
					Position: state.NewPositionFromState(curState),
					Cause:    nil,
				}
			}

			prev := curState.Save()
			curState.Consume(len(s))
			return NewResult(
				s,
				curState,
				state.Span{
					Start: prev,
					End:   state.NewPositionFromState(curState),
				}), Error{}

		},
		Label: label,
	}
}

// Or tries each parser in order and returns the result of the first one that succeeds.
// If all parsers fail, it returns the error from the parser that got the furthest.
// This is useful for alternatives, e.g. parsing either an integer or a string.
//
// Example usage:
//
//   intParser := parser.StringParser("int", "123")
//   strParser := parser.StringParser("str", "abc")
//   altParser := parser.Or("int or str", intParser, strParser)
//   res, err := altParser.Run(state)
//   // res.Value will be "123" or "abc" depending on input
// // If both parsers fail, err will contain the error from the last parser that was tried.
// // Note: The error returned will have the position of the last parser that was tried,
// // so you can see where the failure occurred in the input.
// // If you want to handle the error, you can check if err.HasError() is true.
func Or[T any](label string, parsers ...Parser[T]) Parser[T] {
	return Parser[T]{
		Run: func(curState *state.State) (Result[T], Error) {
			var lastErr Error
			for _, parser := range parsers {
				cp := curState.Save()
				res, err := parser.Run(curState) // sends a copy
				if !err.HasError() {
					return res, Error{}
				}
				curState.Rollback(cp) // rollback to previous safe state on error
				lastErr = err
			}

			// furthest error with position
			return Result[T]{}, Error{
				Message:  "Or combinator failed",
				Expected: lastErr.Expected,
				Got:      lastErr.Got,
				Snippet:  state.GetSnippetStringFromCurrentContext(curState),
				Position: lastErr.Position,
				Cause:    &lastErr,
			}
		},
		Label: label,
	}
}

// And runs all provided parsers at the same input position (without advancing the state).
// It succeeds only if all parsers succeed at that position, returning the last parser's result.
// If any parser fails, it returns an error for that parser.
//
// Example usage:
//
//   alpha := parser.Alpha("alphabet") // See parser/primitives.go for more details
//   a := parser.RuneParser("a", 'a')
//   andParser := parser.And("alphabetic and a", alpha, a)
//   res, err := andParser.Run(state)
//   // res.Value will be the result of the last parser if both succeed at the same position.
//   // If either fails, err will contain the error.
func And[T any](label string, parsers ...Parser[T]) Parser[T] {
	return Parser[T]{
		Run: func(curState *state.State) (Result[T], Error) {
			var lastRes Result[T]
			for _, parser := range parsers {
				cp := curState.Save()
				res, err := parser.Run(curState)
				if err.HasError() {
					curState.Rollback(cp) // rollback on error
					return Result[T]{}, Error{
						Message:  "And combinator failed.",
						Expected: err.Expected,
						Got:      err.Got,
						Snippet:  state.GetSnippetStringFromCurrentContext(curState),
						Position: err.Position,
						Cause:    &err,
					}
				}
				curState.Rollback(cp) // run on the same input
				lastRes = res
			}

			return lastRes, Error{}
		},
		Label: label,
	}
}

// Many0 applies the given parser zero or more times, collecting the results in a slice.
// It always succeeds, returning an empty slice if the parser never succeeds.
// No error is returned, even if the parser fails on the first attempt.
//
// Example usage:
//
//   digit := parser.RuneParser("digit", '1')
//   digits := parser.Many0("zero or more 1s", digit)
//   res, err := digits.Run(state)
//   // res.Value will be []rune containing all parsed '1's in sequence (possibly empty).
func Many0[T any](label string, p Parser[T]) Parser[[]T] {
	return Parser[[]T]{
		Run: func(curState *state.State) (Result[[]T], Error) {
			var results []T
			initialPos := state.NewPositionFromState(curState)
			for {
				res, err := p.Run(curState)
				if err.HasError() {
					break
				}
				curState = res.NextState
				results = append(results, res.Value)
			}
			return Result[[]T]{
				Value:     results,
				NextState: curState,
				Span: state.Span{
					Start: initialPos,
					End:   state.NewPositionFromState(curState),
				},
			}, Error{}
		},
		Label: label,
	}
}

// Many1 applies the given parser one or more times, collecting the results in a slice.
// It succeeds only if the parser matches at least once; otherwise, it returns an error.
//
// Example usage:
//
//   digit := parser.RuneParser("digit", '1')
//   digits := parser.Many1("one or more 1s", digit)
//   res, err := digits.Run(state)
//   // res.Value will be []rune containing all parsed '1's in sequence (must be non-empty).
//   // If no '1' is found at the current position, err will be non-nil.
func Many1[T any](label string, p Parser[T]) Parser[[]T] {
	return Parser[[]T]{
		Run: func(curState *state.State) (Result[[]T], Error) {
			var results []T
			var cp state.Position
			initialPos := state.NewPositionFromState(curState)
			var lastErr Error
			for {
				cp = curState.Save()
				res, err := p.Run(curState)
				if err.HasError() {
					lastErr = err
					break
				}
				curState = res.NextState
				results = append(results, res.Value)
			}
			if len(results) > 0 {
				return Result[[]T]{
					Value:     results,
					NextState: curState,
					Span: state.Span{
						Start: initialPos,
						End:   state.NewPositionFromState(curState),
					},
				}, Error{}
			}

			curState.Rollback(cp) // rollback on error
			return Result[[]T]{}, Error{
				Message:  "Many1 parser failed.",
				Expected: fmt.Sprintf("<%s> at least once", p.Label),
				Got:      fmt.Sprintf("<%s> zero times", p.Label),
				Snippet:  state.GetSnippetStringFromCurrentContext(curState),
				Position: curState.Save(),
				Cause:    &lastErr,
			}
		},
		Label: label,
	}
}

// Optional tries to apply the given parser once, returning its result if it succeeds,
// or a zero value if it fails. It never returns an error.
//
// Example usage:
//
//   digit := parser.RuneParser("digit", '1')
//   optDigit := parser.Optional("optional 1", digit)
//   res, err := optDigit.Run(state)
//   // res.Value will be '1' if present, or the zero value for rune if not.
//   // err will always be nil.
func Optional[T any](label string, p Parser[T]) Parser[T] {
	return Parser[T]{
		Run: func(curState *state.State) (Result[T], Error) {
			cp := curState.Save()
			res, err := p.Run(curState)
			if err.HasError() {
				curState.Rollback(cp)
				return Result[T]{
					NextState: curState, // TODO: should I return this????
				}, Error{}
			}

			return res, Error{}
		},
		Label: label,
	}
}

// Sequence runs a list of parsers in order, advancing the input for each.
// It returns the result of the last parser if all succeed.
// If any parser fails, it returns an error and rolls back the input.
//
// Example usage:
//
//   p1 := parser.StringParser("hello", "hello")
//   p2 := parser.StringParser("world", "world")
//   seq := parser.Sequence("hello then world", []parser.Parser[string]{p1, p2})
//   res, err := seq.Run(state)
//   // res.Value will be "world" if both parsers succeed in sequence.
func Sequence[T any](label string, parsers []Parser[T]) Parser[T] {
	return Parser[T]{
		Run: func(curState *state.State) (Result[T], Error) {
			var ret Result[T]
			for _, parser := range parsers {
				cp := curState.Save()
				res, err := parser.Run(curState)
				if err.HasError() {
					curState.Rollback(cp)
					return Result[T]{}, Error{
						Message:  "Sequence parser failed.",
						Expected: err.Expected,
						Got:      err.Got,
						Snippet:  state.GetSnippetStringFromCurrentContext(curState),
						Position: state.NewPositionFromState(curState),
						Cause:    &err,
					}
				}
				ret = res
				curState = res.NextState
			}
			return ret, Error{}
		},
		Label: label,
	}
}

// Map transforms the result of a parser using a provided function.
// It runs the parser p1, and if it succeeds, applies the function f to its result.
// If p1 fails, Map returns the error from p1.
//
// Example usage:
//
//   digitParser := parser.RuneParser("digit", '1')
//   toInt := func(r rune) int { return int(r - '0') }
//   intParser := parser.Map("digit to int", digitParser, toInt)
//   res, err := intParser.Run(state)
//   // res.Value will be 1 if the input is '1'
func Map[A, B any](label string, p1 Parser[A], f func(A) B) Parser[B] {
	return Parser[B]{
		Run: func(curState *state.State) (result Result[B], error Error) {
			cp := curState.Save()
			res, err := p1.Run(curState)
			if err.HasError() {
				curState.Rollback(cp)
				return Result[B]{}, Error{
					Message:  "Map parser failed",
					Expected: err.Expected,
					Got:      err.Got,
					Snippet:  err.Snippet,
					Position: err.Position,
					Cause:    &err,
				}
			}

			return Result[B]{
				Value:     f(res.Value),
				NextState: res.NextState,
				Span: state.Span{
					Start: cp,
					End:   state.NewPositionFromState(res.NextState),
				},
			}, Error{}
		},
		Label: label,
	}
}

// Then runs two parsers sequentially: first p1, then p2, advancing the input for each.
// It returns a Pair containing the results of both parsers if both succeed.
// If either parser fails, it returns an error and rolls back the input.
//
// Example usage:
//
//   p1 := parser.StringParser("hello", "hello")
//   p2 := parser.StringParser("world", "world")
//   seq := parser.Then("hello then world", p1, p2)
//   res, err := seq.Run(state)
//   // res.Value.Left will be "hello", res.Value.Right will be "world" if both succeed.
func Then[A, B any](label string, p1 Parser[A], p2 Parser[B]) Parser[Pair[A, B]] {
	return Parser[Pair[A, B]]{
		Run: func(curState *state.State) (result Result[Pair[A, B]], error Error) {
			cp := curState.Save()
			leftRes, err := p1.Run(curState)
			if err.HasError() {
				curState.Rollback(cp)
				return Result[Pair[A, B]]{}, Error{
					Message:  "Left of Then failed",
					Expected: err.Expected,
					Got:      err.Got,
					Snippet:  err.Snippet,
					Position: err.Position,
					Cause:    &err,
				}
			}

			rightRes, err := p2.Run(leftRes.NextState)
			if err.HasError() {
				curState.Rollback(cp)
				return Result[Pair[A, B]]{}, Error{
					Message:  "Right of Then failed",
					Expected: err.Expected,
					Got:      err.Got,
					Snippet:  err.Snippet,
					Position: err.Position,
					Cause:    &err,
				}
			}

			return Result[Pair[A, B]]{
				Value:     Pair[A, B]{leftRes.Value, rightRes.Value},
				NextState: rightRes.NextState,
				Span: state.Span{
					Start: cp,
					End:   state.NewPositionFromState(rightRes.NextState),
				},
			}, Error{}
		},
		Label: label,
	}
}

// KeepLeft returns a parser that keeps only the Left value from a Pair produced by the given parser.
// This is useful when you want to sequence two parsers but only care about the result of the first.
//
// Example usage:
//
//   p1 := parser.StringParser("hello", "hello")
//   p2 := parser.StringParser("world", "world")
//   pairParser := parser.Then("hello then world", p1, p2)
//   leftOnly := parser.KeepLeft("keep hello", pairParser)
//   res, err := leftOnly.Run(state)
//   // res.Value will be "hello" if both parsers succeed.
func KeepLeft[A, B any](label string, p Parser[Pair[A, B]]) Parser[A] {
	return Parser[A]{
		Run: func(curState *state.State) (result Result[A], error Error) {
			cp := curState.Save()
			res, err := p.Run(curState)
			if err.HasError() {
				curState.Rollback(cp)
				return Result[A]{}, Error{
					Message:  "KeepLeft failed.",
					Expected: err.Expected,
					Got:      err.Got,
					Position: err.Position,
					Snippet:  err.Snippet,
					Cause:    &err,
				}
			}

			return Result[A]{
				Value:     res.Value.Left,
				NextState: res.NextState,
				Span:      res.Span,
			}, Error{}
		},
		Label: label,
	}
}

// KeepRight returns a parser that keeps only the Right value from a Pair produced by the given parser.
// This is useful when you want to sequence two parsers but only care about the result of the second.
//
// Example usage:
//
//   p1 := parser.StringParser("hello", "hello")
//   p2 := parser.StringParser("world", "world")
//   pairParser := parser.Then("hello then world", p1, p2)
//   rightOnly := parser.KeepRight("keep world", pairParser)
//   res, err := rightOnly.Run(state)
//   // res.Value will be "world" if both parsers succeed.
func KeepRight[A, B any](label string, p Parser[Pair[A, B]]) Parser[B] {
	return Parser[B]{
		Run: func(curState *state.State) (result Result[B], error Error) {
			cp := curState.Save()
			res, err := p.Run(curState)
			if err.HasError() {
				curState.Rollback(cp)
				return Result[B]{}, Error{
					Message:  "KeepRight failed.",
					Expected: err.Expected,
					Got:      err.Got,
					Position: err.Position,
					Snippet:  err.Snippet,
					Cause:    &err,
				}
			}

			return Result[B]{
				Value:     res.Value.Right,
				NextState: res.NextState,
				Span:      res.Span,
			}, Error{}
		},
		Label: label,
	}
}

// Between parses content that is surrounded by an open and a close parser.
// It returns the result of the content parser if all three parsers succeed in sequence.
// If any of open, content, or close fails, it returns an error.
//
// Example usage:
//
//   openParen := parser.RuneParser("open paren", '(')
//   closeParen := parser.RuneParser("close paren", ')')
//   inner := parser.StringParser("digits", "123")
//   betweenParens := parser.Between("digits in parens", openParen, inner, closeParen)
//   res, err := betweenParens.Run(state)
//   // res.Value will be "123" if the input is "(123)"
func Between[L, C, R any](label string, open Parser[L], content Parser[C], close Parser[R]) Parser[C] {
	return Parser[C]{
		Run: func(curState *state.State) (result Result[C], error Error) {
			left := KeepLeft("", Then("", content, close))
			right := KeepRight("", Then("", open, left))

			cp := curState.Save()
			res, err := right.Run(curState)
			if err.HasError() {
				curState.Rollback(cp)
				return Result[C]{}, Error{
					Message:  "Between combinator failed.",
					Expected: err.Expected,
					Got:      err.Got,
					Position: err.Position,
					Snippet:  err.Snippet,
					Cause:    &err,
				}
			}

			return res, Error{}
		},
		Label: label,
	}
}

// Lazy creates a parser that defers the construction of its inner parser until first use.
// This is useful for defining recursive parsers, such as for left-recursive grammars.
//
// Example usage:
//
//   var expr Parser[int]
//   expr = Lazy("expr", func() Parser[int] {
//       // expr can reference itself recursively here
//       return Or("sum",
//           Then("add", expr, plusOp), // left-recursive
//           numberParser,
//       )
//   })
func Lazy[T any](label string, f func() Parser[T]) Parser[T] {
	var p Parser[T]
	var once sync.Once // thread-safe Lazy init

	return Parser[T]{
		Run: func(curState *state.State) (Result[T], Error) {
			once.Do(func() {
				p = f()
			})
			return p.Run(curState)
		},
		Label: label,
	}
}

// Chainl1 parses one or more values using parser p, separated by the operator parser op,
// and folds them left-associatively. This is useful for parsing left-associative binary
// operations such as addition or subtraction.
//
// Example usage:
//
//   num := parser.StringParser("number", "1")
//   plus := parser.Map("plus", parser.RuneParser("plus", '+'), func(_ rune) func(int, int) int {
//       return func(a, b int) int { return a + b }
//   })
//   expr := parser.Chainl1("sum", num, plus)
//   res, err := expr.Run(state)
//   // Parses "1+1+1" as ((1+1)+1)
func Chainl1[T any](label string, p Parser[T], op Parser[func(T, T) T]) Parser[T] {
	return Parser[T]{
		Run: func(curState *state.State) (result Result[T], error Error) {
			cp := curState.Save()
			left, err := p.Run(curState)
			if err.HasError() {
				curState.Rollback(cp)
				return Result[T]{}, Error{
					Message:  "Chainl1: failed to parse initial value.",
					Expected: err.Expected,
					Got:      err.Got,
					Position: err.Position,
					Snippet:  err.Snippet,
					Cause:    &err,
				}
			}

			ass := left.Value
			curState = left.NextState
			for {
				f, err := op.Run(curState)
				if err.HasError() {
					break
				}

				right, err := p.Run(f.NextState)
				if err.HasError() {
					curState.Rollback(cp)
					return Result[T]{}, Error{
						Message:  "Chainl1: failed to parse right value.",
						Expected: err.Expected,
						Got:      err.Got,
						Position: err.Position,
						Snippet:  err.Snippet,
						Cause:    &err,
					}
				}
				ass = f.Value(ass, right.Value)
				curState = right.NextState
			}

			return Result[T]{
				Value:     ass,
				NextState: curState,
				Span: state.Span{
					Start: cp,
					End:   state.NewPositionFromState(curState),
				},
			}, Error{}
		},
		Label: label,
	}
}

// Chainr1 parses one or more values using parser p, separated by the operator parser op,
// and folds them right-associatively. This is useful for parsing right-associative binary
// operations such as exponentiation.
//
// Example usage:
//
//   num := parser.StringParser("number", "2")
//   pow := parser.Map("pow", parser.RuneParser("pow", '^'), func(_ rune) func(int, int) int {
//       return func(a, b int) int { return int(math.Pow(float64(a), float64(b))) }
//   })
//   expr := parser.Chainr1("power", num, pow)
//   res, err := expr.Run(state)
//   // Parses "2^3^2" as 2^(3^2)
func Chainr1[T any](label string, p Parser[T], op Parser[func(T, T) T]) Parser[T] {
	return Parser[T]{
		Run: func(curState *state.State) (result Result[T], error Error) {
			var vals []T
			var fs []func(T, T) T
			cp := curState.Save()
			leftVal, err := p.Run(curState)
			if err.HasError() {
				return Result[T]{}, Error{
					Message:  "Chainr1: failed to parse initial value.",
					Expected: err.Expected,
					Got:      err.Got,
					Position: err.Position,
					Snippet:  err.Snippet,
					Cause:    &err,
				}
			}

			vals = append(vals, leftVal.Value)
			curState = leftVal.NextState
			for {
				f, err := op.Run(curState)
				if err.HasError() {
					break
				}

				fs = append(fs, f.Value)
				rightVal, err := p.Run(f.NextState)
				if err.HasError() {
					curState.Rollback(cp)
					return Result[T]{}, Error{
						Message:  "Chainr1: failed to parse right value.",
						Expected: err.Expected,
						Got:      err.Got,
						Position: err.Position,
						Cause:    &err,
					}
				}
				vals = append(vals, rightVal.Value)
				curState = rightVal.NextState
			}

			for len(vals) > 1 {
				a := vals[len(vals)-1]
				b := vals[len(vals)-2]
				f := fs[len(fs)-1]
				fs = fs[:len(fs)-1]
				vals = vals[:len(vals)-2]
				vals = append(vals, f(a, b))
			}

			return Result[T]{
				Value:     vals[0],
				NextState: curState,
				Span: state.Span{
					Start: cp,
					End:   state.NewPositionFromState(curState),
				},
			}, Error{}
		},
		Label: label,
	}
}
