package parser

import (
	"fmt"
	"github.com/fatih/color"
)

type Error struct {
	Message  string
	Expected string
	Got      string
	Snippet  string
	Position Position
}

func (e *Error) HasError() bool {
	return e.Message != ""
}

func (e *Error) String() string {
	res := ""
	if e.HasError() {
		res = color.RedString(e.Message + "\n")
		res += color.RedString(fmt.Sprintf("Error occured at line %d, column %d, offset %d\n", e.Position.Line, e.Position.Column, e.Position.Offset))
		res += color.HiWhiteString(e.FormattedSnippet())
		res += color.HiGreenString(fmt.Sprintf("Expected value: <%s>\nInstead got: <%s>\n", e.Expected, e.Got))
	}

	return res
}

func (e *Error) FormattedSnippet() string {
	res := e.Snippet
	res += "\n"
	for _ = range e.Position.Column {
		res += " "
	}
	res += "^"

	return res
}
