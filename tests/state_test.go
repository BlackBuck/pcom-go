package parser_test

import (
	"testing"

	"github.com/BlackBuck/pcom-go/state"
	"github.com/stretchr/testify/assert"
)

func TestStateConsumeBasic(t *testing.T) {
	s := state.NewState("abcdef", state.Position{Offset: 0, Line: 1, Column: 1})

	str, span, ok := s.Consume(3)
	assert.True(t, ok)
	assert.Equal(t, "abc", str)
	assert.Equal(t, 4, s.Column) // Consumed 3 chars, started at column 1
	assert.Equal(t, 3, s.Offset)
	assert.Equal(t, 1, s.Line)

	assert.Equal(t, 0, span.Start.Offset)
	assert.Equal(t, 3, span.End.Offset)
}

func TestStateConsumeBeyondInput(t *testing.T) {
	s := state.NewState("ab", state.Position{Offset: 0, Line: 1, Column: 1})

	str, span, ok := s.Consume(5)
	assert.False(t, ok)
	assert.Equal(t, "", str)
	assert.Equal(t, 0, span.Start.Offset)
	assert.Equal(t, 0, span.End.Offset)
}

func TestProgressLineUnix(t *testing.T) {
	s := state.NewState("line1\nline2", state.Position{Offset: 0, Line: 1, Column: 1})
	s.ProgressLine()

	assert.Equal(t, 2, s.Line)
	assert.Equal(t, 1, s.Column)
	assert.Equal(t, 6, s.Offset)
}

func TestProgressLineWindows(t *testing.T) {
	s := state.NewState("line1\r\nline2", state.Position{Offset: 0, Line: 1, Column: 1})
	s.ProgressLine()

	assert.Equal(t, 2, s.Line)
	assert.Equal(t, 1, s.Column)
	assert.Equal(t, 7, s.Offset) // Consumed \r\n
}

func TestLineStartBeforeCurrentOffset(t *testing.T) {
	s := state.NewState("line1\nline2\nline3", state.Position{Offset: 0, Line: 1, Column: 1})
	s.ProgressLine() // at offset 6
	s.ProgressLine() // at offset 12

	index := s.LineStartBeforeCurrentOffset()
	assert.Equal(t, 2, index) // Line starts: [0, 6, 12]
}

func TestGetSnippetStringFromCurrentContextSingleLine(t *testing.T) {
	s := state.NewState("abcdef", state.Position{Offset: 3, Line: 1, Column: 4})

	snippet := state.GetSnippetStringFromCurrentContext(s)
	assert.Equal(t, "abcdef", snippet)
}

func TestGetSnippetStringFromCurrentContextMultiLine(t *testing.T) {
	s := state.NewState("line1\nline2\nline3", state.Position{Offset: 0, Line: 1, Column: 1})

	// Progress to second line
	s.ProgressLine()
	s.UpdateOffset(2) // inside "line2"
	s.UpdateColumn(2)	

	snippet := state.GetSnippetStringFromCurrentContext(s)
	assert.Equal(t, "line2", snippet)
}

func TestEdgeCaseEmptyLineStarts(t *testing.T) {
	s := state.NewState("", state.Position{Offset: 0, Line: 1, Column: 1})
	s.LineStarts = []int{} // Force edge case

	snippet := state.GetSnippetStringFromCurrentContext(s)
	assert.Equal(t, "", snippet)
}
