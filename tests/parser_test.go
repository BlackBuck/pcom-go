package parser_test

import (
	parser "github.com/BlackBuck/pcom-go/parser"
	state "github.com/BlackBuck/pcom-go/state"
	"testing"
)

func testRuneParserPass(t *testing.T, input string, expected rune, parser parser.Parser[rune]) {
	state := state.NewState(input, state.Position{Offset: 0, Line: 1, Column: 1})
	result, err := parser.Run(state)

	if err.HasError() {
		t.Error(err.String())
	}

	if result.Value != expected {
		t.Errorf("Expected value %q, got %q", expected, result.Value)
	}
}

func TestRuneParser_A(t *testing.T) {
	parser := parser.RuneParser("char a", 'a')
	testRuneParserPass(t, "abc", 'a', parser)
}

// func TestRuneParser_B(t *testing.T) {
// 	parser := parser.RuneParser('b')

// 	testRuneParserFail(t, "abc", 'a', parser)
// }

func TestRuneParser(t *testing.T) {
	cases := []struct {
		input    string
		expected rune
		parser   parser.Parser[rune]
	}{
		{"abc", 'a', parser.RuneParser("char a", 'a')},
		{"bcd", 'b', parser.RuneParser("char b", 'b')},
	}

	for _, c := range cases {
		testRuneParserPass(t, c.input, c.expected, c.parser)
	}
}

func TestStringParser(t *testing.T) {
	cases := []struct {
		input    string
		expected string
		parser   parser.Parser[string]
	}{
		{"helloworld", "hello", parser.StringParser("string hello", "hello")},
		{"Mr. Doofinsmurts", "Mr.", parser.StringParser("honorific", "Mr.")},
	}

	for _, c := range cases {
		res, err := c.parser.Run(state.NewState(c.input, state.Position{Offset: 0, Line: 1, Column: 1}))

		if err.HasError() {
			t.Error(err.String())
		}

		if res.Value != c.expected {
			t.Errorf("expected %s, got %s", c.expected, res.Value)
		}
	}
}

func TestOr(t *testing.T) {
	tests := []struct {
		name     string
		parser   parser.Parser[rune]
		input    string
		expected rune
	}{
		{
			"match first alternative",
			parser.Or("a or b", parser.RuneParser("char a", 'a'), parser.RuneParser("char b", 'b')),
			"abc",
			'a',
		},
		{
			"match second alternative",
			parser.Or("x or b", parser.RuneParser("char x", 'x'), parser.RuneParser("char b", 'b')),
			"bcd",
			'b',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := state.NewState(tt.input, state.Position{Offset: 0, Line: 1, Column: 1})
			result, err := tt.parser.Run(state)
			if err.HasError() {
				t.Errorf("unexpected error: %v", err)
			}
			if result.Value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.Value)
			}
		})
	}
}

func TestAnd(t *testing.T) {
	tests := []struct {
		name     string
		parsers  []parser.Parser[rune]
		input    string
		expected rune
	}{
		{
			"match all a",
			[]parser.Parser[rune]{parser.Or("a or b or c", parser.RuneParser("char a", 'a'), parser.RuneParser("char b", 'a')), parser.RuneParser("char c", 'a')},
			"abc",
			'a',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := state.NewState(tt.input, state.Position{Offset: 0, Line: 1, Column: 1})
			result, err := parser.And("And test", tt.parsers...).Run(state)
			if err.HasError() {
				t.Errorf("unexpected error: %v", err)
			}
			if result.Value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.Value)
			}
		})
	}
}

func TestMany0(t *testing.T) {
	tests := []struct {
		name     string
		parser   parser.Parser[[]rune]
		input    string
		expected []rune
	}{
		{
			"zero or more 'a'",
			parser.Many0("one or more a", parser.RuneParser("char a", 'a')),
			"aaab",
			[]rune{'a', 'a', 'a'},
		},
		{
			"zero matches",
			parser.Many0("x oncr more", parser.RuneParser("char x", 'x')),
			"abc",
			[]rune{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := state.NewState(tt.input, state.Position{Offset: 0, Line: 1, Column: 1})
			result, err := tt.parser.Run(state)
			if err.HasError() {
				t.Errorf("unexpected error: %v", err)
			}
			if len(result.Value) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(result.Value))
			}
			for i := range tt.expected {
				if result.Value[i] != tt.expected[i] {
					t.Errorf("expected %q at %d, got %q", tt.expected[i], i, result.Value[i])
				}
			}
		})
	}
}

