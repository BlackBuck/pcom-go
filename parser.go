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
	return s.Position().Offset < len(s.Input)-n+1
}

func isNewLineChar(c rune) bool {
	return c == '\r' || c == '\n'
}

func (s *State) Consume(n int) (string, Span, bool) {
	startPos := s.Position()

	start := startPos.Offset
	end := start
	consumed := 0

	for consumed < n && s.InBounds(end) {
		r, size := utf8.DecodeRuneInString(s.Input[end:])

		if r == utf8.RuneError && size == 1 {
			return "", Span{}, false
		}

		if isNewLineChar(r) {
			s.progressLine()
		} else {
			s.updateColumn(1)
		}

		consumed += 1
		end += size
	}

	if consumed < n {
		return "", Span{}, false
	}

	return s.Input[start:end], Span{*startPos, *s.Position()}, true
}

func (s *State) UpdatePosition(pos Position) {
	s.offset = pos.Offset
	s.column = pos.Column
	s.line = pos.Line
}

func (s *State) updateColumn(n int) {
	s.Position().Column += n
}

func (s *State) udpateOffset(n int) {
	s.Position().Offset += n
}

func (s *State) progressLine() {
	// CR|LF
	if isCRLF(s) {
		s.Position().Offset += 2
	} else {
		s.Position().Offset += 1
	}
	s.Position().Line += 1
	s.Position().Column = 1
}

func isCRLF(s *State) bool {
	if s.Input[s.Position().Offset] == '\r' && (len(s.Input) > s.Position().Offset+1 && s.Input[s.Position().Offset+1] == '\n') {
		return true
	}
	return false
}

func (s *State) Position() *Position {
	return &Position{
		Offset: s.offset,
		Line:   s.line,
		Column: s.column,
	}
}

func (s *State) Advance(n int) {
	i := 0
	for i < n {
		if isNewLineChar(rune(s.Input[i])) {
			s.progressLine()
		} else {
			s.updateColumn(1)
		}
	}
	s.udpateOffset(n)
}

// parser a single rune
func RuneParser(c rune) Parser[rune] {
	return Parser[rune]{
		Run: func(curState State) (Result[rune], Error) {
			if !curState.InBounds(curState.Position().Offset) {
				return Result[rune]{}, Error{
					Message:  "Reached the end of file while parsing",
					Expected: string(c),
					Got:      "EOF",
					Position: *curState.Position(),
				}
			}
			if curState.Input[curState.Position().Offset] == byte(c) {
				prev := curState.Position()
				curState.Advance(1)
				return NewResult(
					c,
					curState,
					Span{
						Start: *prev,
						End:   *curState.Position(),
					}), Error{}
			}

			return Result[rune]{}, Error{
				Message:  "Reached the end of file while parsing",
				Expected: string(c),
				Got:      string(curState.Input[curState.Position().Offset]),
				Position: *curState.Position(),
			}
		},
		Label: fmt.Sprintf("the character <%s>", string(c)),
	}
}

func StringParser(s string) Parser[string] {
	return Parser[string]{
		Run: func(curState State) (Result[string], Error) {
			if !curState.InBounds(curState.Position().Offset + len(s)) {
				return Result[string]{}, Error{
					Message:  "Reached the end of file while parsing",
					Expected: s,
					Got:      "EOF",
					Position: *curState.Position(),
				}
			}

			if curState.Input[curState.Position().Offset:curState.Position().Offset+len(s)] != s {
				// TODO: Run which is better, passing the state (all state functions without pointer)
				// or updating the state in-place (all state functions with a pointer)
				t := curState
				t.Advance(len(s))
				return Result[string]{}, Error{
					Message:  "Strings do not match.",
					Expected: s,
					Got:      curState.Input[curState.Position().Offset : curState.Position().Offset+len(s)],
					Position: *curState.Position(),
				}
			}

			prev := curState.Position()
			curState.Advance(1)
			return NewResult(
				s,
				curState,
				Span{
					Start: *prev,
					End:   *curState.Position(),
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
					Start: *originalState.Position(),
					End:   *curState.Position(),
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
						Start: *originalState.Position(),
						End:   *curState.Position(),
					},
				}, Error{}
			}

			return Result[[]T]{}, Error{
				Message:  "Many1 parser failed.",
				Expected: fmt.Sprintf("<%s> at least once", p.Label),
				Got:      fmt.Sprintf("<%s> zero times", p.Label),
				Position: *curState.Position(),
			}
		},
		Label: fmt.Sprintf("<%s> one or more times.", p.Label),
	}
}

func Optional[T any](p Parser[T]) Parser[T] {
	return Parser[T]{
		Run: func (curState State) (Result[T], Error) {
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
		Run: func (curState State) (Result[T], Error) {
			var ret Result[T]
			for _, parser := range parsers {
				res, err := parser.Run(curState)
				if err.HasError() {
					return Result[T]{}, Error{
						Message: "Sequence parser failed.",
						Expected: err.Expected,
						Got: err.Got,
						Position: *curState.Position(),
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
					Message: "Between parser failed.",	
					Expected: open.Label,
					Got: err.Got,
					Position: *curState.Position(),
				}
			}

			contentRes, err := content.Run(openRes.NextState)
			if err.HasError() {
				return Result[T]{}, Error{
					Message: "Between parser failed.",	
					Expected: content.Label,
					Got: err.Got,
					Position: *curState.Position(),
				}
			}

			closeRes, err := close.Run(contentRes.NextState)
			if err.HasError() {
				return Result[T]{}, Error{
					Message: "Between parser failed.",	
					Expected: close.Label,
					Got: err.Got,
					Position: *curState.Position(),
				}
			}

			return closeRes, Error{}
		},
		Label: fmt.Sprintf("<%s> between <%s> and <%s>", content.Label, open.Label, close.Label),
	}
}