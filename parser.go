package parser

import (
	"fmt"
)

// state defines the state of the current parsing logic.
// input is the input string
// offset is used to determine the position at which the next parser will start parsing
type state struct {
	input  string
	offset int
}

// result struct will be returned by all parsers alongside an error (if present)
// parsedResult is the parsed result of the parser.
// parsedResult should've (ideally) been a generic type, but that would create unnecessary overhead.
// nextState is the state after the current parser is done parsing the input.
type result struct {
	parsedResult interface{}
	nextState    state
}

// The Parser type should've(ideally) been a struct, but with generic types it created a LOT of overhead.
type Parser func(curState state) (result, error)

// advance n places
func (s state) advance(n int) state {
	return state{
		s.input,
		s.offset + n,
	}
}

// basic char parser
func CharParser(c byte) Parser {
	return func(curState state) (result, error) {
		if curState.offset >= len(curState.input) {
			return result{
				nil,
				curState,
			}, fmt.Errorf("reached the end of input string while parsing")
		}

		if curState.input[curState.offset] != c {
			return result{
				nil,
				curState,
			}, fmt.Errorf("expected %c but received %c", c, curState.input[curState.offset])
		}

		return result{
			string(c),
			curState.advance(1),
		}, nil
	}
}

// @params s (string)
// It parses a string exactly and advances the current state
func String(s string) Parser {
	return func(curState state) (result, error) {
		if curState.input[curState.offset:] != s {
			return result{
				nil,
				curState,
			}, fmt.Errorf("expected %s", s)
		}

		return result{
			s,
			curState.advance(len(s)),
		}, nil
	}
}

// An OR combinator
// @params left, right are Parsers
// @returns Parser
// performs a logical OR operation between the left and right parsers.
func Or(left Parser, right Parser) Parser {
	return func(curState state) (result, error) {
		leftRes, err := left(curState)
		if err != nil {
			curState = leftRes.nextState
			return right(curState)
		}
		return leftRes, nil
	}
}

// An AND combinator
// @params left, right are Parsers
// @returns Parser
// performs a logical AND operation between the left and right parsers.
func And(left Parser, right Parser) Parser {
	return func(curState state) (result, error) {
		leftRes, err := left(curState)
		if err != nil {
			return result{
				nil,
				curState,
			}, fmt.Errorf("err in parsing \"and\"")
		}

		// Don't assign leftRes.nextState directly to curState
		// because if right parser results in an error, we
		// won't have anything for fallback
		next := leftRes.nextState
		rightRes, err := right(next)
		if err != nil {
			return result{
				nil,
				curState,
			}, fmt.Errorf("err in parsing \"and\"")
		}

		return result{
			[]interface{}{leftRes.parsedResult, rightRes.parsedResult},
			rightRes.nextState,
		}, nil
	}
}

// @param p -> Parser
// @param mapping -> A function
// @returns Parser
// It maps the output of the parser(p) through the mappping func
func Map[A, B any](p Parser, mapping func(A) B) Parser {
	return func(curState state) (result, error) {
		res, err := p(curState)

		if err != nil {
			return result{
				nil,
				curState,
			}, err
		}

		return result{
			mapping(res.parsedResult.(A)),
			res.nextState,
		}, nil
	}
}

// @param p -> Parser
// @returns Parser
// It checks for the presence of zero or more occurences of the parser in the input
func Many0(p Parser) Parser {
	return func(curState state) (result, error) {
		var res []interface{}
		for curState.offset < len(curState.input) {
			x, err := p(curState)
			if err != nil {
				return result{
					res,
					curState, 
				}, nil
			}
			res = append(res, x.parsedResult)
			curState = x.nextState
		}

		return result{
			res,
			curState,
		}, nil
	}
}

// @param p -> Parser
// @returns Parser
// It checks for the presence of one or more occurence of the parser in the input
func Many1(p Parser) Parser {
	return func(curState state) (result, error) {
		var res []interface{}
		for curState.offset < len(curState.input) {
			x, err := p(curState)
			if err != nil {
				if len(res) == 0 {
					return result{
						nil,
						curState,
					}, err
				}

				return result{
					res,
					x.nextState,
				}, nil
			}
			res = append(res, x.parsedResult)
			curState = x.nextState
		}

		return result{
			res,
			curState,
		}, nil
	}
}

// @params parsers -> Parser
// @returns Parser
// It sequentially parses the input.
// The output of the first parser goes as input for the second and so on.
func Seq(parsers ...Parser) Parser {
	return func(curState state) (result, error) {
		var res []interface{}
		next := curState
		for _, parser := range parsers {
			x, err := parser(next)
			if err != nil {
				return result{
					nil,
					curState, // fallback to the initial state
				}, err
			}
			res = append(res, x.parsedResult)
			next = x.nextState
		}

		return result{
			res,
			next,
		}, nil
	}
}

// @params p -> Parser
// @returns Parser
// It checks for the presence of zero or one occurence of the parser in the input.
func Optional(p Parser) Parser {
	return func(curState state) (result, error) {
		res, err := p(curState)
		if err != nil {
			return result{
				nil,
				curState,
			}, nil
		}

		return result{
			res.parsedResult,
			res.nextState,
		}, nil
	}
}

// @params open, context, close -> Parser
// @returns Parser
// It parses only the input that is present between open and close.
// It then returns the output produced by the context parser.
func Between(open, content, close Parser) Parser {
	return func(curState state) (result, error) {
		openRes, err := open(curState)
		if err != nil {
			return result{
				nil,
				curState,
			}, err
		}

		contentRes, err := content(openRes.nextState)
		if err != nil {
			return result{
				nil,
				curState,
			}, err
		}

		closeRes, err := close(contentRes.nextState)
		if err != nil {
			return result{
				nil, 
				curState,
			}, err
		}

		return result{
			contentRes.parsedResult,
			closeRes.nextState,
		}, nil
	}
}

// @param f -> function that returns a Parser
// @returns Parser
// It delays the creation of a parser unless required.
func Lazy(f func() Parser) Parser {
	var memo Parser
	return func(curState state) (result, error) {
		if memo == nil {
			memo = f()
		} 
		return memo(curState)
	}
}
