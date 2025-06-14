package parser

import (
	"fmt"

	state "github.com/BlackBuck/pcom-go/state"
)

type Pair[A, B any] struct {
	Left  A
	Right B
}
type Result[T any] struct {
	Value     T
	NextState state.State
	Span      state.Span
}

type Parser[T any] struct {
	Run   func(state state.State) (result Result[T], error Error)
	Label string
}

func NewResult[T any](value T, nextState state.State, span state.Span) Result[T] {
	return Result[T]{value, nextState, span}
}

// parser a single rune
func RuneParser(label string, c rune) Parser[rune] {
	return Parser[rune]{
		Run: func(curState state.State) (Result[rune], Error) {
			if !curState.InBounds(curState.Offset) {

				return Result[rune]{}, Error{
					Message:  "Reached the end of file while parsing",
					Expected: string(c),
					Got:      "EOF",
					Snippet:  state.GetSnippetStringFromCurrentContext(curState),
					Position: state.NewPositionFromState(curState),
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
			}
		},
		Label: label,
	}
}

func StringParser(label string, s string) Parser[string] {
	return Parser[string]{
		Run: func(curState state.State) (Result[string], Error) {
			if !curState.InBounds(curState.Offset + len(s) - 1) {

				return Result[string]{}, Error{
					Message:  "Reached the end of file while parsing",
					Expected: s,
					Got:      "EOF",
					Snippet:  state.GetSnippetStringFromCurrentContext(curState),
					Position: state.NewPositionFromState(curState),
				}
			}

			if curState.Input[curState.Offset:curState.Offset+len(s)] != s {
				// TODO: Run which is better, passing the state (all state functions without pointer)
				// or updating the state in-place (all state functions with a pointer)

				t := curState
				t.Consume(len(s))
				return Result[string]{}, Error{
					Message:  "Strings do not match.",
					Expected: s,
					Snippet:  state.GetSnippetStringFromCurrentContext(curState),
					Got:      curState.Input[curState.Offset : curState.Offset+len(s)],
					Position: state.NewPositionFromState(curState),
				}
			}

			prev := state.NewPositionFromState(curState)
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

//TODO: Handle empty arrays for empty Parser[T] arrays as well

// the OR combinator
func Or[T any](label string, parsers ...Parser[T]) Parser[T] {
	return Parser[T]{
		Run: func(curState state.State) (Result[T], Error) {
			var lastErr Error
			for _, parser := range parsers {
				res, err := parser.Run(curState) // sends a copy
				if !err.HasError() {
					return res, Error{}
				}
				lastErr = err
			}

			// furthest error with position
			return Result[T]{}, Error{
				Message:  "Or combinator failed",
				Expected: lastErr.Expected,
				Got:      lastErr.Got,
				Snippet:  state.GetSnippetStringFromCurrentContext(curState),
				Position: lastErr.Position,
			}
		},
		Label: label,
	}
}

func And[T any](label string, parsers ...Parser[T]) Parser[T] {
	return Parser[T]{
		Run: func(curState state.State) (Result[T], Error) {
			var lastRes Result[T]
			for _, parser := range parsers {
				res, err := parser.Run(curState) // sends a copy
				if err.HasError() {

					return Result[T]{}, Error{
						Message:  "And combinator failed.",
						Expected: err.Expected,
						Got:      err.Got,
						Snippet:  state.GetSnippetStringFromCurrentContext(curState),
						Position: err.Position,
					}
				}
				lastRes = res
			}

			return lastRes, Error{}
		},
		Label: label,
	}
}

func Many0[T any](label string, p Parser[T]) Parser[[]T] {
	return Parser[[]T]{
		Run: func(curState state.State) (Result[[]T], Error) {
			var results []T
			originalState := curState
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
					Start: state.NewPositionFromState(originalState),
					End:   state.NewPositionFromState(curState),
				},
			}, Error{}
		},
		Label: label,
	}
}

func Many1[T any](label string, p Parser[T]) Parser[[]T] {
	return Parser[[]T]{
		Run: func(curState state.State) (Result[[]T], Error) {
			var results []T
			originalState := curState
			for {
				res, err := p.Run(curState)
				if err.HasError() {
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
						Start: state.NewPositionFromState(originalState),
						End:   state.NewPositionFromState(curState),
					},
				}, Error{}
			}

			return Result[[]T]{}, Error{
				Message:  "Many1 parser failed.",
				Expected: fmt.Sprintf("<%s> at least once", p.Label),
				Got:      fmt.Sprintf("<%s> zero times", p.Label),
				Snippet:  state.GetSnippetStringFromCurrentContext(curState),
				Position: state.NewPositionFromState(curState),
			}
		},
		Label: label,
	}
}

func Optional[T any](label string, p Parser[T]) Parser[T] {
	return Parser[T]{
		Run: func(curState state.State) (Result[T], Error) {
			res, err := p.Run(curState)
			if err.HasError() {
				return Result[T]{}, Error{}
			}

			return res, Error{}
		},
		Label: label,
	}
}

func Sequence[T any](label string, parsers []Parser[T]) Parser[T] {
	return Parser[T]{
		Run: func(curState state.State) (Result[T], Error) {
			var ret Result[T]
			for _, parser := range parsers {
				res, err := parser.Run(curState)
				if err.HasError() {

					return Result[T]{}, Error{
						Message:  "Sequence parser failed.",
						Expected: err.Expected,
						Got:      err.Got,
						Snippet:  state.GetSnippetStringFromCurrentContext(curState),
						Position: state.NewPositionFromState(curState),
					}
				}
				ret = res
			}
			return ret, Error{}
		},
		Label: label,
	}
}

func Map[A, B any](label string, p1 Parser[A], f func(A) B) Parser[B] {
	return Parser[B]{
		Run: func(curState state.State) (result Result[B], error Error) {
			res, err := p1.Run(curState)
			if err.HasError() {
				return Result[B]{}, Error{
					Message:  "Map parser failed",
					Expected: err.Expected,
					Got:      err.Got,
					Snippet:  err.Snippet,
					Position: err.Position,
				}
			}

			return Result[B]{
				Value:     f(res.Value),
				NextState: res.NextState,
				Span: state.Span{
					Start: state.NewPositionFromState(curState),
					End:   state.NewPositionFromState(res.NextState),
				},
			}, Error{}
		},
		Label: label,
	}
}

func Then[A, B any](label string, p1 Parser[A], p2 Parser[B]) Parser[Pair[A, B]] {
	return Parser[Pair[A, B]]{
		Run: func(curState state.State) (result Result[Pair[A, B]], error Error) {
			leftRes, err := p1.Run(curState)
			if err.HasError() {
				return Result[Pair[A, B]]{}, Error{
					Message:  "Left of Then failed",
					Expected: err.Expected,
					Got:      err.Got,
					Snippet:  err.Snippet,
					Position: err.Position,
				}
			}

			rightRes, err := p2.Run(leftRes.NextState)
			if err.HasError() {
				return Result[Pair[A, B]]{}, Error{
					Message:  "Right of Then failed",
					Expected: err.Expected,
					Got:      err.Got,
					Snippet:  err.Snippet,
					Position: err.Position,
				}
			}

			return Result[Pair[A, B]]{
				Value:     Pair[A, B]{leftRes.Value, rightRes.Value},
				NextState: rightRes.NextState,
				Span: state.Span{
					Start: state.NewPositionFromState(curState),
					End:   state.NewPositionFromState(rightRes.NextState),
				},
			}, Error{}
		},
		Label: label,
	}
}

func KeepLeft[A, B any](label string, p Parser[Pair[A, B]]) Parser[A] {
	return Parser[A]{
		Run: func(curState state.State) (result Result[A], error Error) {
			res, err := p.Run(curState)
			if err.HasError() {
				return Result[A]{}, Error{
					Message:  "KeepLeft failed.",
					Expected: err.Expected,
					Got:      err.Got,
					Position: err.Position,
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

func KeepRight[A, B any](label string, p Parser[Pair[A, B]]) Parser[B] {
	return Parser[B]{
		Run: func(curState state.State) (result Result[B], error Error) {
			res, err := p.Run(curState)
			if err.HasError() {
				return Result[B]{}, Error{
					Message:  "KeepRight failed.",
					Expected: err.Expected,
					Got:      err.Got,
					Position: err.Position,
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

func Between[L, C, R any](label string, open Parser[L], content Parser[C], close Parser[R]) Parser[C] {
	return Parser[C]{
		Run: func(curState state.State) (result Result[C], error Error) {
			left := KeepLeft("", Then("", content, close))
			right := KeepRight("", Then("", open, left))

			res, err := right.Run(curState)
			if err.HasError() {
				return Result[C]{}, Error{
					Message:  "Between combinator failed.",
					Expected: err.Expected,
					Got:      err.Got,
					Position: err.Position,
				}
			}

			return res, Error{}
		},
		Label: label,
	}
}

func Lazy[T any](label string, f func() Parser[T]) Parser[T] {
	var memo *Parser[T] // cached result

	return Parser[T]{
		Run: func(curState state.State) (result Result[T], error Error) {
			// lazily initialize once
			if memo == nil {
				p := f()
				memo = &p
			}

			return memo.Run(curState)
		},
		Label: label,
	}
}
