package parser

import (
	"testing"
)

func testRuneParserPass(t *testing.T, input string, expected rune, parser Parser[rune]) {
	state := NewState(input, Position{0, 1, 1})
	result, err := parser.Run(state)

	if err.HasError() {
		t.Error(err.String())
	}

	if result.Value != expected {
		t.Errorf("Expected value %q, got %q", expected, result.Value)
	}
}

func TestRuneParser_A(t *testing.T) {
	parser := RuneParser("char a", 'a')
	testRuneParserPass(t, "abc", 'a', parser)
}

// func TestRuneParser_B(t *testing.T) {
// 	parser := RuneParser('b')

// 	testRuneParserFail(t, "abc", 'a', parser)
// }

func TestRuneParser(t *testing.T) {
	cases := []struct {
		input    string
		expected rune
		parser   Parser[rune]
	}{
		{"abc", 'a', RuneParser("char a", 'a')},
		{"bcd", 'b', RuneParser("char b", 'b')},
	}

	for _, c := range cases {
		testRuneParserPass(t, c.input, c.expected, c.parser)
	}
}

func TestStringParser(t *testing.T) {
	cases := []struct {
		input    string
		expected string
		parser   Parser[string]
	}{
		{"helloworld", "hello", StringParser("string hello", "hello")},
		{"Mr. Doofinsmurts", "Mr.", StringParser("honorific", "Mr.")},
	}

	for _, c := range cases {
		res, err := c.parser.Run(NewState(c.input, Position{0, 1, 1}))

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
		parser   Parser[rune]
		input    string
		expected rune
	}{
		{
			"match first alternative",
			Or("a or b", RuneParser("char a", 'a'), RuneParser("char b", 'b')),
			"abc",
			'a',
		},
		{
			"match second alternative",
			Or("x or b", RuneParser("char x", 'x'), RuneParser("char b", 'b')),
			"bcd",
			'b',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewState(tt.input, Position{0, 1, 1})
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
		parsers  []Parser[rune]
		input    string
		expected rune
	}{
		{
			"match all in sequence",
			[]Parser[rune]{Or("a or b or c", RuneParser("char a", 'a'), RuneParser("char b", 'b')), RuneParser("char c", 'c')},
			"abc",
			'a',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewState(tt.input, Position{0, 1, 1})
			result, err := And("And test", tt.parsers...).Run(state)
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
		parser   Parser[[]rune]
		input    string
		expected []rune
	}{
		{
			"zero or more 'a'",
			Many0("one or more a", RuneParser("char a", 'a')),
			"aaab",
			[]rune{'a', 'a', 'a'},
		},
		{
			"zero matches",
			Many0("x oncr more", RuneParser("char x", 'x')),
			"abc",
			[]rune{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewState(tt.input, Position{0, 1, 1})
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
		parser   Parser[[]rune]
		input    string
		expected []rune
		wantErr  bool
	}{
		{
			"match many a",
			Many1("a once or more", RuneParser("char a", 'a')),
			"aaab",
			[]rune{'a', 'a', 'a'},
			false,
		},
		{
			"no match error",
			Many1("x once or more", RuneParser("char x", 'x')),
			"abc",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewState(tt.input, Position{0, 1, 1})
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
		parser   Parser[rune]
		input    string
		expected rune
		wantErr  bool
	}{
		{
			"match between parentheses",
			Between("x in brackets", RuneParser("bracket open", '('), RuneParser("char x", 'x'), RuneParser("bracket close", ')')),
			"(x)",
			'x',
			false,
		},
		{
			"fail on missing close",
			Between("x in brackets", RuneParser("bracket open", '('), RuneParser("char x", 'x'), RuneParser("bracket close", ')')),
			"(x",
			0,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewState(tt.input, Position{0, 1, 1})
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
	letter := RuneParser("x", 'x')
	semicolon := RuneParser(";", ';')
	expr := Then("x then semicolon", letter, semicolon)
	tests := []struct {
		name string
		input string
		expected Pair[rune, rune]
		wantErr bool
	}{
		{
			"Then parser test 1",
			"x;",
			Pair[rune, rune]{'x', ';'},
			false,
		},
		{
			"Then parser test 2",
			"x",
			Pair[rune, rune]{},
			true,
		},
		{
			"Then parser test 3",
			"\n",
			Pair[rune, rune]{},
			true,
		},
		{
			"Then parser test 4",
			"",
			Pair[rune, rune]{},
			true,
		},
	}

	for _, test := range tests {
		res, err := expr.Run(NewState(test.input, Position{0, 1, 1}))
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
	letter := RuneParser("x", 'x')
	semicolon := RuneParser(";", ';')
	expr := KeepLeft("keep x before the semicolon", Then("x then semicolon", letter, semicolon))
	tests := []struct {
		name string
		input string
		expected rune
		wantErr bool
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
		res, err := expr.Run(NewState(test.input, Position{0, 1, 1}))
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
	letter := RuneParser("x", 'x')
	semicolon := RuneParser(";", ';')
	expr := KeepRight("keep x before the semicolon", Then("x then semicolon", letter, semicolon))
	tests := []struct {
		name string
		input string
		expected rune
		wantErr bool
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
		res, err := expr.Run(NewState(test.input, Position{0, 1, 1}))
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
	var parens Parser[rune]
	letter := RuneParser("x", 'x')

	parens = Lazy("paren expr", func() Parser[rune] {
		return Or("paren expression",
			letter,
			Between("in parens", RuneParser("(", '('), parens, RuneParser(")", ')')),
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
		res, err := parens.Run(NewState(tt.input, Position{0, 1, 1}))
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
