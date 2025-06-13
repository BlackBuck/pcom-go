package parser

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func Digit() Parser[rune] {
	var ret []Parser[rune]
	for r := '0'; r <= '9'; r++ {
		ret = append(ret, RuneParser(r))
	}

	return Or(ret...)
}

func Alpha() Parser[rune] {
	var ret []Parser[rune]
	for r := 'A'; r <= 'Z'; r++ {
		ret = append(ret, RuneParser(r))
	}

	for r := 'a'; r <= 'z'; r++ {
		ret = append(ret, RuneParser(r))
	}

	return Or(ret...)
}

func AlphaNum() Parser[rune] {
	alpha := Alpha()
	num := Digit()

	return Or([]Parser[rune]{alpha, num}...)
}

func Whitespace() Parser[rune] {
	return RuneParser(' ')
}

func CharWhere(predicate func(rune) bool, label string) Parser[rune] {
	return Parser[rune]{
		Run: func(curState State) (Result[rune], Error) {
			if !curState.InBounds(curState.offset) {
				lastLineStart := curState.LineStartBeforeCurrentOffset()
				return Result[rune]{}, Error{
					Message:  "Char parser with predicate failed.",
					Expected: label,
					Got:      "EOF",
					Snippet:  curState.Input[curState.lineStarts[lastLineStart]:curState.lineStarts[min(len(curState.lineStarts)-1, lastLineStart+1)]],
					Position: NewPositionFromState(curState),
				}
			}

			r, size := utf8.DecodeRuneInString(curState.Input[curState.offset:])
			if predicate(r) {
				newState := curState
				newState.Consume(size)
				return Result[rune]{
					Value:     r,
					NextState: curState,
					Span: Span{
						Start: NewPositionFromState(curState),
						End:   NewPositionFromState(newState),
					},
				}, Error{}
			}
			lastLineStart := curState.LineStartBeforeCurrentOffset()
			return Result[rune]{}, Error{
				Message:  "Char parser with predicate failed.",
				Expected: label,
				Got:      string(r),
				Snippet:  curState.Input[curState.lineStarts[lastLineStart]:curState.lineStarts[min(len(curState.lineStarts)-1, lastLineStart+1)]],
				Position: NewPositionFromState(curState),
			}
		},
		Label: fmt.Sprintf("Char where <%s>", label),
	}
}

// case-insensitive string matching
func StringCI(s string) Parser[string] {
	lower := strings.ToLower(s)
	return Parser[string]{
		Run: func(curState State) (Result[string], Error) {
			if !curState.InBounds(curState.offset + len(lower) - 1) {
				return Result[string]{}, Error{
					Message:  "Reached the end of file while parsing",
					Expected: fmt.Sprintf("String (case-insensitive) %s", s),
					Got:      "EOF",
					Snippet:  GetSnippetStringFromCurrentContext(curState),
					Position: NewPositionFromState(curState),
				}
			}

			got := curState.Input[curState.offset : curState.offset+len(lower)]
			if strings.ToLower(got) != lower {
				t := curState
				t.Consume(len(lower))
				return Result[string]{}, Error{
					Message:  "Strings do not match (case-insensitive).",
					Expected: fmt.Sprintf("String (case-insensitive) %s", s),
					Snippet:  GetSnippetStringFromCurrentContext(curState),
					Got:      curState.Input[curState.offset : curState.offset+len(lower)],
					Position: NewPositionFromState(curState),
				}
			}

			prev := NewPositionFromState(curState)
			curState.Consume(len(lower))
			return NewResult(
				got,
				curState,
				Span{
					Start: prev,
					End:   NewPositionFromState(curState),
				}), Error{}

		},
		Label: fmt.Sprintf("The string (case-insensitive) <%s>", s),
	}
}

func OneOf(chars string) Parser[rune] {
	set := make(map[rune]bool)
	for _, c := range chars {
		set[c] = true
	}

	return CharWhere(func(r rune) bool {
		return set[r]
	}, fmt.Sprintf("one of <%s>", chars))
}

// print trace every time it runs
func Debug[T any](p Parser[T], name string) Parser[T] {
	return Parser[T]{
		Run: func(curState State) (result Result[T], error Error) {
			fmt.Printf("Trying %s at position %v\n", name, NewPositionFromState(curState))
			res, err := p.Run(curState)
			fmt.Printf("Parser returned with\nResult: %v\nError: %v", res.Value, err)
			return res, err
		},
		Label: p.Label,
	}
}

// don't consume state on failing
func Try[T any](p Parser[T]) Parser[T] {
	return Parser[T]{
		Run: func(curState State) (result Result[T], error Error) {
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
		Run: func(s State) (Result[T], Error) {
			res, err := p.Run(s)
			if err.HasError() {
				return res, err
			}
			_, _ = Whitespace().Run(res.NextState) // consume trailing space
			return res, Error{}
		},
	}
}
