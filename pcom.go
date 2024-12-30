package main

import (
	"errors"
	"fmt"
)

type token struct {
	val uint8
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
				input[i:],
			}, nil
		}
	}

	return result{}, fmt.Errorf("couldn't parse string for token %c", rune(t.val))
}