func TestMany1(t *testing.T) {
	tests := []struct {
		name     string
		parser   parser.Parser[[]rune]
		input    string
		expected []rune
		wantErr  bool
	}{
		{
			"match many a",
			parser.Many1("a once or more", parser.RuneParser("char a", 'a')),
			"aaab",
			[]rune{'a', 'a', 'a'},
			false,
		},
		{
			"no match error",
			parser.Many1("x once or more", parser.RuneParser("char x", 'x')),
			"abc",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := state.NewState(tt.input, state.Position{Offset: 0, Line: 1, Column: 1})
			result, err := tt.parser.Run(state)
			if tt.wantErr {
				if !err.HasError() {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err.HasError() {
					t.Errorf("unexpected error: %v", err)
				}
				if len(result.Value) != len(tt.expected) {
					t.Errorf("expected length %d, got %d", len(tt.expected), len(result.Value))
				}
				for i := range tt.expected {
					if result.Value[i] != tt.expected[i] {
						t.Errorf("expected %q at %d, got %q", tt.expected[i], i, result.Value[i])
					}
				}
			}
		})
	}
}

func TestBetween(t *testing.T) {
	tests := []struct {
		name     string
		parser   parser.Parser[rune]
		input    string
		expected rune
		wantErr  bool
	}{
		{
			"match between parentheses",
			parser.Between("x in brackets", parser.RuneParser("bracket open", '('), parser.RuneParser("char x", 'x'), parser.RuneParser("bracket close", ')')),
			"(x)",
			'x',
			false,
		},
		{
			"fail on missing close",
			parser.Between("x in brackets", parser.RuneParser("bracket open", '('), parser.RuneParser("char x", 'x'), parser.RuneParser("bracket close", ')')),
			"(x",
			0,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := state.NewState(tt.input, state.Position{Offset: 0, Line: 1, Column: 1})
			result, err := tt.parser.Run(state)
			if tt.wantErr {
				if !err.HasError() {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err.HasError() {
					t.Errorf("unexpected error: %v", err)
				}
				if result.Value != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result.Value)
				}
			}
		})
	}
}

func TestThenParser(t *testing.T) {
	letter := parser.RuneParser("x", 'x')
	semicolon := parser.RuneParser(";", ';')
	expr := parser.Then("x then semicolon", letter, semicolon)
	tests := []struct {
		name     string
		input    string
		expected parser.Pair[rune, rune]
		wantErr  bool
	}{
		{
			"Then parser test 1",
			"x;",
			parser.Pair[rune, rune]{Left: 'x', Right: ';'},
			false,
		},
		{
			"Then parser test 2",
			"x",
			parser.Pair[rune, rune]{},
			true,
		},
		{
			"Then parser test 3",
			"\n",
			parser.Pair[rune, rune]{},
			true,
		},
		{
			"Then parser test 4",
			"",
			parser.Pair[rune, rune]{},
			true,
		},
	}

	for _, test := range tests {
		res, err := expr.Run(state.NewState(test.input, state.Position{Offset: 0, Line: 1, Column: 1}))
		if test.wantErr {
			if !err.HasError() {
				t.Errorf("%s failed\nexpected error, got nil\n", test.name)
			}
		} else {
			if err.HasError() {
				t.Errorf("%s failed\nExpected: %v\tGot: error\n", test.name, res.Value)
			}

			if res.Value != test.expected {
				t.Errorf("%s failed\nExpected: %v\tGot: %v\n", test.name, test.expected, res.Value)
			}
		}
	}
}

func TestKeepLeft(t *testing.T) {
	letter := parser.RuneParser("x", 'x')
	semicolon := parser.RuneParser(";", ';')
	expr := parser.KeepLeft("keep x before the semicolon", parser.Then("x then semicolon", letter, semicolon))
	tests := []struct {
		name     string
		input    string
		expected rune
		wantErr  bool
	}{
		{
			"Then parser test 1",
			"x;",
			'x',
			false,
		},
		{
			"Then parser test 2",
			"x",
			0,
			true,
		},
		{
			"Then parser test 3",
			"\n",
			0,
			true,
		},
		{
			"Then parser test 4",
			"",
			0,
			true,
		},
	}

	for _, test := range tests {
		res, err := expr.Run(state.NewState(test.input, state.Position{Offset: 0, Line: 1, Column: 1}))
		if test.wantErr {
			if !err.HasError() {
				t.Errorf("%s failed\nexpected error, got nil\n", test.name)
			}
		} else {
			if err.HasError() {
				t.Errorf("%s failed\nExpected: %v\tGot: error\n", test.name, res.Value)
			}

			if res.Value != test.expected {
				t.Errorf("%s failed\nExpected: %v\tGot: %v\n", test.name, test.expected, res.Value)
			}
		}
	}
}

