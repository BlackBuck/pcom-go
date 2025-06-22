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

// Result `struct` stores the result of a parser.
// `Value` depends on the type of the parser.
// `NextState` stores a copy of the state after parser has done its work.
// `Span` determines the start and end position of the result in the Input.
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

// RuneParser parses a single rune.
// It returns an EOF error if entire input had been parsed earlier.
// If it matches the input rune successfully, it returns it with the `Result` else returns an Error.
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

// StringParser parses a string(case-sensitive).
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

// Or performs a logical OR operation between the input parsers.
// It returns, lazily, the Result after the first parser succeeds.
// If no parser succeeds, it returns the furthest error.
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

// And performs a logical AND operation between the parsers.
// It returns the Result after all the parsers succeed.
// If any parser fails, it returns an error of that parser failing.
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

// Many0 checks for the presence of a parser zero or more times.
// It returns an array of Result (empty if none succeeds).
// It does not return an error.
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

// Many1 checks for the presence of a parser one or more times.
// It is similar to Many0, but returns an error if it cannot parse even once.
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

// Optional checks for the presence of a parser zero or one times.
// It does not return any error.
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

// Sequence sequentially parses the input based on the parsers.
// It returns the Result after successfully running the last parser.
// If any parser fails, it returns an error.
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

// Map parses the output of one parser(p1) to a function.
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

// Then is used to run two parses one after the other.
// It returns a Pair containing the result of both parsers.
// If any parser fails, it returns an adequate error.
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

// KeepLeft is used to keep the result of the Left parser and discard the Right part.
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

// KeepRight is used to keep teh result of the Right parser and discard the Left part.
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

// Between is used to parse any content that is present between open and close parsers.
// It returns the Result of the content parser.
// It returns an error if any of open, content, or close fails.
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

// Lazy parser is used to lazily parse a parser.
// Useful for left-recursive parsing.
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

// Chainl1 parses one or more p values separated by op, and folds them left-associatively
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
