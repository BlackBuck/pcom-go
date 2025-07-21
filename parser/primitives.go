package parser

import (
	"fmt"
	"strings"
	"unicode/utf8"

	state "github.com/BlackBuck/pcom-go/state"
)

// AnyChar parses any single character.
// It is a parser that matches any rune and returns it.
// It is useful for cases where you want to match any character without any specific condition.
// Eaxmple usage: 
//   p := AnyChar()
//   result, err := p.Run(state.NewState("abc", state.Position{Offset: 0, Line: 1, Column: 1}))
//   if err.HasError() {
//       fmt.Println("Error:", err) // EOF
//   } else {
//       fmt.Println("Matched character:", result.Value)
//   }
func AnyChar() Parser[rune] {
	return CharWhere("Any character", func(r rune) bool { return true })
}

// Digit parses a single digit (0-9).
// Example usage:
//   p := Digit()
//   result, err := p.Run(state.NewState("5abc", state.Position{Offset: 0, Line: 1, Column: 1}))
//   if err.HasError() {
//       fmt.Println("Error:", err)
//   } else {
//       fmt.Println("Matched digit:", result.Value) // Output: Matched digit: 5
//   }
func Digit() Parser[rune] {
	return CharWhere("Digit parser", func(r rune) bool { return r >= '0' && r <= '9' })
}