func TestKeepRight(t *testing.T) {
	letter := parser.RuneParser("x", 'x')
	semicolon := parser.RuneParser(";", ';')
	expr := parser.KeepRight("keep x before the semicolon", parser.Then("x then semicolon", letter, semicolon))
	tests := []struct {
		name     string
		input    string
		expected rune
		wantErr  bool
	}{
		{
			"Then parser test 1",
			"x;",
			';',
			false,
		},
		{
			"Then parser test 2",
			"x",
			0,
			true,
		},
		{
			"Then parser test 3",
			"\n",
			0,
			true,
		},
		{
			"Then parser test 4",
			"",
			0,
			true,
		},
	}

	for _, test := range tests {
		res, err := expr.Run(state.NewState(test.input, state.Position{Offset: 0, Line: 1, Column: 1}))
		if test.wantErr {
			if !err.HasError() {
				t.Errorf("%s failed\nexpected error, got nil\n", test.name)
			}
		} else {
			if err.HasError() {
				t.Errorf("%s failed\nExpected: %v\tGot: error\n", test.name, res.Value)
			}

			if res.Value != test.expected {
				t.Errorf("%s failed\nExpected: %v\tGot: %v\n", test.name, test.expected, res.Value)
			}
		}
	}
}

func TestLazyRecursive(t *testing.T) {
	var parens parser.Parser[rune]
	letter := parser.RuneParser("x", 'x')

	parens = parser.Lazy("paren expr", func() parser.Parser[rune] {
		return parser.Or("paren expression",
			letter,
			parser.Between("in parens", parser.RuneParser("(", '('), parens, parser.RuneParser(")", ')')),
		)
	})

	tests := []struct {
		input    string
		expected rune
		wantErr  bool
	}{
		{"x", 'x', false},
		{"(x)", 'x', false},
		{"((x))\n((x))", 'x', false},
		{"(((x)))", 'x', false},
		{"((y))", 0, true},
	}

	for _, tt := range tests {
		res, err := parens.Run(state.NewState(tt.input, state.Position{Offset: 0, Line: 1, Column: 1}))
		if tt.wantErr {
			if !err.HasError() {
				t.Errorf("expected error, got result: %v", res.Value)
			}
		} else {
			if err.HasError() {
				t.Errorf("unexpected error: %s", err.String())
			}
			if res.Value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, res.Value)
			}
		}
	}
}

func TestChainl1(t *testing.T) {
	op := parser.Map("+", parser.RuneParser("+", '+'), func(r rune) func(a, b int) int { return func(a, b int) int { return a + b } })
	val := parser.Map("Rune digit to int", parser.Digit(), func(r rune) int { return int(r - '0') })
	chain := parser.Chainl1("Left-associative addition", val, op)

	tests := []struct {
		name     string
		input    string
		expected int
		hasErr   bool
	}{
		{
			"Chainl1 test 1",
			"1+2",
			3,
			false,
		},
		{
			"Chainl1 test 2",
			"1+2+9",
			12,
			false,
		},
		{
			"Chainl1 test 3",
			"1+",
			0,
			true,
		},
		{
			"Chainl1 test 4",
			"+2",
			0,
			true,
		},
		{
			"Chainl1 test 5",
			"",
			0,
			true,
		},
	}

	for _, tt := range tests {
		res, err := chain.Run(state.NewState(tt.input, state.Position{Offset: 0, Line: 1, Column: 1}))
		if tt.hasErr {
			if !err.HasError() {
				t.Errorf("expected error, got result: %v", res.Value)
			}
		} else {
			if err.HasError() {
				t.Errorf("unexpected error: %s", err.String())
			}
			if res.Value != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, res.Value)
			}
		}
	}
}

func TestChainr1(t *testing.T) {
	op := parser.Map("+", parser.RuneParser("+", '+'), func(r rune) func(a, b int) int { return func(a, b int) int { return a + b } })
	val := parser.Map("Rune digit to int", parser.Digit(), func(r rune) int { return int(r - '0') })
	chain := parser.Chainr1("Left-associative addition", val, op)

	tests := []struct {
		name     string
		input    string
		expected int
		hasErr   bool
	}{
		{
			"Chainr1 test 1",
			"1+2",
			3,
			false,
		},
		{
			"Chainr1 test 2",
			"1+2+9",
			12,
			false,
		},
		{
			"Chainr1 test 3",
			"1+",
			0,
			true,
		},
		{
			"Chainr1 test 4",
			"+2",
			0,
			true,
		},
		{
			"Chainr1 test 5",
			"",
			0,
			true,
		},
	}

	for _, tt := range tests {
		res, err := chain.Run(state.NewState(tt.input, state.Position{Offset: 0, Line: 1, Column: 1}))
		if tt.hasErr {
			if !err.HasError() {
				t.Errorf("expected error, got result: %v", res.Value)
			}
		} else {
			if err.HasError() {
				t.Errorf("unexpected error: %s", err.String())
			}
			if res.Value != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, res.Value)
			}
		}
	}
}
