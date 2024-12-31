package main

import (
	"fmt"
	"strconv"
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
	
}

// TODO: write test for checking if the returned result
// match the expected result
func TestAndCombinator(t *testing.T) {
	left := token{'1'}
	right := token{'2'}
	and := andCombinator[uint8]{
		left, right,
	}

	expectedResult := result[any]{
		[]token{{'1'}, {'2'}},
		"3",
	}

	_, err := and.parse("123")
	if err != nil {
		t.Fatalf("AndCombinator test failed. Expected %v and error for input \"123\"", expectedResult)
	}
}

func TestMany0(t *testing.T) {

}

func TestMapCombinator(t *testing.T) {
	// Inner parser matches a single 'a'
	innerParser := token{'a'}

	// Map function converts parsed result to uppercase
	mapFunc := func(v uint8) string {
		return fmt.Sprintf("Mapped(%s)", strconv.Itoa(int(v)))
	}

	// Create mapCombinator
	mapParser := mapCombinator[uint8, string]{
		p:       innerParser,
		mapFunc: mapFunc,
	}

	// Test input
	input := "abc"

	// Expected output
	expected := result[string]{
		parsedResult: "Mapped(97)", // 'a'
		remString:    "bc",
	}

	// Run the parser
	actual, err := mapParser.parse(input)

	// Assertions
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if actual.parsedResult != expected.parsedResult {
		t.Errorf("expected parsedResult '%s', got '%s'", expected.parsedResult, actual.parsedResult)
	}

	if actual.remString != expected.remString {
		t.Errorf("expected remString '%s', got '%s'", expected.remString, actual.remString)
	}
}