package parser

import (
	"testing"

	parser "github.com/BlackBuck/pcom-go/parser"
	state "github.com/BlackBuck/pcom-go/state"
	"github.com/stretchr/testify/assert"
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

func TestLexeme(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		parser   parser.Parser[string]
		expected string
		expPos   state.Position
		hasErr   bool
	}{
		{
			"Lexeme test 1",
			"1 + 2",
			parser.Lexeme(parser.StringCI("1")),
			"1",
			state.Position{Offset: 2, Line: 1, Column: 3},
			false,
		},
		{
			"Lexeme test 2",
			"abcd efgh",
			parser.Lexeme(parser.StringCI("abcd")),
			"abcd",
			state.Position{Offset: 5, Line: 1, Column: 6},
			false,
		},
		{
			"Lexeme test 3",
			"abcd \nefgh",
			parser.Lexeme(parser.StringCI("abcd")),
			"abcd",
			state.Position{Offset: 5, Line: 1, Column: 6},
			false,
		},
	}

	for _, test := range tests {
		res, err := test.parser.Run(state.NewState(test.input, state.Position{Offset: 0, Line: 1, Column: 1}))

		if test.hasErr {
			if !err.HasError() {
				t.Errorf("%s failed\nExpected: error\nGot: %v\n", test.name, res.Value)
			}
		} else {
			assert.False(t, err.HasError(), test.name)
			assert.Equal(t, test.expected, res.Value, test.name)
			assert.Equal(t, test.expPos.Offset, res.NextState.Offset, test.name)
			assert.Equal(t, test.expPos.Line, res.NextState.Line, test.name)
			assert.Equal(t, test.expPos.Column, res.NextState.Column, test.name)
		}
	}

}

func TestSeparatedBy(t *testing.T) {
	tests := []struct{
		name     string
		input    string
		parser   parser.Parser[[]rune]
		expected []rune
		expPos   state.Position
		hasErr   bool
	}{
		{
			"SeparatedBy test 1",
			"a, B, c, D",
			parser.SeparatedBy("letters separated by comma", parser.Alpha(), parser.Lexeme(parser.RuneParser("delimiter", ','))),
			[]rune{'a', 'B', 'c', 'D'},
			state.Position{Offset: 10, Line: 1, Column: 11},
			false,
		},
		{
			"SeparatedBy test 2",
			"1, 2, 3, 4",
			parser.SeparatedBy("digits separated by comma", parser.Digit(), parser.Lexeme(parser.RuneParser("delimiter", ','))),
			[]rune{'1', '2', '3', '4'},
			state.Position{Offset: 10, Line: 1, Column: 11},
			false,
		},
		{
			"SeparatedBy test 3",
			"",
			parser.SeparatedBy("digits separated by comma", parser.Digit(), parser.Lexeme(parser.RuneParser("delimiter", ','))),
			[]rune{},
			state.Position{},
			true,
		},
		{
			"SeparatedBy test 4",
			"1,",
			parser.SeparatedBy("digits separated by comma", parser.Digit(), parser.Lexeme(parser.RuneParser("delimiter", ','))),
			[]rune{},
			state.Position{},
			true,			
		},
		{
			"SeparatedBy test 4",
			",",
			parser.SeparatedBy("digits separated by comma", parser.Digit(), parser.Lexeme(parser.RuneParser("delimiter", ','))),
			[]rune{},
			state.Position{},
			true,
		},
		{
			"SeparatedBy test 5",
			",,",
			parser.SeparatedBy("digits separated by comma", parser.Digit(), parser.Lexeme(parser.RuneParser("delimiter", ','))),
			[]rune{},
			state.Position{},
			true,
		},
		// TODO: decide if the following test should pass or fail
		// if it needs to fail, a lot of refactoring might be required
		// (then, the SeparatedBy parser must take input till it encounters a comma or EOF)

		// {
		// 	"SeparatedBy test 6",
		// 	"1, 2c,",
		// 	parser.SeparatedBy("digits separated by comma", parser.Digit(), parser.Lexeme(parser.RuneParser("delimiter", ','))),
		// 	[]rune{},
		// 	state.Position{},
		// 	true,
		// },
	}

	for _, test := range tests {
		res, err := test.parser.Run(state.NewState(test.input, state.Position{Offset: 0, Line: 1, Column: 1}))

		if test.hasErr {
			if !err.HasError() {
				t.Errorf("%s failed\nExpected: error\nGot: %v\n", test.name, res.Value)
			}
		} else {
			assert.False(t, err.HasError(), test.name)
			assert.Equal(t, test.expected, res.Value, test.name)
			assert.Equal(t, test.expPos.Offset, res.NextState.Offset, test.name)
			assert.Equal(t, test.expPos.Line, res.NextState.Line, test.name)
			assert.Equal(t, test.expPos.Column, res.NextState.Column, test.name)
		}
	}
}

