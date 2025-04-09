package parser

import (
	"fmt"
)

// State defines the State of the current parsing logic.
// input is the input string.
// offset is used to determine the position at which the next parser will start parsing.
type State struct {
	input  string
	offset int
}

// Result struct will be returned by all parsers alongside an error (if present).
// parsedResult is the parsed Result of the parser.
// parsedResult should've (ideally) been a generic type, but that would create unnecessary overhead.
// nextState is the State after the current parser is done parsing the input.
type Result struct {
	parsedResult interface{}
	nextState    State
}

// constructor for Result
func NewResult(parsedResult interface{}, nextState State) Result {
	return Result{parsedResult, nextState}
}

// constructor for State
func NewState(input string, offset int) State {
	return State{input, offset}
}

// The Parser type should've(ideally) been a struct, but with generic types it created a LOT of overhead.
type Parser func(curState State) (Result, error)

// Check if there are characters available for parsing
func (s State) HasAvailableChars(n int) bool {
	return s.offset < (len(s.input) - n + 1)
}

// Consume one character and return
func (s State) Consume(n int) (string, error) {
	if !s.HasAvailableChars(n) {
		return "", fmt.Errorf("not sufficient characters available for parsing")
	}

	res := s.input[s.offset : s.offset+n]
	s = s.Advance(n)

	return res, nil
}

// Peek one char -- consume without advancing
func (s State) PeekChar() (byte, error) {
	if !s.HasAvailableChars(1) {
		return 0, fmt.Errorf("not sufficient characters available for parsing")
	}

	res := s.input[s.offset]
	return res, nil
}

// Advance n places
func (s State) Advance(n int) State {
	return State{
		s.input,
		s.offset + n,
	}
}

// basic char parser
func CharParser(c byte) Parser {
	return func(curState State) (Result, error) {
		if curState.offset >= len(curState.input) {
			return NewResult(
				nil,
				curState,
			), fmt.Errorf("reached the end of input string while parsing")
		}

		if curState.input[curState.offset] != c {
			return NewResult(
				nil,
				curState,
			), fmt.Errorf("expected %c but received %c", c, curState.input[curState.offset])
		}

		return NewResult(
			string(c),
			curState.Advance(1),
		), nil
	}
}

// @params s (string).
// It parses a string exactly and advances the current State.
func String(s string) Parser {
	return func(curState State) (Result, error) {
		if curState.input[curState.offset:] != s {
			return NewResult(
				nil,
				curState,
			), fmt.Errorf("expected %s", s)
		}

		return NewResult(
			s,
			curState.Advance(len(s)),
		), nil
	}
}

// An OR combinator.
// @params left, right are Parsers.
// @returns Parser.
// performs a logical OR operation between the left and right parsers.
func Or(left Parser, right Parser) Parser {
	return func(curState State) (Result, error) {
		leftRes, err := left(curState)
		if err != nil {
			curState = leftRes.nextState
			return right(curState)
		}
		return leftRes, nil
	}
}

// An AND combinator.
// @params left, right are Parsers.
// @returns Parser.
// performs a logical AND operation between the left and right parsers.
func And(left Parser, right Parser) Parser {
	return func(curState State) (Result, error) {
		leftRes, err := left(curState)
		if err != nil {
			return NewResult(
				nil,
				curState,
			), fmt.Errorf("err in parsing \"and\"")
		}

		// Don't assign leftRes.nextState directly to curState
		// because if right parser Results in an error, we
		// won't have anything for fallback
		next := leftRes.nextState
		rightRes, err := right(next)
		if err != nil {
			return NewResult(
				nil,
				curState,
			), fmt.Errorf("err in parsing \"and\"")
		}

		return NewResult(
			[]interface{}{leftRes.parsedResult, rightRes.parsedResult},
			rightRes.nextState,
		), nil
	}
}

// @param p -> Parser.
// @param mapping -> A function.
// @returns Parser.
// It maps the output of the parser(p) through the mappping func.
func Map[A, B any](p Parser, mapping func(A) B) Parser {
	return func(curState State) (Result, error) {
		res, err := p(curState)

		if err != nil {
			return NewResult(
				nil,
				curState,
			), err
		}

		return NewResult(
			mapping(res.parsedResult.(A)),
			res.nextState,
		), nil
	}
}

// @param p -> Parser
// @returns Parser
// It checks for the presence of zero or more occurences of the parser in the input
func Many0(p Parser) Parser {
	return func(curState State) (Result, error) {
		var res []interface{}
		for curState.offset < len(curState.input) {
			x, err := p(curState)
			if err != nil {
				return NewResult(
					res,
					curState,
				), nil
			}
			res = append(res, x.parsedResult)
			curState = x.nextState
		}

		return NewResult(
			res,
			curState,
		), nil
	}
}

// @param p -> Parser.
// @returns Parser.
// It checks for the presence of one or more occurence of the parser in the input.
func Many1(p Parser) Parser {
	return func(curState State) (Result, error) {
		var res []interface{}
		for curState.offset < len(curState.input) {
			x, err := p(curState)
			if err != nil {
				if len(res) == 0 {
					return NewResult(
						nil,
						curState,
					), err
				}

				return NewResult(
					res,
					x.nextState,
				), nil
			}
			res = append(res, x.parsedResult)
			curState = x.nextState
		}

		return NewResult(
			res,
			curState,
		), nil
	}
}

// @params parsers -> Parser.
// @returns Parser.
// It sequentially parses the input.
// The output of the first parser goes as input for the second and so on.
func Seq(parsers ...Parser) Parser {
	return func(curState State) (Result, error) {
		var res []interface{}
		next := curState
		for _, parser := range parsers {
			x, err := parser(next)
			if err != nil {
				return NewResult(
					nil,
					curState, // fallback to the initial State
				), err
			}
			res = append(res, x.parsedResult)
			next = x.nextState
		}

		return NewResult(
			res,
			next,
		), nil
	}
}

// @params p -> Parser.
// @returns Parser.
// It checks for the presence of zero or one occurence of the parser in the input.
func Optional(p Parser) Parser {
	return func(curState State) (Result, error) {
		res, err := p(curState)
		if err != nil {
			return NewResult(
				nil,
				curState,
			), nil
		}

		return NewResult(
			res.parsedResult,
			res.nextState,
		), nil
	}
}

// @params open, context, close -> Parser.
// @returns Parser.
// It parses only the input that is present between open and close.
// It then returns the output produced by the context parser.
func Between(open, content, close Parser) Parser {
	return func(curState State) (Result, error) {
		openRes, err := open(curState)
		if err != nil {
			return NewResult(
				nil,
				curState,
			), err
		}

		contentRes, err := content(openRes.nextState)
		if err != nil {
			return NewResult(
				nil,
				curState,
			), err
		}

		closeRes, err := close(contentRes.nextState)
		if err != nil {
			return NewResult(
				nil,
				curState,
			), err
		}

		return NewResult(
			contentRes.parsedResult,
			closeRes.nextState,
		), nil
	}
}

// @param f -> function that returns a Parser.
// @returns Parser.
// It delays the creation of a parser unless required.
func Lazy(f func() Parser) Parser {
	var memo Parser
	return func(curState State) (Result, error) {
		if memo == nil {
			memo = f()
		}
		return memo(curState)
	}
}
