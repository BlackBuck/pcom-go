package parser

import (
	"fmt"

	state "github.com/BlackBuck/pcom-go/state"
	"github.com/fatih/color"
)

type Error struct {
	Message  string
	Expected string
	Got      string
	Snippet  string
	Position state.Position
}

func (e *Error) HasError() bool {
	return e.Message != ""
}

func (e *Error) String() string {
	res := ""
	if e.HasError() {
		res = color.RedString(e.Message + "\n")
		res += color.RedString(fmt.Sprintf("Error occurred at line %d, column %d, offset %d\n", e.Position.Line, e.Position.Column, e.Position.Offset))
		res += color.HiWhiteString(e.FormattedSnippet())
		res += color.HiGreenString(fmt.Sprintf("Expected value: <%s>\t", e.Expected))
		res += color.HiRedString(fmt.Sprintf("Instead Got: <%s>\n", e.Got))
	}

	return res
}

func (e *Error) FormattedSnippet() string {
	res := fmt.Sprintf("%d| %s", e.Position.Line, e.Snippet)
	res += "\n"
	for range e.Position.Column + 2 {
		res += " "
	}
	res += "^ "

	return res
}
