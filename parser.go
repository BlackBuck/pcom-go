package parser

import (
	"fmt"
	"unicode/utf8"
)

type State struct {
	Input string		
	offset int
	line   int
	column int
}

type Span struct {
	Start Position
	End Position
}

type Result[T any] struct {
	Value T
	NextState State
	Span Span
	Error *Error
}
type Position struct {
	Offset int // byte offset	
	Line int // line numbers - 1-indexed
	Column int // column numbers - 1-indexed
}

type Parser[T any] struct {
	Run func(state State) (result Result[T], error Error)
	Label string
} 

// TODO: change this as well
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
	return s.Position().Offset < len(s.Input) - n + 1
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
	if s.Input[s.Position().Offset] == '\r' && (len(s.Input) > s.Position().Offset + 1 && s.Input[s.Position().Offset+1] == '\n') {
		return true
	}
	return false
}

func (s *State) Position() *Position {
	return &Position{
		Offset: s.offset,
		Line: s.line,
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
			return NewResult[rune](
				0,
				curState,
				Span{
					*curState.Position(),
					*curState.Position(),
				}), Error{
					Message: "Reached the end of file while parsing",
					Expected: []string{"char"},
					Got: "EOF",
					Position: *curState.Position(),
				}
		}

		prev := curState.Position()
		curState.Advance(1)
		return NewResult(
			c,
			curState,
			Span{
				Start: *prev,
				End: *curState.Position(),
			}), Error{}
		
	},
	Label: fmt.Sprintf("the character %s", c),
	}
}

func StringParser(s string) Parser[string] {
	return Parser[string]{
		Run: func(curState State) (Result[string], Error) {
		if !curState.InBounds(curState.Position().Offset + len(s)) {
			return NewResult(
				"",
				curState,
				Span{
					*curState.Position(),
					*curState.Position(),
				}), Error{
					Message: "Reached the end of file while parsing",
					Expected: []string{"string"},
					Got: "EOF",
					Position: *curState.Position(),
				}
		}

		if curState.Input[curState.Position().Offset:curState.Position().Offset+len(s)] != s {
			// TODO: check which is better, passing the state (all state functions without pointer) 
			// or updating the state in-place (all state functions with a pointer)
			t := curState
			t.Advance(len(s))
			return NewResult[string](
				"",
				curState,
				Span{
					*curState.Position(),
					*t.Position(),
				}), Error{
					Message: "Strings do not match.",
					Expected: []string{s},
					Got: curState.Input[curState.Position().Offset:curState.Position().Offset+len(s)],
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
				End: *curState.Position(),
			}), Error{}
		
	},
	Label: fmt.Sprintf("The string <%s>", s),
	}
}

