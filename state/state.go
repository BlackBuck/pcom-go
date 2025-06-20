package state

import (
	"strings"
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
	// precalculate LineStarts
	lineStarts := []int{0}
	for i := 0; i < len(input); {
		if input[i] == '\r' && (i+1 < len(input) && input[i+1] == '\n') {
			i += 2
			lineStarts = append(lineStarts, i)
		} else if input[i] == '\n' {
			i += 1
			lineStarts = append(lineStarts, i)
		} else {
			i += 1
		}
	}
	if len(input) == 0 {
		return State{input, position.Offset, position.Line, position.Column, []int{}}
	}

	return State{input, position.Offset, position.Line, position.Column, lineStarts}
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
		r := s.Input[end]
		if isNewLineChar(rune(r)) {
			s.ProgressLine()
		} else {
			s.UpdateColumn(1)
		}

		consumed += 1
		end += 1
	}

	if consumed < n {
		// re-trace back to original position
		s.UpdatePosition(startPos)
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
	// if called before line ends, go to that index
	// before updating line
	for s.InBounds(s.Offset) && !isCRLF(s) && s.Input[s.Offset] != '\n' {
		s.Offset += 1
	}

	// CR|LF
	if isCRLF(s) {
		s.UpdateOffset(2)
	} else {
		s.UpdateOffset(1)
	}
	s.Line += 1
	s.Column = 1
}

func (s *State) LineStartBeforeCurrentOffset() int {
	lo, hi := 0, len(s.LineStarts)-1

	for lo <= hi {
		mid := lo + (hi-lo)/2

		if s.LineStarts[mid] == s.Offset {
			return mid
		} else if s.LineStarts[mid] < s.Offset {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}

	return hi
}

func GetSnippetStringFromCurrentContext(s State) string {
	// If LineStarts is empty, fall back to entire input
	if len(s.LineStarts) == 0 {
		return s.Input
	}

	// Validate that offset is within bounds
	if s.Offset < 0 || s.Offset > len(s.Input) {
		return s.Input // fallback to entire input for invalid offset
	}

	// Find the index of the line start that comes before or at the current offset
	currentLineIndex := s.LineStartBeforeCurrentOffset()
	if currentLineIndex < 0 { // offset is before the start
		currentLineIndex = 0
	}

	lineStartOffset := s.LineStarts[currentLineIndex]

	var lineEndOffset int

	// Use LineStarts to find the end of the current line
	if currentLineIndex+1 < len(s.LineStarts) {
		// Not the last line - end is just before the next line start
		nextLineStart := s.LineStarts[currentLineIndex+1]
		lineEndOffset = nextLineStart

		// Trim the newline character(s) that caused the next line to start
		// Check for CRLF (\r\n) first, then just \n or \r
		if lineEndOffset > 0 && s.Input[lineEndOffset-1] == '\n' {
			lineEndOffset--
			if lineEndOffset > 0 && s.Input[lineEndOffset-1] == '\r' {
				lineEndOffset--
			}
		} else if lineEndOffset > 0 && s.Input[lineEndOffset-1] == '\r' {
			lineEndOffset--
		}
	} else {
		// Last line - end is the end of input
		lineEndOffset = len(s.Input)
	}

	// Ensure we don't go past the input bounds or before line start
	if lineEndOffset > len(s.Input) {
		lineEndOffset = len(s.Input)
	}
	if lineStartOffset > lineEndOffset {
		lineStartOffset = lineEndOffset
	}

	lineContent := s.Input[lineStartOffset:lineEndOffset]
	return strings.TrimRight(lineContent, "\r\n")
}

func isCRLF(s *State) bool {
	if s.Input[s.Offset] == '\r' && (len(s.Input) > s.Offset+1 && s.Input[s.Offset+1] == '\n') {
		return true
	}
	return false
}
