package parser

import (
	"fmt"
)

type state struct {
	input 	string
	offset 	int
}

type result struct {
	parsedResult interface{}
	nextState 	 state
}

type Parser func(curState state) (result, error)

// advance n places
func (s state) advance(n int) state {
	return state{
		s.input,
		s.offset+n,
	}
}

// basic char parser
func charParser(c byte) Parser {
	return func(curState state) (result, error) {
		if curState.offset >= len(curState.input) || curState.input[curState.offset] != c {
			return result{}, fmt.Errorf("expected %c", c)
		}

		return result{
			string(c), 
			curState.advance(1),
		}, nil
	}
}

func String(s string) Parser{
	return func(curState state) (result, error) {
		if curState.input[curState.offset:] != s{
			return result{}, fmt.Errorf("expected %s", s)
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
			return result{}, fmt.Errorf("err in parsing \"and\"")
		}
		
		curState = leftRes.nextState
		rightRes, err := right(curState)
		if err != nil {
			return result{}, fmt.Errorf("err in parsing \"and\"")
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

		if err != nil{
			return result{}, err
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
		for curState.offset < len(curState.input){
			x, err := p(curState)
			if err != nil {
				return result{
					res,
					curState, // fallback
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
		for curState.offset < len(curState.input){
			x, err := p(curState)
			if err != nil{
				if len(res) == 0{
					return result{}, err
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
	return func (curState state) (result, error) {
		var res []interface{}
		for _, parser := range parsers{
			x, err := parser(curState)
			if err != nil{
				return result{}, err
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