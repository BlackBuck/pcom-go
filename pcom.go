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

type parser struct {
	tokens 		[]token	
	op 			operationType
}

type result struct {
	parsedToken token
	remString 	string
}

func (t *token) parse(input string) (result, error) {
	if len(input) == 0 {
		return result{}, errors.New("please provide a non-empty string")
	}

	for i := 0;i < len(input);i++ {
		x := rune(input[i])
		if x == rune(t.val) {
			return result{
				*t, 
				input[i+1:],
			}, nil
		}
	}

	return result{}, fmt.Errorf("couldn't parse string for token %c", rune(t.val))
}

func (p *parser) parse(input string) (result, error) {
	switch p.op {
	case or:
		for _, token := range p.tokens {
			res, err := token.parse(input)
			if err == nil {
				return res, nil
			}
		}
		return result{}, fmt.Errorf("could not parse the input")
	default:
		return result{}, nil
	}
}