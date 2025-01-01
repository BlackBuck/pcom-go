package parser

import (
	"fmt"
)

type result struct {
	parsedResult interface{}
	remString    string
}

type Parser func(input string) (result, error)

// basic char parser
func charParser(c byte) Parser {
	return func(input string) (result, error) {
		if len(input) == 0 || input[0] != c {
			return result{}, fmt.Errorf("expected %c", c)
		}

		return result{
			string(c), 
			input[1:],
		}, nil
	}
}

func String(s string) Parser{
	return func(input string) (result, error) {
		if input != s{
			return result{}, fmt.Errorf("expected %s", s)
		}

		return result{
			s,
			"",
		}, nil
	}
}

func or(left Parser, right Parser) Parser {
	return func(input string) (result, error) {
		leftRes, err := left(input)
		if err != nil {
			return right(input)
		}
		return leftRes, nil
	}
}

func and(left Parser, right Parser) Parser {
	return func(input string) (result, error) {
		leftRes, err := left(input)
		if err != nil {
			return result{}, fmt.Errorf("err in parsing \"and\"")
		}
	
		rightRes, err := right(leftRes.remString)
		if err != nil {
			return result{}, fmt.Errorf("err in parsing \"and\"")
		}

		return result{
			[]interface{}{leftRes.parsedResult, rightRes.parsedResult},
			rightRes.remString,
		}, nil
	}
}

func Map[A, B any](p Parser, mapping func(A) B) Parser {
	return func(input string) (result, error) {
		res, err := p(input)

		if err != nil{
			return result{}, err
		}

		return result{
			mapping(res.parsedResult.(A)),
			res.remString,
		}, nil
	}
}

func Many0(p Parser) Parser {
	return func(input string) (result, error) {
		var res []interface{}
		for len(input) != 0{
			x, err := p(input)
			if err != nil {
				return result{
					res,
					input,
				}, nil
			}
			res = append(res, x.parsedResult)
			input = x.remString
		}

		return result{
			res,
			input,	
		}, nil
	}
}

func Many1(p Parser) Parser {
	return func(input string) (result, error) {
		var res []interface{}
		for len(input) != 0{
			x, err := p(input)
			if err != nil{
				if len(res) == 0{
					return result{}, err
				}

				return result{
					res,
					input,
				}, nil
			}
			res = append(res, x.parsedResult)
			input = x.remString
		}

		return result{
			res,
			input,	
		}, nil
	}
}

func Seq(parsers ...Parser) Parser {
	return func (input string) (result, error) {
		var res []interface{}
		for _, parser := range parsers{
			x, err := parser(input)
			if err != nil{
				return result{}, err
			}
			res = append(res, x.parsedResult)
			input = x.remString
		}

		return result{
			res,
			input,
		}, nil
	}
}