func TestTakeWhile(t *testing.T) {
	tests := []struct{
		name     string
		input    string
		parser   parser.Parser[string]
		expected string
		expPos   state.Position
		hasErr   bool
	}{
		{
			"TakeWhile test 1",
			"abcD",
			parser.TakeWhile("take while letter", func(b byte) bool {return b >= 'a' && b <= 'z'}),
			"abc",
			state.Position{Offset: 3, Line: 1, Column: 4},
			false,
		},
		{
			"TakeWhile test 2",
			"1234a",
			parser.TakeWhile("take while letter", func(b byte) bool {return b >= '0' && b <= '9'}),
			"1234",
			state.Position{Offset: 4, Line: 1, Column: 5},
			false,
		},
		{
			"TakeWhile test 3",
			"1234",
			parser.TakeWhile("take while letter", func(b byte) bool {return b >= '0' && b <= '9'}),
			"1234",
			state.Position{Offset: 4, Line: 1, Column: 5},
			false,
		},
		{
			"TakeWhile test 4",
			"c1234",
			parser.TakeWhile("take while letter", func(b byte) bool {return b >= '0' && b <= '9'}),
			"",
			state.Position{Offset: 0, Line: 1, Column: 1},
			false,			
		},
	}

	for _, test := range tests {
		res, err := test.parser.Run(state.NewState(test.input, state.Position{Offset: 0, Line: 1, Column: 1}))

		if test.hasErr {
			if !err.HasError() {
				t.Errorf("%s failed\nExpected: error\nGot: %v\n", test.name, res.Value)
			}
		} else {
			assert.False(t, err.HasError(), test.name)
			assert.Equal(t, test.expected, res.Value, test.name)
			assert.Equal(t, test.expPos.Offset, res.NextState.Offset, test.name)
			assert.Equal(t, test.expPos.Line, res.NextState.Line, test.name)
			assert.Equal(t, test.expPos.Column, res.NextState.Column, test.name)
		}
	}
}

func TestManyTill(t *testing.T)	{
	tests := []struct{
		name     string
		input    string
		parser   parser.Parser[[]rune]
		expected []rune
		expPos   state.Position
		hasErr   bool
	}{
		{
			"ManyTill test 1",
			"abc,D",
			parser.ManyTill("take while letter", parser.AnyChar(), parser.RuneParser("comma", ',')),
			[]rune{'a', 'b', 'c'},
			state.Position{Offset: 3, Line: 1, Column: 4},
			false,
		},
		{
			"ManyTill test 2",
			"1234,a",
			parser.ManyTill("take while letter", parser.Digit(), parser.RuneParser("comma", ',')),
			[]rune{'1', '2', '3', '4'},
			state.Position{Offset: 4, Line: 1, Column: 5},
			false,
		},
		{
			"ManyTill test 3",
			"abcd,",
			parser.ManyTill("take while letter", parser.Digit(), parser.RuneParser("comma", ',')),
			[]rune{},
			state.Position{},
			true,
		},
		{
			"ManyTill test 4",
			"12c",
			parser.ManyTill("take while letter", parser.Digit(), parser.RuneParser("comma", ',')),
			[]rune{},
			state.Position{},
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
			assert.False(t, err.HasError(), test.name)
			assert.Equal(t, test.expected, res.Value, test.name)
			assert.Equal(t, test.expPos.Offset, res.NextState.Offset, test.name)
			assert.Equal(t, test.expPos.Line, res.NextState.Line, test.name)
			assert.Equal(t, test.expPos.Column, res.NextState.Column, test.name)
		}
	}
}