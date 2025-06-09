package parser

import (
	"unicode/utf8"
)

type State struct {
	Input string		
	CurrentPos Position  // current position
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

type Parser[T any] func(state *State) (result Result[T]) 

// TODO: change this as well
func NewResult[T any](value T, nextState State, span Span) Result[T] {
	return Result[T]{value, nextState, span, nil}
}

func NewState(input string, position Position) State {
	return State{input, position}
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
	return &s.CurrentPos
}

//TODO: Finish building this
// func CharParser(c byte) Parser[byte] {
// 	return func(curState State) (Result[byte], Error) {
// 		if curState.Position.Offset >= len(curState.Input) {
// 			return NewResult(
// 				nil,
// 				curState,
// 			)
// 		}
// 	}
// }