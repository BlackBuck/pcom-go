package parser

type Error struct {
	Message  string
	Expected string
	Got      string
	Position Position
}

func (e *Error) HasError() bool {
	return e.Message != ""
}
