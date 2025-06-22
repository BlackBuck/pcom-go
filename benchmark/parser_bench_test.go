package parser_bench

import (
	"fmt"
	"strings"
	"testing"

	parser "github.com/BlackBuck/pcom-go/parser"
	state "github.com/BlackBuck/pcom-go/state"
)

func BenchmarkRuneParser(b *testing.B) {
	p := parser.RuneParser("char c", 'c')
	s := state.NewState("ciao", state.Position{Offset: 0, Line: 1, Column: 1})

	for i := 0; i < b.N; i++ {
		_, _ = p.Run(&s)
	}
}
func BenchmarkStringParser(b *testing.B) {
	p := parser.StringParser("hello", "hello")
	s := state.NewState("hello world", state.Position{Offset: 0, Line: 1, Column: 1})

	for i := 0; i < b.N; i++ {
		_, _ = p.Run(&s)
	}
}

func BenchmarkOrParser(b *testing.B) {
	charA := parser.RuneParser("char a", 'a')
	s := state.NewState("abcd", state.Position{Offset: 0, Line: 1, Column: 1})
	tests := []struct {
		name   string
		parser parser.Parser[rune]
	}{
		{
			"Or benchmark depth 1",
			parser.Or("no nesting", charA, charA),
		},
		{
			"Or benchmark depth 2",
			parser.Or("level 0", parser.Or("level 1", charA, charA), charA),
		},
		{
			"Or benchmark depth 3",
			parser.Or("level 0", parser.Or("level 1", parser.Or("level 2", charA, charA), charA), charA),
		},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = test.parser.Run(&s)
			}
		})
	}
}

func BenchmarkAndParser(b *testing.B) {
	charA := parser.RuneParser("char a", 'a')
	s := state.NewState("abcd", state.Position{Offset: 0, Line: 1, Column: 1})
	tests := []struct {
		name   string
		parser parser.Parser[rune]
	}{
		{
			"And benchmark depth 1",
			parser.And("no nesting", charA, charA),
		},
		{
			"And benchmark depth 2",
			parser.And("level 0", parser.And("level 1", charA, charA), charA),
		},
		{
			"And benchmark depth 3",
			parser.And("level 0", parser.And("level 1", parser.And("level 2", charA, charA), charA), charA),
		},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = test.parser.Run(&s)
			}
		})
	}
}

func BenchmarkMany0(b *testing.B) {
	charA := parser.RuneParser("char a", 'a')

	for j := range 10 {
		s := state.NewState(strings.Repeat("a", j+1), state.Position{Offset: 0, Line: 1, Column: 1})
		p := parser.Many0("0 or more a", charA)

		b.Run(fmt.Sprintf("Many0 with length %d", j+1), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = p.Run(&s)
			}
		})
	}
}

func BenchmarkMany1(b *testing.B) {
	charA := parser.RuneParser("char a", 'a')

	for j := range 10 {
		s := state.NewState(strings.Repeat("a", j+1), state.Position{Offset: 0, Line: 1, Column: 1})
		p := parser.Many0("1 or more a", charA)

		b.Run(fmt.Sprintf("Many1 with length %d", j+1), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = p.Run(&s)
			}
		})
	}
}
