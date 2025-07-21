package state

// Position represents a position in the input string.
// It contains the byte offset, line number, and column number.
// Example: Position{Offset: 10, Line: 2, Column: 5} means the position is at byte offset 10, on line 2, and column 5.
// Note: Line and Column are 1-indexed, meaning the first line and first column are both 1.
// This is useful for error reporting and debugging, as it allows us to pinpoint exactly where an error occurred in the input string.
type Position struct {
	Offset int // byte offset
	Line   int // line numbers - 1-indexed
	Column int // column numbers - 1-indexed
}

// NewPositionFromState creates a new Position from the current state.
// This is used to create a Position from the current state of the parser.
func NewPositionFromState(s *State) Position {
	return Position{
		Offset: s.Offset,
		Line:   s.Line,
		Column: s.Column,
	}
}
