package parser

import (
	"fmt"
	state "github.com/BlackBuck/pcom-go/state"
	"strings"
	"unicode/utf8"
)

// Digit parses a single digit.
func Digit() Parser[rune] {
	return CharWhere(func(r rune) bool { return r >= '0' && r <= '9' }, "Digit parser")
}

// Alphabet parses the letters a-z and A-Z.
func Alpha() Parser[rune] {
	return CharWhere(func(r rune) bool { return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') }, "Alphabet parser")
}

// AlphaNum parses alphanumeric values (single rune only)
func AlphaNum() Parser[rune] {
	alpha := Alpha()
	num := Digit()

	return Or("Alphanumeric", []Parser[rune]{alpha, num}...)
}

// Parse a whitespace
func Whitespace() Parser[rune] {
	return RuneParser("whitespace", ' ')
}

// CharWhere parses runes that satisfy a predicate
func CharWhere(predicate func(rune) bool, label string) Parser[rune] {
	return Parser[rune]{
		Run: func(curState state.State) (Result[rune], Error) {
			if !curState.InBounds(curState.Offset) {
				return Result[rune]{}, Error{
					Message:  "Char parser with predicate failed.",
					Expected: label,
					Got:      "EOF",
					Snippet:  state.GetSnippetStringFromCurrentContext(curState),
					Position: state.NewPositionFromState(curState),
				}
			}

			r, size := utf8.DecodeRuneInString(curState.Input[curState.Offset:])
			if predicate(r) {
				newState := curState
				newState.Consume(size)
				return Result[rune]{
					Value:     r,
					NextState: newState,
					Span: state.Span{
						Start: state.NewPositionFromState(curState),
						End:   state.NewPositionFromState(newState),
					},
				}, Error{}
			}
			return Result[rune]{}, Error{
				Message:  "Char parser with predicate failed.",
				Expected: label,
				Got:      string(r),
				Snippet:  state.GetSnippetStringFromCurrentContext(curState),
				Position: state.NewPositionFromState(curState),
			}
		},
		Label: label,
	}
}

// StringCI performs case-insensitive string matching.
func StringCI(s string) Parser[string] {
	lower := strings.ToLower(s)
	return Parser[string]{
		Run: func(curState state.State) (Result[string], Error) {
			if !curState.InBounds(curState.Offset + len(lower) - 1) {
				return Result[string]{}, Error{
					Message:  "Reached the end of file while parsing",
					Expected: fmt.Sprintf("String (case-insensitive) %s", s),
					Got:      "EOF",
					Snippet:  state.GetSnippetStringFromCurrentContext(curState),
					Position: state.NewPositionFromState(curState),
				}
			}

			got := curState.Input[curState.Offset : curState.Offset+len(lower)]
			if strings.ToLower(got) != lower {
				t := curState
				t.Consume(len(lower))
				return Result[string]{}, Error{
					Message:  "Strings do not match (case-insensitive).",
					Expected: fmt.Sprintf("String (case-insensitive) %s", s),
					Snippet:  state.GetSnippetStringFromCurrentContext(curState),
					Got:      curState.Input[curState.Offset : curState.Offset+len(lower)],
					Position: state.NewPositionFromState(curState),
				}
			}

			prev := state.NewPositionFromState(curState)
			curState.Consume(len(lower))
			return NewResult(
				got,
				curState,
				state.Span{
					Start: prev,
					End:   state.NewPositionFromState(curState),
				}), Error{}

		},
		Label: fmt.Sprintf("The string (case-insensitive) <%s>", s),
	}
}

// OneOf parses any one of the runes in the string.
func OneOf(chars string) Parser[rune] {
	set := make(map[rune]bool)
	for _, c := range chars {
		set[c] = true
	}

	return CharWhere(func(r rune) bool {
		return set[r]
	}, fmt.Sprintf("one of <%s>", chars))
}

// Debug prints the trace every time it runs.
func Debug[T any](p Parser[T], name string) Parser[T] {
	return Parser[T]{
		Run: func(curState state.State) (result Result[T], error Error) {
			fmt.Printf("Trying %s at position %v\n", name, state.NewPositionFromState(curState))
			res, err := p.Run(curState)
			fmt.Printf("Parser returned with\nResult: %v\nError: %v", res.Value, err)
			return res, err
		},
		Label: p.Label,
	}
}

// Try doesn't consume the state if the parser fails.
func Try[T any](p Parser[T]) Parser[T] {
	return Parser[T]{
		Run: func(curState state.State) (result Result[T], error Error) {
			prevState := curState

			res, err := p.Run(curState)
			if err.HasError() {
				return Result[T]{
					NextState: prevState,
				}, Error{}
			}

			return res, Error{}
		},
	}
}

// lexeme - a wrapper with whitespace skipping
func Lexeme[T any](p Parser[T]) Parser[T] {
	return Parser[T]{
		Label: fmt.Sprintf("lexeme <%s>", p.Label),
		Run: func(s state.State) (Result[T], Error) {
			res, err := p.Run(s)
			if err.HasError() {
				return res, err
			}
			r, err := Whitespace().Run(res.NextState) // consume trailing space

			if !err.HasError() {
				return Result[T]{
					Value:     res.Value,
					NextState: r.NextState,
					Span: state.Span{
						Start: res.Span.Start,
						End:   r.Span.End,
					},
				}, Error{}
			}

			return res, Error{}
		},
	}
}
