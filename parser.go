package parser

import (
	"fmt"
	"unicode/utf8"
)

type State struct {
	Input  string
	offset int
	line   int
	column int
}

type Span struct {
	Start Position
	End   Position
}

type Result[T any] struct {
	Value     T
	NextState State
	Span      Span
	Error     *Error
}
type Position struct {
	Offset int // byte offset
	Line   int // line numbers - 1-indexed
	Column int // column numbers - 1-indexed
}

func NewPositionFromState(s State) Position {
	return Position{
		Offset: s.offset,
		Line:   s.line,
		Column: s.column,
	}
}

type Parser[T any] struct {
	Run   func(state State) (result Result[T], error Error)
	Label string
}

func NewResult[T any](value T, nextState State, span Span) Result[T] {
	return Result[T]{value, nextState, span, nil}
}

func NewState(input string, position Position) State {
	return State{input, position.Offset, position.Line, position.Column}
}

func (s *State) InBounds(offset int) bool {
	return offset < len(s.Input)
}
func (s *State) HasAvailableChars(n int) bool {
	return s.offset < len(s.Input)-n+1
}

func isNewLineChar(c rune) bool {
	return c == '\r' || c == '\n'
}

func (s *State) Consume(n int) (string, Span, bool) {
	startPos := NewPositionFromState(*s)

	start := startPos.Offset
	end := start
	consumed := 0

	for consumed < n && s.InBounds(end) {
		r, size := utf8.DecodeRuneInString(s.Input[end:])

		if r == utf8.RuneError && size == 1 {
			return "", Span{}, false
		}

		if isNewLineChar(r) {
			s.ProgressLine()
		} else {
			s.UpdateColumn(1)
		}

		consumed += 1
		end += size
	}

	if consumed < n {
		return "", Span{}, false
	}

	return s.Input[start:end], Span{startPos, NewPositionFromState(*s)}, true
}

func (s *State) UpdatePosition(pos Position) {
	s.offset = pos.Offset
	s.column = pos.Column
	s.line = pos.Line
}

func (s *State) UpdateColumn(n int) {
	s.column += n
}

func (s *State) UpdateOffset(n int) {
	s.offset += n
}

func (s *State) ProgressLine() {
	// CR|LF
	if isCRLF(s) {
		s.UpdateOffset(2)
	} else {
		s.UpdateOffset(1)
	}
	s.line += 1
	s.column = 1
}

func isCRLF(s *State) bool {
	if s.Input[s.offset] == '\r' && (len(s.Input) > s.offset+1 && s.Input[s.offset+1] == '\n') {
		return true
	}
	return false
}

// parser a single rune
func RuneParser(c rune) Parser[rune] {
	return Parser[rune]{
		Run: func(curState State) (Result[rune], Error) {
			if !curState.InBounds(curState.offset) {
				return Result[rune]{}, Error{
					Message:  "Reached the end of file while parsing",
					Expected: string(c),
					Got:      "EOF",
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

			return Result[rune]{}, Error{
				Message:  "Reached the end of file while parsing",
				Expected: string(c),
				Got:      string(curState.Input[curState.offset]),
				Position: NewPositionFromState(curState),
			}
		},
		Label: fmt.Sprintf("the character <%s>", string(c)),
	}
}

func StringParser(s string) Parser[string] {
	return Parser[string]{
		Run: func(curState State) (Result[string], Error) {
			if !curState.InBounds(curState.offset + len(s)) {
				return Result[string]{}, Error{
					Message:  "Reached the end of file while parsing",
					Expected: s,
					Got:      "EOF",
					Position: NewPositionFromState(curState),
				}
			}

			if curState.Input[curState.offset:curState.offset+len(s)] != s {
				// TODO: Run which is better, passing the state (all state functions without pointer)
				// or updating the state in-place (all state functions with a pointer)
				t := curState
				t.Consume(len(s))
				return Result[string]{}, Error{
					Message:  "Strings do not match.",
					Expected: s,
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

			return Result[T]{}, Error{
				Message:  "Or combinator failed",
				Expected: lastErr.Expected,
				Got:      lastErr.Got,
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
					return Result[T]{}, Error{
						Message:  "And combinator failed.",
						Expected: err.Expected,
						Got:      err.Got,
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

			return Result[[]T]{}, Error{
				Message:  "Many1 parser failed.",
				Expected: fmt.Sprintf("<%s> at least once", p.Label),
				Got:      fmt.Sprintf("<%s> zero times", p.Label),
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
					return Result[T]{}, Error{
						Message:  "Sequence parser failed.",
						Expected: err.Expected,
						Got:      err.Got,
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
				return Result[T]{}, Error{
					Message:  "Between parser failed.",
					Expected: open.Label,
					Got:      err.Got,
					Position: NewPositionFromState(curState),
				}
			}

			contentRes, err := content.Run(openRes.NextState)
			if err.HasError() {
				return Result[T]{}, Error{
					Message:  "Between parser failed.",
					Expected: content.Label,
					Got:      err.Got,
					Position: NewPositionFromState(curState),
				}
			}

			closeRes, err := close.Run(contentRes.NextState)
			if err.HasError() {
				return Result[T]{}, Error{
					Message:  "Between parser failed.",
					Expected: close.Label,
					Got:      err.Got,
					Position: NewPositionFromState(curState),
				}
			}

			return closeRes, Error{}
		},
		Label: fmt.Sprintf("<%s> between <%s> and <%s>", content.Label, open.Label, close.Label),
	}
}
