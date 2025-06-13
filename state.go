package parser

import (
	"unicode/utf8"
)

type State struct {
	Input      string
	offset     int
	line       int
	column     int
	lineStarts []int // offsets where newline chracters are present
}

func NewState(input string, position Position) State {
	return State{input, position.Offset, position.Line, position.Column, []int{0}}
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
	s.UpdateOffset(n)
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
	s.lineStarts = append(s.lineStarts, s.offset)
	s.line += 1
	s.column = 1
}

func (s *State) LineStartBeforeCurrentOffset() int {
	lo, hi := 0, len(s.lineStarts)-1
	var mid int
	for lo < hi {
		mid = (hi + lo) / 2

		if s.lineStarts[mid] == s.offset {
			return mid
		} else if s.lineStarts[mid] > s.offset {
			hi = mid - 1
		} else {
			lo = mid + 1
		}
	}

	return hi
}

func GetSnippetStringFromCurrentContext(s State) string {
	if len(s.lineStarts) == 1 {
		return s.Input[:min(len(s.Input), s.column)]
	}

	lastLine := s.LineStartBeforeCurrentOffset()
	return s.Input[s.lineStarts[lastLine]:s.lineStarts[min(len(s.lineStarts)-1, lastLine+1)]]
}

func isCRLF(s *State) bool {
	if s.Input[s.offset] == '\r' && (len(s.Input) > s.offset+1 && s.Input[s.offset+1] == '\n') {
		return true
	}
	return false
}