// Alpha parses a single alphabetic character (a-z or A-Z).
// Example usage:
//   p := Alpha()
//   result, err := p.Run(state.NewState("abc", state.Position{Offset: 0, Line: 1, Column: 1}))
//   if err.HasError() {
//       fmt.Println("Error:", err)
//   } else {
//       fmt.Println("Matched letter:", result.Value) // Output: Matched letter: a
//   }
func Alpha() Parser[rune] {
	return CharWhere("Alphabet parser", func(r rune) bool { return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') })
}

// AlphaNum parses a single alphanumeric character (a-z, A-Z, or 0-9).
// Example usage:
//   p := AlphaNum()
//   result, err := p.Run(state.NewState("a1b", state.Position{Offset: 0, Line: 1, Column: 1}))
//   if err.HasError() {
//       fmt.Println("Error:", err)
//   } else {
//       fmt.Println("Matched alphanumeric:", result.Value) // Output: Matched alphanumeric: a
//   }
func AlphaNum() Parser[rune] {
	alpha := Alpha()
	num := Digit()

	return Or("Alphanumeric", []Parser[rune]{alpha, num}...)
}

// Whitespace parses a single space character (' ').
// Example usage:
//   p := Whitespace()
//   result, err := p.Run(state.NewState(" hello", state.Position{Offset: 0, Line: 1, Column: 1}))
//   if err.HasError() {
//       fmt.Println("Error:", err)
//   } else {
//       fmt.Printf("Matched whitespace: %q\n", result.Value) // Output: Matched whitespace: ' '
//   }
func Whitespace() Parser[rune] {
	return RuneParser("whitespace", ' ')
}

// CharWhere parses a single rune that satisfies the given predicate function.
// It is a generic parser for matching characters based on custom logic.
// Example usage:
//   // Match any vowel
//   p := CharWhere(func(r rune) bool {
//       return strings.ContainsRune("aeiouAEIOU", r)
//   }, "vowel")
//   result, err := p.Run(state.NewState("apple", state.Position{Offset: 0, Line: 1, Column: 1}))
//   if err.HasError() {
//       fmt.Println("Error:", err)
//   } else {
//       fmt.Println("Matched vowel:", result.Value) // Output: Matched vowel: a
//   }
func CharWhere(label string, predicate func(rune) bool) Parser[rune] {
	return Parser[rune]{
		Run: func(curState *state.State) (Result[rune], Error) {
			if !curState.InBounds(curState.Offset) {
				return Result[rune]{}, Error{
					Message:  "Char parser with predicate failed.",
					Expected: label,
					Got:      "EOF",
					Snippet:  state.GetSnippetStringFromCurrentContext(curState),
					Position: state.NewPositionFromState(curState),
				}
			}

			cp := curState.Save()
			r, size := utf8.DecodeRuneInString(curState.Input[curState.Offset:])
			if predicate(r) {
				curState.Consume(size)
				return Result[rune]{
					Value:     r,
					NextState: curState,
					Span: state.Span{
						Start: cp,
						End:   curState.Save(),
					},
				}, Error{}
			}

			curState.Rollback(cp)
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
// It returns a parser that matches the given string, ignoring case.
// Example usage:
//   p := StringCI("hello")
//   result, err := p.Run(state.NewState("HeLLo world", state.Position{Offset: 0, Line: 1, Column: 1}))
//   if err.HasError() {
//       fmt.Println("Error:", err)
//   } else {
//       fmt.Println("Matched string:", result.Value) // Output: Matched string: HeLLo
//   }
func StringCI(s string) Parser[string] {
	lower := strings.ToLower(s)
	return Parser[string]{
		Run: func(curState *state.State) (Result[string], Error) {
			if !curState.InBounds(curState.Offset + len(lower) - 1) {
				return Result[string]{}, Error{
					Message:  "Reached the end of file while parsing",
					Expected: fmt.Sprintf("String (case-insensitive) %s", s),
					Got:      "EOF",
					Snippet:  state.GetSnippetStringFromCurrentContext(curState),
					Position: state.NewPositionFromState(curState),
				}
			}

			cp := curState.Save()
			got := curState.Input[curState.Offset : curState.Offset+len(lower)]
			if strings.ToLower(got) != lower {
				return Result[string]{}, Error{
					Message:  "Strings do not match (case-insensitive).",
					Expected: fmt.Sprintf("String (case-insensitive) %s", s),
					Snippet:  state.GetSnippetStringFromCurrentContext(curState),
					Got:      curState.Input[curState.Offset : curState.Offset+len(lower)],
					Position: state.NewPositionFromState(curState),
				}
			}

			curState.Consume(len(lower))
			return NewResult(
				got,
				curState,
				state.Span{
					Start: cp,
					End:   curState.Save(),
				}), Error{}

		},
		Label: fmt.Sprintf("The string (case-insensitive) <%s>", s),
	}
}

// OneOf parses a single rune that is present in the provided string of characters.
// It returns a parser that matches any one of the specified runes.
//
// Example usage:
//   p := OneOf("abc")
//   result, err := p.Run(state.NewState("bxyz", state.Position{Offset: 0, Line: 1, Column: 1}))
//   if err.HasError() {
//       fmt.Println("Error:", err)
//   } else {
//       fmt.Printf("Matched rune: %q\n", result.Value) // Output: Matched rune: 'b'
//   }
func OneOf(chars string) Parser[rune] {
	set := make(map[rune]bool)
	for _, c := range chars {
		set[c] = true
	}

	return CharWhere(fmt.Sprintf("one of <%s>", chars), func(r rune) bool {
		return set[r]
	})
}

// Debug prints the trace every time it runs.
// It wraps a parser and logs its input position, result, and error for debugging purposes.
//
// Example usage:
//   p := Debug(Digit(), "DigitParser")
//   result, err := p.Run(state.NewState("5abc", state.Position{Offset: 0, Line: 1, Column: 1}))
//   // Output will include trace logs for the parser execution.
func Debug[T any](p Parser[T], name string) Parser[T] {
	return Parser[T]{
		Run: func(curState *state.State) (result Result[T], error Error) {
			fmt.Printf("Trying %s at position %v\n", name, state.NewPositionFromState(curState))
			res, err := p.Run(curState)
			fmt.Printf("Parser returned with\nResult: %v\nError: %v", res.Value, err)
			return res, err
		},
		Label: p.Label,
	}
}

// Try attempts to run the given parser, but if it fails, it does not consume any input (the state is rolled back).
// This is useful for backtracking: if the parser fails, parsing can continue as if nothing happened.
//
// Example usage:
//   p := Try(Digit())
//   result, err := p.Run(state.NewState("abc", state.Position{Offset: 0, Line: 1, Column: 1}))
//   if err.HasError() {
//       fmt.Println("No digit found, but input was not consumed.")
//   } else {
//       fmt.Println("Matched digit:", result.Value)
//   }
func Try[T any](p Parser[T]) Parser[T] {
	return Parser[T]{
		Run: func(curState *state.State) (result Result[T], error Error) {
			cp := curState.Save()
			res, err := p.Run(curState)
			if err.HasError() {
				curState.Rollback(cp)
				return Result[T]{
					NextState: curState,
				}, Error{}
			}

			return res, Error{}
		},
	}
}

// Lexeme wraps a parser and consumes any trailing whitespace after it.
// This is useful for token parsers where you want to ignore spaces after a token.
//
// Example usage:
//   p := Lexeme(Digit())
//   result, err := p.Run(state.NewState("5   abc", state.Position{Offset: 0, Line: 1, Column: 1}))
//   if err.HasError() {
//       fmt.Println("Error:", err)
//   } else {
//       fmt.Printf("Matched digit: %v, next input: %q\n", result.Value, result.NextState.Input[result.NextState.Offset:])
//       // Output: Matched digit: 5, next input: "abc"
//   }
func Lexeme[T any](p Parser[T]) Parser[T] {
	return Parser[T]{
		Label: fmt.Sprintf("lexeme <%s>", p.Label),
		Run: func(curState *state.State) (Result[T], Error) {
			cp := curState.Save()
			res, err := p.Run(curState)
			if err.HasError() {
				curState.Rollback(cp)
				return res, err
			}
			r, err := Whitespace().Run(res.NextState) // consume trailing space

			for !err.HasError() {
				r, err := Whitespace().Run(r.NextState)
				if err.HasError() {
					break
				}
				res.NextState = r.NextState
			}

			return res, Error{}
		},
	}
}

// TakeWhile parses a sequence of characters while the predicate function returns true.
// It continues consuming characters until the predicate returns false or the end of input is reached.
// It returns the matched string and the next state.
//// Example usage:
//   p := TakeWhile("Take while digit", func(r byte) bool {
//       return r >= '0' && r <= '9'
//   })
//   result, err := p.Run(state.NewState("123abc", state.Position{Offset: 0, Line: 1, Column: 1}))
//   if err.HasError() {
//       fmt.Println("Error:", err)
//   } else {
//	   fmt.Println("Matched digits:", result.Value) // Output: Matched digits: 123
//   }
func TakeWhile(label string, f func(byte) bool) Parser[string] {
	return Parser[string]{
		Run: func(curState *state.State) (result Result[string], error Error) {
			var ret string
			cp := curState.Save()
			for curState.InBounds(curState.Offset) && f(curState.Input[curState.Offset]) {
				r, _, _ := curState.Consume(1)
				ret += r
			}

			return Result[string]{
				Value:     ret,
				NextState: curState,
				Span: state.Span{
					Start: cp,
					End:   state.NewPositionFromState(curState),
				},
			}, Error{}
		},
		Label: label,
	}
}

// SeparatedBy parses a sequence of elements separated by a delimiter.
// It returns a slice of the parsed elements.
// The first element is parsed by the provided parser, and subsequent elements are parsed by the same parser after each delimiter.
// If the delimiter is not found, it stops parsing and returns the elements parsed so far.
//// Example usage:
//   p := SeparatedBy("Separated by comma", Digit(), CharWhere(func(r rune) bool { return r == ',' }, "comma"))	
//  result, err := p.Run(state.NewState("1,2,3", state.Position{Offset: 0, Line: 1, Column: 1}))
//  if err.HasError() {
//      fmt.Println("Error:", err)
//  } else {
// 	fmt.Println("Parsed numbers:", result.Value) // Output: Parsed numbers: [1 2 3]
// }
func SeparatedBy[A, B any](label string, p Parser[A], delimiter Parser[B]) Parser[[]A] {
	return Parser[[]A]{
		Run: func(curState *state.State) (result Result[[]A], error Error) {
			var ret []A
			cp := state.NewPositionFromState(curState)
			first, err := p.Run(curState)
			if err.HasError() {
				curState.Rollback(cp)
				return Result[[]A]{}, Error{
					Message:  "SeparatedBy failed.",
					Expected: err.Expected,
					Got:      err.Got,
					Position: err.Position,
					Snippet:  err.Snippet,
					Cause:    &err,
				}
			}

			ret = append(ret, first.Value)
			curState = first.NextState
			for {
				del, err := delimiter.Run(curState)
				if err.HasError() {
					break
				}

				res, err := p.Run(del.NextState)
				if err.HasError() {
					curState.Rollback(cp)
					return Result[[]A]{}, Error{
						Message:  "SeparatedBy failed after delimiter.",
						Expected: err.Expected,
						Got:      err.Got,
						Position: err.Position,
						Snippet:  err.Snippet,
						Cause:    &err,
					}
				}
				ret = append(ret, res.Value)
				curState = res.NextState
			}

			return Result[[]A]{
				Value:     ret,
				NextState: curState,
				Span: state.Span{
					Start: cp,
					End:   state.NewPositionFromState(curState),
				},
			}, Error{}
		},
		Label: label,
	}
}

// ManyTill parses zero or more occurrences of the parser `p` until the parser `end` succeeds.
// It returns a slice of the parsed elements.
// If `end` is not found, it continues parsing until the end of input.
// If `end` is found, it stops parsing and returns the elements parsed so far.
// Example usage:
//   p := ManyTill("Many till digit", Digit(), CharWhere("semicolon", func(r rune) bool { return r == ';' }))
//  result, err := p.Run(state.NewState("123;", state.Position{Offset: 0, Line: 1, Column: 1}))
// if err.HasError() {
//    fmt.Println("Error:", err)
// } else {
//   fmt.Println("Parsed numbers:", result.Value) // Output: Parsed numbers: [1 2 3]
// }
func ManyTill[A, B any](label string, p Parser[A], end Parser[B]) Parser[[]A] {
	return Parser[[]A]{
		Run: func(curState *state.State) (result Result[[]A], error Error) {
			var ret []A
			initialPos := state.NewPositionFromState(curState)
			for curState.InBounds(curState.Offset) {
				cp := curState.Save()
				_, err := end.Run(curState)
				if !err.HasError() {
					curState.Rollback(cp)
					return Result[[]A]{
						Value:     ret,
						NextState: curState,
						Span: state.Span{
							Start: cp,
							End:   state.NewPositionFromState(curState),
						},
					}, Error{}
				}

				res, err := p.Run(curState)
				if err.HasError() {
					curState.Rollback(cp)
					return Result[[]A]{}, Error{
						Message:  "ManyTill parser failed.",
						Expected: err.Expected,
						Got:      err.Got,
						Position: err.Position,
						Snippet:  err.Snippet,
						Cause:    &err,
					}
				}

				ret = append(ret, res.Value)
				curState = res.NextState
			}

			return Result[[]A]{
				Value:     ret,
				NextState: curState,
				Span: state.Span{
					Start: initialPos,
					End:   state.NewPositionFromState(curState),
				},
			}, Error{}
		},
		Label: label,
	}
}

// Not is a lookahead parser that succeeds only if the given parser fails at the current position.
// It does not consume any input. This is useful for preventing unwanted matches or implementing negative lookahead.
//
// Example usage:
//   p := Not("not digit", Digit())
//   result, err := p.Run(state.NewState("abc", state.Position{Offset: 0, Line: 1, Column: 1}))
//   if err.HasError() {
//       fmt.Println("Matched a digit (unexpected).")
//   } else {
//       fmt.Println("No digit found at the current position.") // Output: No digit found at the current position.
//   }
func Not[T any](label string, p Parser[T]) Parser[struct{}] {
	return Parser[struct{}]{
		Run: func(curState *state.State) (result Result[struct{}], error Error) {
			_, err := p.Run(curState)
			cp := curState.Save()
			if err.HasError() {
				curState.Rollback(cp)
				return Result[struct{}]{
					Value:     struct{}{},
					NextState: curState,
					Span: state.Span{
						Start: cp,
						End:   cp,
					},
				}, Error{}
			}

			return Result[struct{}]{}, Error{
				Message:  "Unexpected match in not.",
				Expected: "Not " + p.Label,
				Got:      p.Label,
				Position: state.NewPositionFromState(curState),
				Snippet:  state.GetSnippetStringFromCurrentContext(curState),
				Cause:    nil,
			}
		},
	}
}
