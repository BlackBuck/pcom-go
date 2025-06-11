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

func testRuneParserFail(t *testing.T, input string, expected rune, parser Parser[rune]) {
	state := NewState(input, Position{0, 1, 1})
	result, err := parser.Run(state)

	if !err.HasError() {
		t.Errorf("expected error %s\ninstead got nothing", err.String())
	}

	if result.Value == expected {
		t.Errorf("Expected value '\x00', got %q", result.Value)
	}
}

func TestRuneParser_A(t *testing.T) {
	parser := RuneParser('a')
	testRuneParserPass(t, "abc", 'a', parser)
}

func TestRuneParser_B(t *testing.T) {
	parser := RuneParser('b')

	testRuneParserFail(t, "abc", 'a', parser)
}

func TestRuneParser(t *testing.T) {
	cases := []struct {
		input    string
		expected rune
		parser   Parser[rune]
	}{
		{"abc", 'a', RuneParser('a')},
		{"bcd", 'b', RuneParser('b')},
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
		{"helloworld", "hello", StringParser("hello")},
		{"Mr. Doofinsmurts", "Mr.", StringParser("Mr.")},
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
			Or(RuneParser('a'), RuneParser('b')),
			"abc",
			'a',
		},
		{
			"match second alternative",
			Or(RuneParser('x'), RuneParser('b')),
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
			[]Parser[rune]{Or(RuneParser('a'), RuneParser('b')), RuneParser('a')},
			"abc",
			'a',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewState(tt.input, Position{0, 1, 1})
			result, err := And(tt.parsers...).Run(state)
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
			Many0(RuneParser('a')),
			"aaab",
			[]rune{'a', 'a', 'a'},
		},
		{
			"zero matches",
			Many0(RuneParser('x')),
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
			Many1(RuneParser('a')),
			"aaab",
			[]rune{'a', 'a', 'a'},
			false,
		},
		{
			"no match error",
			Many1(RuneParser('x')),
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
			Between(RuneParser('('), RuneParser('x'), RuneParser(')')),
			"(x)",
			'x',
			false,
		},
		{
			"fail on missing close",
			Between(RuneParser('('), RuneParser('x'), RuneParser(')')),
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
