package main

import "testing"

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