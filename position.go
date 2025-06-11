package parser

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