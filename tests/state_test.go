package parser_test

import (
	"testing"

	"github.com/BlackBuck/pcom-go/state"
	"github.com/stretchr/testify/assert"
)

func TestStateConsume(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		consumeSize int
		expectOK    bool
		expectStr   string
		expectOff   int
		expectCol   int
		expectLine  int
	}{
		{
			name:        "Normal consume",
			input:       "abcdef",
			consumeSize: 3,
			expectOK:    true,
			expectStr:   "abc",
			expectOff:   3,
			expectCol:   4,
			expectLine:  1,
		},
		{
			name:        "Consume beyond input",
			input:       "ab",
			consumeSize: 5,
			expectOK:    false,
			expectStr:   "",
			expectOff:   0,
			expectCol:   1,
			expectLine:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := state.NewState(tt.input, state.Position{Offset: 0, Line: 1, Column: 1})

			str, _, ok := s.Consume(tt.consumeSize)

			assert.Equal(t, tt.expectOK, ok, tt.name)
			assert.Equal(t, tt.expectStr, str, tt.name)
			assert.Equal(t, tt.expectOff, s.Offset, tt.name)
			assert.Equal(t, tt.expectCol, s.Column, tt.name)
			assert.Equal(t, tt.expectLine, s.Line, tt.name)
		})
	}
}

func TestProgressLine(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectOffset int
		expectLine   int
		expectCol    int
	}{
		{
			name:         "Progress line with \\n",
			input:        "line1\nline2",
			expectOffset: 6,
			expectLine:   2,
			expectCol:    1,
		},
		{
			name:         "Progress line with \\r\\n",
			input:        "line1\r\nline2",
			expectOffset: 7,
			expectLine:   2,
			expectCol:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := state.NewState(tt.input, state.Position{Offset: 0, Line: 1, Column: 1})
			s.ProgressLine()

			assert.Equal(t, tt.expectOffset, s.Offset)
			assert.Equal(t, tt.expectLine, s.Line)
			assert.Equal(t, tt.expectCol, s.Column)
		})
	}
}

func TestGetSnippetStringFromCurrentContext(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		offset      int
		line        int
		column      int
		lineStarts  []int
		expected    string
		description string
	}{
		{
			name:       "Single line input",
			input:      "abcdef",
			offset:     3,
			line:       1,
			column:     4,
			lineStarts: []int{0},
			expected:   "abcdef",
		},
		{
			name:       "Multi-line input second line",
			input:      "line1\nline2\nline3",
			offset:     8, // in "line2"
			line:       2,
			column:     3,
			lineStarts: []int{0, 6, 12},
			expected:   "line2",
		},
		{
			name:       "Empty line starts edge case",
			input:      "test input",
			offset:     3,
			line:       1,
			column:     4,
			lineStarts: []int{}, // Force edge case
			expected:   "test input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := state.NewState(tt.input, state.Position{Offset: tt.offset, Line: tt.line, Column: tt.column})
			s.LineStarts = tt.lineStarts

			snippet := state.GetSnippetStringFromCurrentContext(s)

			assert.Equal(t, tt.expected, snippet)
		})
	}
}

func TestLineStartBeforeCurrentOffset(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		offsetAdvance int
		lineAdvances  int
		expectedIndex int
	}{
		{
			name:          "Second line offset",
			input:         "line1\nline2\nline3",
			offsetAdvance: 8,
			lineAdvances:  2,
			expectedIndex: 1,
		},
		{
			name:          "First line offset",
			input:         "line1\nline2\nline3",
			offsetAdvance: 2,
			lineAdvances:  0,
			expectedIndex: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := state.NewState(tt.input, state.Position{Offset: 0, Line: 1, Column: 1})

			for i := 0; i < tt.lineAdvances; i++ {
				s.ProgressLine()
			}
			s.Offset = tt.offsetAdvance

			index := s.LineStartBeforeCurrentOffset()
			assert.Equal(t, tt.expectedIndex, index)
		})
	}
}
