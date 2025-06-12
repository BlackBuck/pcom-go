package parser

import (
	"testing"
)

func TestWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected rune
		wantErr  bool
	}{
		{
			"whitespace test 1",
			" hello world",
			' ',
			false,
		},
		{
			"whitespace test 2",
			"hello world",
			0,
			true,
		},
	}

	for _, test := range tests {
		res, err := Whitespace().Run(NewState(test.input, Position{0, 1, 1}))
		if test.wantErr {
			if !err.HasError() {
				t.Errorf("%s failed\nExpected: error\nGot: %v\n", test.name, res.Value)
			}
		} else {
			if err.HasError() {
				t.Errorf("%s failed\nExpected: %s\nGot: \n%s\n", test.name, string(test.expected), err.String())
			}

			if res.Value != test.expected {
				t.Errorf("%s failed\nExpected: %s\nGot: %s\n", test.name, string(test.expected), string(res.Value))
			}
		}
	}
}

func TestDigit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		parser   Parser[rune]
		expected rune
		hasErr   bool
	}{
		{
			"Digit test 1",
			"1234",
			Digit(),
			'1',
			false,
		},
		{
			"Digit test 2",
			"",
			Digit(),
			0,
			true,
		},
		{
			"Digit test 2",
			"abcd",
			Digit(),
			0,
			true,
		},
	}

	for _, test := range tests {
		res, err := test.parser.Run(NewState(test.input, Position{0, 1, 1}))
		if test.hasErr {
			if !err.HasError() {
				t.Errorf("%s failed\nExpected: error\nGot: %v\n", test.name, res.Value)
			}
		} else {
			if err.HasError() {
				t.Errorf("%s failed\nExpected: %s\nGot: \n%s\n", test.name, string(test.expected), err.String())
			}

			if res.Value != test.expected {
				t.Errorf("%s failed\nExpected: %s\nGot: %s\n", test.name, string(test.expected), string(res.Value))
			}
		}
	}
}

func TestAlpha(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		parser   Parser[rune]
		expected rune
		hasErr   bool
	}{
		{
			"Alphabet test 1",
			"abcd",
			Alpha(),
			'a',
			false,
		},
		{
			"Alphabet test 2",
			"1234",
			Alpha(),
			0,
			true,		
		},
		{
			"Alphabet test 3",
			"$$123",
			Alpha(),
			0,
			true,
		},	
		{
			"Alphabet test 4",
			"",
			Alpha(),
			0,
			true,
		},
	}

	for _, test := range tests {
		res, err := test.parser.Run(NewState(test.input, Position{0, 1, 1}))
		if test.hasErr {
			if !err.HasError() {
				t.Errorf("%s failed\nExpected: error\nGot: %v\n", test.name, res.Value)
			}
		} else {
			if err.HasError() {
				t.Errorf("%s failed\nExpected: %s\nGot: \n%s\n", test.name, string(test.expected), err.String())
			}

			if res.Value != test.expected {
				t.Errorf("%s failed\nExpected: %s\nGot: %s\n", test.name, string(test.expected), string(res.Value))
			}
		}
	}
}

func TestAlphaNum(t *testing.T)	{
	tests := []struct {
		name     string
		input    string
		parser   Parser[rune]
		expected rune
		hasErr   bool
	}{
		{
			"Alphanumeric test 1",
			"abcd",
			AlphaNum(),
			'a',
			false,
		},
		{
			"Alphanumeric test 2",
			"1234",
			AlphaNum(),
			'1',
			false,		
		},
		{
			"Alphanumeric test 3",
			"$$123",
			AlphaNum(),
			0,
			true,
		},	
		{
			"Alphanumeric test 4",
			"",
			AlphaNum(),
			0,
			true,
		},
	}

	for _, test := range tests {
		res, err := test.parser.Run(NewState(test.input, Position{0, 1, 1}))
		if test.hasErr {
			if !err.HasError() {
				t.Errorf("%s failed\nExpected: error\nGot: %v\n", test.name, res.Value)
			}
		} else {
			if err.HasError() {
				t.Errorf("%s failed\nExpected: %s\nGot: \n%s\n", test.name, string(test.expected), err.String())
			}

			if res.Value != test.expected {
				t.Errorf("%s failed\nExpected: %s\nGot: %s\n", test.name, string(test.expected), string(res.Value))
			}
		}
	}
}

func TestCharWhere(t *testing.T) {
	tests := []struct{
		name string
		input string
		parser Parser[rune]
		expected rune
		hasErr bool
	}{
		{
			"Predicate char test 1",
			"abcd",
			CharWhere(func(r rune) bool {return r == 'a' || r == 'b'}, "chars a or b"),
			'a',
			false,
		},
		{
			"Predicate char test 2",
			"bbcd",
			CharWhere(func(r rune) bool {return r == 'a' || r == 'b'}, "chars a or b"),
			'b',
			false,
		},
		{
			"Predicate char test 3",
			"ccdd",
			CharWhere(func(r rune) bool {return r == 'a' || r == 'b'}, "chars a or b"),
			0,
			true,
		},
		{
			"Predicate char test 4",
			"",
			CharWhere(func(r rune) bool {return r == 'a' || r == 'b'}, "chars a or b"),
			0,
			true,
		},
	}

	for _, test := range tests {
		res, err := test.parser.Run(NewState(test.input, Position{0, 1, 1}))
		if test.hasErr {
			if !err.HasError() {
				t.Errorf("%s failed\nExpected: error\nGot: %v\n", test.name, res.Value)
			}
		} else {
			if err.HasError() {
				t.Errorf("%s failed\nExpected: %s\nGot: \n%s\n", test.name, string(test.expected), err.String())
			}

			if res.Value != test.expected {
				t.Errorf("%s failed\nExpected: %s\nGot: %s\n", test.name, string(test.expected), string(res.Value))
			}
		}
	}
}

func TestStringCI(t *testing.T) {
	tests := []struct{
		name string
		input string
		parser	 Parser[string]
		expected string
		hasErr bool
	}{
		{
			"StringCI test 1",
			"AAbb",
			StringCI("aabb"),
			"AAbb",
			false,
		},
		{
			"StringCI test 2",
			"AA bb",
			StringCI("aa"),
			"AA",
			false,
		},
		{
			"StringCI test 3",
			"AbCd",
			StringCI("abcd"),
			"AbCd",
			false,
		},
		{
			"StringCI test 4",
			"",
			StringCI("a"),
			"",
			true,	
		},
		{
			"StringCI test 5",
			"Mr. Bihari",
			StringCI("Mr."),
			"Mr.",
			false,
		},
		{
			"StringCI test 6",
			"%#!$",
			StringCI("abc"),
			"",
			true,
		},
	}

	for _, test := range tests {
		res, err := test.parser.Run(NewState(test.input, Position{0, 1, 1}))
		if test.hasErr {
			if !err.HasError() {
				t.Errorf("%s failed\nExpected: error\nGot: %v\n", test.name, res.Value)
			}
		} else {
			if err.HasError() {
				t.Errorf("%s failed\nExpected: %s\nGot: \n%s\n", test.name, test.expected, err.String())
			}

			if res.Value != test.expected {
				t.Errorf("%s failed\nExpected: %s\nGot: %s\n", test.name, test.expected, string(res.Value))
			}
		}
	}
}

