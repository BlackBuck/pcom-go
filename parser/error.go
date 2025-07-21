package parser

import (
	"fmt"

	state "github.com/BlackBuck/pcom-go/state"
	"github.com/fatih/color"
)

// Error represents an error that occurred during parsing.
// It contains a message, expected value, got value, snippet of the input string, and
// the position in the input string where the error occurred.
// It also has a cause field to chain errors together.
type Error struct {
	Message  string
	Expected string
	Got      string
	Snippet  string
	Position state.Position
	Cause    *Error
}

// HasError checks if the error has a message.
func (e *Error) HasError() bool {
	return e.Message != ""
}

// String returns a string representation of the error.
// It includes the full trace of the error, which is useful for debugging.
func (e *Error) String() string {
	res := ""
	if e.HasError() {
		res += e.FullTrace()
	}

	return res
}

// FullTrace returns the full trace of the error, including the message, position, expected and got values, and the snippet.
// It formats the error in a way that is easy to read and understand.
// It also includes the cause of the error if it exists.
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

// FormattedSnippet returns a formatted snippet of the input string where the error occurred.
// It highlights the position of the error with a caret (^) below the snippet.
// This is useful for pinpointing the exact location of the error in the input string.
func (e *Error) FormattedSnippet() string {
	res := fmt.Sprintf("%d| %s", e.Position.Line, e.Snippet)
	res += "\n"
	for range e.Position.Column + 2 {
		res += " "
	}
	res += "^ "

	return res
}
