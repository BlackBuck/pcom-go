package main

import (
	"errors"
	"fmt"
)

type operationType int

const (
	or = iota
	and
)

type token struct {
	val uint8
}

type result[T any] struct {
	parsedResult any
	remString 	 string
}

type Parser[T any] interface{
	parse(string) (result[T], error)
}

type orCombinator[T any] struct {
	left, right Parser[T]
}

type mapCombinator[A, B any] struct {
	p 		Parser[A]
	mapFunc func(A) B
}

type Many0[T any] struct {
	p Parser[T]
}

type andCombinator[T any] struct {
	left, right Parser[T]
}

func (p orCombinator[T]) parse(input string) (result[T], error) {
	res, err := p.left.parse(input);
	if err != nil {
		return p.right.parse(input);
	}
	return res, nil
}

func (p andCombinator[T]) parse(input string) (result[T], error) {
	leftRes, err := p.left.parse(input)
	if err != nil {
		return result[T]{}, err
	}
	
	rightRes, err := p.right.parse(leftRes.remString)
	
	if err != nil {
		return result[T]{}, err
	}

	res := []any{leftRes.parsedResult, rightRes.parsedResult}


	return result[T]{
		res,
		rightRes.remString,
	}, nil
}

func (p mapCombinator[A, B]) parse(input string) (result[B], error) {
	res, err := p.p.parse(input)
	if err != nil {
		return result[B]{}, nil
	}

	return result[B]{
		p.mapFunc(res.parsedResult.(A)),
		res.remString,
	}, nil
}

func (p Many0[T]) parse(input string) (result[T], error) {
	var res []any
	for len(input) != 0 {
		x, err := p.parse(input)
		input = x.remString
		res = append(res, x.parsedResult)
		if err != nil {
			return result[T]{
				res,
				input,
			}, nil
		}
	}

	return result[T]{
		res,
		"",
	}, nil
}

func (t token) parse(input string) (result[uint8], error) {
	if len(input) == 0 {
		return result[uint8]{}, errors.New("please provide a non-empty string")
	}

	if rune(input[0]) == rune(t.val) {
		return result[uint8]{
			parsedResult: t.val,
			remString:    input[1:], // Consume the matched character
		}, nil
	}

	return result[uint8]{}, fmt.Errorf("couldn't parse string for token %c", rune(t.val))
}