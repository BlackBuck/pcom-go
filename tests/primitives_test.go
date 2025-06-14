package parser

import (
	"testing"
	parser "github.com/BlackBuck/pcom-go/parser"
	state "github.com/BlackBuck/pcom-go/state"
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
		res, err := parser.Whitespace().Run(state.NewState(test.input, state.Position{Offset: 0, Line: 1, Column: 1}))
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
		parser   parser.Parser[rune]
		expected rune
		hasErr   bool
	}{
		{
			"parser.Digit test 1",
			"1234",
			parser.Digit(),
			'1',
			false,
		},
		{
			"parser.Digit test 2",
			"",
			parser.Digit(),
			0,
			true,
		},
		{
			"parser.Digit test 2",
			"abcd",
			parser.Digit(),
			0,
			true,
		},
	}

	for _, test := range tests {
		res, err := test.parser.Run(state.NewState(test.input, state.Position{Offset: 0, Line: 1, Column: 1}))
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
		parser   parser.Parser[rune]
		expected rune
		hasErr   bool
	}{
		{
			"Alphabet test 1",
			"abcd",
			parser.Alpha(),
			'a',
			false,
		},
		{
			"Alphabet test 2",
			"1234",
			parser.Alpha(),
			0,
			true,
		},
		{
			"Alphabet test 3",
			"$$123",
			parser.Alpha(),
			0,
			true,
		},
		{
			"Alphabet test 4",
			"",
			parser.Alpha(),
			0,
			true,
		},
	}

	for _, test := range tests {
		res, err := test.parser.Run(state.NewState(test.input, state.Position{Offset: 0, Line: 1, Column: 1}))
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

func TestAlphaNum(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		parser   parser.Parser[rune]
		expected rune
		hasErr   bool
	}{
		{
			"Alphanumeric test 1",
			"abcd",
			parser.AlphaNum(),
			'a',
			false,
		},
		{
			"Alphanumeric test 2",
			"1234",
			parser.AlphaNum(),
			'1',
			false,
		},
		{
			"Alphanumeric test 3",
			"$$123",
			parser.AlphaNum(),
			0,
			true,
		},
		{
			"Alphanumeric test 4",
			"",
			parser.AlphaNum(),
			0,
			true,
		},
	}

	for _, test := range tests {
		res, err := test.parser.Run(state.NewState(test.input, state.Position{Offset: 0, Line: 1, Column: 1}))
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
	tests := []struct {
		name     string
		input    string
		parser   parser.Parser[rune]
		expected rune
		hasErr   bool
	}{
		{
			"Predicate char test 1",
			"abcd",
			parser.CharWhere(func(r rune) bool { return r == 'a' || r == 'b' }, "chars a or b"),
			'a',
			false,
		},
		{
			"Predicate char test 2",
			"bbcd",
			parser.CharWhere(func(r rune) bool { return r == 'a' || r == 'b' }, "chars a or b"),
			'b',
			false,
		},
		{
			"Predicate char test 3",
			"ccdd",
			parser.CharWhere(func(r rune) bool { return r == 'a' || r == 'b' }, "chars a or b"),
			0,
			true,
		},
		{
			"Predicate char test 4",
			"",
			parser.CharWhere(func(r rune) bool { return r == 'a' || r == 'b' }, "chars a or b"),
			0,
			true,
		},
	}

	for _, test := range tests {
		res, err := test.parser.Run(state.NewState(test.input, state.Position{Offset: 0, Line: 1, Column: 1}))
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
	tests := []struct {
		name     string
		input    string
		parser   parser.Parser[string]
		expected string
		hasErr   bool
	}{
		{
			"parser.StringCI test 1",
			"AAbb",
			parser.StringCI("aabb"),
			"AAbb",
			false,
		},
		{
			"parser.StringCI test 2",
			"AA bb",
			parser.StringCI("aa"),
			"AA",
			false,
		},
		{
			"parser.StringCI test 3",
			"AbCd",
			parser.StringCI("abcd"),
			"AbCd",
			false,
		},
		{
			"parser.StringCI test 4",
			"",
			parser.StringCI("a"),
			"",
			true,
		},
		{
			"parser.StringCI test 5",
			"Mr. Bihari",
			parser.StringCI("Mr."),
			"Mr.",
			false,
		},
		{
			"parser.StringCI test 6",
			"%#!$",
			parser.StringCI("abc"),
			"",
			true,
		},
	}

	for _, test := range tests {
		res, err := test.parser.Run(state.NewState(test.input, state.Position{Offset: 0, Line: 1, Column: 1}))
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
