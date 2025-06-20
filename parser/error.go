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
	Cause    *Error
}

func (e *Error) HasError() bool {
	return e.Message != ""
}

func (e *Error) String() string {
	res := ""
	if e.HasError() {
		res += e.FullTrace()
	}

	return res
}

func (e *Error) FullTrace() string {
	trace := ""
	current := e
	for current != nil {
		trace += fmt.Sprintf(
			"%s\nAt: %s\n%s\n%s\t%s",
			color.HiRedString(current.Message),
			color.HiRedString(fmt.Sprintf("Line %d, Column %d, Offset %d", current.Position.Line, current.Position.Column, current.Position.Offset)),
			color.HiWhiteString(current.FormattedSnippet()),
			color.HiGreenString(fmt.Sprintf("Expected: %s", current.Expected)),
			color.HiRedString(fmt.Sprintf("Got: %s", current.Got)),
		)
		current = current.Cause
	}

	return trace
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
