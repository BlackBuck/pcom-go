package state

import (
	"unicode/utf8"

)
type Span struct {
	Start Position
	End   Position
}

type State struct {
	Input      string
	Offset     int
	Line       int
	Column     int
	LineStarts []int // offsets where newline chracters are present
}

func NewState(input string, position Position) State {
	return State{input, position.Offset, position.Line, position.Column, []int{0}}
}

func (s *State) InBounds(offset int) bool {
	return offset < len(s.Input)
}
func (s *State) HasAvailableChars(n int) bool {
	return s.Offset < len(s.Input)-n+1
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
	s.Offset = pos.Offset
	s.Column = pos.Column
	s.Line = pos.Line
}

func (s *State) UpdateColumn(n int) {
	s.Column += n
	s.UpdateOffset(n)
}

func (s *State) UpdateOffset(n int) {
	s.Offset += n
}

func (s *State) ProgressLine() {
	// CR|LF
	if isCRLF(s) {
		s.UpdateOffset(2)
	} else {
		s.UpdateOffset(1)
	}
	s.LineStarts = append(s.LineStarts, s.Offset)
	s.Line += 1
	s.Column = 1
}

func (s *State) LineStartBeforeCurrentOffset() int {
	lo, hi := 0, len(s.LineStarts)-1
	var mid int
	for lo < hi {
		mid = (hi + lo) / 2

		if s.LineStarts[mid] == s.Offset {
			return mid
		} else if s.LineStarts[mid] > s.Offset {
			hi = mid - 1
		} else {
			lo = mid + 1
		}
	}

	return hi
}

func GetSnippetStringFromCurrentContext(s State) string {
	if len(s.LineStarts) == 1 {
		return s.Input[:min(len(s.Input), s.Column)]
	}

	lastLine := s.LineStartBeforeCurrentOffset()
	return s.Input[s.LineStarts[lastLine]:s.LineStarts[min(len(s.LineStarts)-1, lastLine+1)]]
}

func isCRLF(s *State) bool {
	if s.Input[s.Offset] == '\r' && (len(s.Input) > s.Offset+1 && s.Input[s.Offset+1] == '\n') {
		return true
	}
	return false
}
