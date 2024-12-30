package main

import (
	"testing"
)

func TestParserWithoutError(t *testing.T) {
	x := &token{
		val: '1',
	}
	
	_, err := x.parse("123")
	if err != nil {
		t.Fatalf("Test failed for token %d and string %s. Expected token but received error.", x.val, "123")
	}
}

func TestParserWithError(t *testing.T) {
	x := &token{
		val: '1',
	}

	_, err := x.parse("222")	
	if err == nil {
		t.Fatalf("Test failed for token %d and string %s. Expected error but nil", x.val, "222")	
	}
}

func TestOrCombinator(t *testing.T) {
	x := &parser{
		tokens: []token{
			token{'1'},
			token{'2'},
			token{'3'},
		},
		op: or,
	}

	_, err := x.parse("123")

	if err != nil {
		t.Fatalf("OrCombinator test failed for tokens %v and string %s", x, "123")
	}
}