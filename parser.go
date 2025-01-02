package parser

import (
	"fmt"
)

type state struct {
	input  string
	offset int
}

type result struct {
	parsedResult interface{}
	nextState    state
}

type Parser func(curState state) (result, error)

// advance n places
func (s state) advance(n int) state {
	return state{
		s.input,
		s.offset + n,
	}
}

// basic char parser
func charParser(c byte) Parser {
	return func(curState state) (result, error) {
		if curState.offset >= len(curState.input) || curState.input[curState.offset] != c {
			return result{
				nil,
				curState,
			}, fmt.Errorf("expected %c but received %c. \ncurrent state: %v", c, curState.input[curState.offset], curState)
		}

		return result{
			string(c),
			curState.advance(1),
		}, nil
	}
}

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
