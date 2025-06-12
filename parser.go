package parser

import (
	"fmt"
)

type Span struct {
	Start Position
	End   Position
}

type Result[T any] struct {
	Value     T
	NextState State
	Span      Span
}

type Parser[T any] struct {
	Run   func(state State) (result Result[T], error Error)
	Label string
}

func NewResult[T any](value T, nextState State, span Span) Result[T] {
	return Result[T]{value, nextState, span}
}

// parser a single rune
func RuneParser(c rune) Parser[rune] {
	return Parser[rune]{
		Run: func(curState State) (Result[rune], Error) {
			if !curState.InBounds(curState.offset) {
				lastLineStart := curState.LineStartBeforeCurrentOffset()
				return Result[rune]{}, Error{
					Message:  "Reached the end of file while parsing",
					Expected: string(c),
					Got:      "EOF",
					Snippet:  curState.Input[curState.lineStarts[lastLineStart]:curState.lineStarts[min(len(curState.lineStarts)-1, lastLineStart+1)]],
					Position: NewPositionFromState(curState),
				}
			}
			if curState.Input[curState.offset] == byte(c) {
				prev := NewPositionFromState(curState)
				curState.Consume(1)
				return NewResult(
					c,
					curState,
					Span{
						Start: prev,
						End:   NewPositionFromState(curState),
					}), Error{}
			}

			lastLineStart := curState.LineStartBeforeCurrentOffset()
			return Result[rune]{}, Error{
				Message:  "Reached the end of file while parsing",
				Expected: string(c),
				Got:      string(curState.Input[curState.offset]),
				Snippet:  curState.Input[curState.lineStarts[lastLineStart]:curState.lineStarts[min(len(curState.lineStarts)-1, lastLineStart+1)]],
				Position: NewPositionFromState(curState),
			}
		},
		Label: fmt.Sprintf("the character <%s>", string(c)),
	}
}

func StringParser(s string) Parser[string] {
	return Parser[string]{
		Run: func(curState State) (Result[string], Error) {
			if !curState.InBounds(curState.offset + len(s) - 1) {
				lastLineStart := curState.LineStartBeforeCurrentOffset()
				return Result[string]{}, Error{
					Message:  "Reached the end of file while parsing",
					Expected: s,
					Got:      "EOF",
					Snippet:  curState.Input[curState.lineStarts[lastLineStart]:curState.lineStarts[min(len(curState.lineStarts)-1, lastLineStart+1)]],
					Position: NewPositionFromState(curState),
				}
			}

			if curState.Input[curState.offset:curState.offset+len(s)] != s {
				// TODO: Run which is better, passing the state (all state functions without pointer)
				// or updating the state in-place (all state functions with a pointer)
				lastLineStart := curState.LineStartBeforeCurrentOffset()
				t := curState
				t.Consume(len(s))
				return Result[string]{}, Error{
					Message:  "Strings do not match.",
					Expected: s,
					Snippet:  curState.Input[curState.lineStarts[lastLineStart]:curState.lineStarts[min(len(curState.lineStarts)-1, lastLineStart+1)]],
					Got:      curState.Input[curState.offset : curState.offset+len(s)],
					Position: NewPositionFromState(curState),
				}
			}

			prev := NewPositionFromState(curState)
			curState.Consume(len(s))
			return NewResult(
				s,
				curState,
				Span{
					Start: prev,
					End:   NewPositionFromState(curState),
				}), Error{}

		},
		Label: fmt.Sprintf("The string <%s>", s),
	}
}

//TODO: Handle empty arrays for empty Parser[T] arrays as well

// the OR combinator
func Or[T any](parsers ...Parser[T]) Parser[T] {
	label := parsers[0].Label
	for _, parser := range parsers[1:] {
		label = fmt.Sprintf("%s, %s", label, parser.Label)
	}
	return Parser[T]{
		Run: func(curState State) (Result[T], Error) {
			var lastErr Error
			for _, parser := range parsers {
				res, err := parser.Run(curState) // sends a copy
				if !err.HasError() {
					return res, Error{}
				}
				lastErr = err
			}

			lastLineStart := curState.LineStartBeforeCurrentOffset()
			// furthest error with position
			return Result[T]{}, Error{
				Message:  "Or combinator failed",
				Expected: lastErr.Expected,
				Got:      lastErr.Got,
				Snippet:  curState.Input[curState.lineStarts[lastLineStart]:curState.lineStarts[min(len(curState.lineStarts)-1, lastLineStart+1)]],
				Position: lastErr.Position,
			}
		},
		Label: label,
	}
}

