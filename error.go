package parser

import "fmt"

type Error struct {
	Message  string
	Expected string
	Got      string
	Position Position
}

func (e *Error) HasError() bool {
	return e.Message != ""
}

func (e *Error) String() string {
	res := ""
	if e.HasError() {
		res = e.Message + "\n"	
		res = res + fmt.Sprintf("Error occured at line %d, column %d, offset %d\n", e.Position.Line, e.Position.Column, e.Position.Offset)
		res = res + fmt.Sprintf("Expected value: <%s>\nInstead got: <%s>\n", e.Expected, e.Got)
	}

	return res
}