func And[T any](parsers ...Parser[T]) Parser[T] {
	label := parsers[0].Label
	for _, parser := range parsers[1:] {
		label = fmt.Sprintf("%s, %s", label, parser.Label)
	}
	return Parser[T]{
		Run: func(curState State) (Result[T], Error) {
			var lastRes Result[T]
			for _, parser := range parsers {
				res, err := parser.Run(curState) // sends a copy
				if err.HasError() {
					lastLineStart := curState.LineStartBeforeCurrentOffset()
					return Result[T]{}, Error{
						Message:  "And combinator failed.",
						Expected: err.Expected,
						Got:      err.Got,
						Snippet:  curState.Input[curState.lineStarts[lastLineStart]:curState.lineStarts[min(len(curState.lineStarts)-1, lastLineStart+1)]],
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

func Many0[T any](p Parser[T]) Parser[[]T] {
	return Parser[[]T]{
		Run: func(curState State) (Result[[]T], Error) {
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
				Span: Span{
					Start: NewPositionFromState(originalState),
					End:   NewPositionFromState(curState),
				},
			}, Error{}
		},
		Label: fmt.Sprintf("<%s> zero or more times.", p.Label),
	}
}

func Many1[T any](p Parser[T]) Parser[[]T] {
	return Parser[[]T]{
		Run: func(curState State) (Result[[]T], Error) {
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
					Span: Span{
						Start: NewPositionFromState(originalState),
						End:   NewPositionFromState(curState),
					},
				}, Error{}
			}

			lastLineStart := curState.LineStartBeforeCurrentOffset()
			return Result[[]T]{}, Error{
				Message:  "Many1 parser failed.",
				Expected: fmt.Sprintf("<%s> at least once", p.Label),
				Got:      fmt.Sprintf("<%s> zero times", p.Label),
				Snippet:  curState.Input[curState.lineStarts[lastLineStart]:curState.lineStarts[min(len(curState.lineStarts)-1, lastLineStart+1)]],
				Position: NewPositionFromState(curState),
			}
		},
		Label: fmt.Sprintf("<%s> one or more times.", p.Label),
	}
}

func Optional[T any](p Parser[T]) Parser[T] {
	return Parser[T]{
		Run: func(curState State) (Result[T], Error) {
			res, err := p.Run(curState)
			if err.HasError() {
				return Result[T]{}, Error{}
			}

			return res, Error{}
		},
	}
}

func Sequence[T any](parsers []Parser[T]) Parser[T] {
	label := parsers[0].Label
	for _, parser := range parsers[1:] {
		label = fmt.Sprintf("<%s> -> <%s>", label, parser.Label)
	}
	return Parser[T]{
		Run: func(curState State) (Result[T], Error) {
			var ret Result[T]
			for _, parser := range parsers {
				res, err := parser.Run(curState)
				if err.HasError() {
					lastLineStart := curState.LineStartBeforeCurrentOffset()
					return Result[T]{}, Error{
						Message:  "Sequence parser failed.",
						Expected: err.Expected,
						Got:      err.Got,
						Snippet:  curState.Input[curState.lineStarts[lastLineStart]:curState.lineStarts[min(len(curState.lineStarts)-1, lastLineStart+1)]],
						Position: NewPositionFromState(curState),
					}
				}
				ret = res
			}
			return ret, Error{}
		},
		Label: label,
	}
}

func Between[T any](open, content, close Parser[T]) Parser[T] {
	return Parser[T]{
		Run: func(curState State) (result Result[T], error Error) {
			openRes, err := open.Run(curState)
			if err.HasError() {
				lastLineStart := curState.LineStartBeforeCurrentOffset()
				return Result[T]{}, Error{
					Message:  "Between parser failed.",
					Expected: open.Label,
					Got:      err.Got,
					Snippet:  curState.Input[curState.lineStarts[lastLineStart]:curState.lineStarts[min(len(curState.lineStarts)-1, lastLineStart+1)]],
					Position: NewPositionFromState(curState),
				}
			}

			contentRes, err := content.Run(openRes.NextState)
			if err.HasError() {
				lastLineStart := curState.LineStartBeforeCurrentOffset()
				return Result[T]{}, Error{
					Message:  "Between parser failed.",
					Expected: content.Label,
					Got:      err.Got,
					Snippet:  curState.Input[curState.lineStarts[lastLineStart]:curState.lineStarts[min(len(curState.lineStarts)-1, lastLineStart+1)]],
					Position: NewPositionFromState(curState),
				}
			}

			lastLineStart := curState.LineStartBeforeCurrentOffset()
			_, err = close.Run(contentRes.NextState)
			if err.HasError() {
				return Result[T]{}, Error{
					Message:  "Between parser failed.",
					Expected: close.Label,
					Got:      err.Got,
					Snippet:  curState.Input[curState.lineStarts[lastLineStart]:curState.lineStarts[min(len(curState.lineStarts)-1, lastLineStart+1)]],
					Position: NewPositionFromState(curState),
				}
			}

			return contentRes, Error{}
		},
		Label: fmt.Sprintf("<%s> between <%s> and <%s>", content.Label, open.Label, close.Label),
	}
}

func Map[A, B any](p1 Parser[A], f func(A) B) Parser[B] {
	return Parser[B]{
		Run: func(curState State) (result Result[B], error Error) {
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
				Span: Span{
					Start: NewPositionFromState(curState),
					End:   NewPositionFromState(res.NextState),
				},
			}, Error{}
		},
	}
}
