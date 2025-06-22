package parser_bench

import (
	"testing"

	parser "github.com/BlackBuck/pcom-go/parser"
	state "github.com/BlackBuck/pcom-go/state"
)

func BenchmarkDigit(b *testing.B) {
	parser := parser.Digit()
	s := state.NewState("1234567890", state.Position{Offset: 0, Line: 1, Column: 1})
	for i := 0; i < b.N; i++ {
		_, _ = parser.Run(&s)
	}
}

func BenchmarkAlpha(b *testing.B) {
	parser := parser.Alpha()
	s := state.NewState("abcdefgXYZ", state.Position{Offset: 0, Line: 1, Column: 1})
	for i := 0; i < b.N; i++ {
		_, _ = parser.Run(&s)
	}
}

func BenchmarkAlphaNum(b *testing.B) {
	parser := parser.AlphaNum()
	s := state.NewState("a1b2c3D4E5", state.Position{Offset: 0, Line: 1, Column: 1})
	for i := 0; i < b.N; i++ {
		_, _ = parser.Run(&s)
	}
}

func BenchmarkWhitespace(b *testing.B) {
	parser := parser.Whitespace()
	s := state.NewState("     ", state.Position{Offset: 0, Line: 1, Column: 1})
	for i := 0; i < b.N; i++ {
		_, _ = parser.Run(&s)
	}
}

func BenchmarkCharWhere(b *testing.B) {
	parser := parser.CharWhere(func(r rune) bool {
		return r == 'a' || r == 'z'
	}, "a or z")
	s := state.NewState("az", state.Position{Offset: 0, Line: 1, Column: 1})
	for i := 0; i < b.N; i++ {
		_, _ = parser.Run(&s)
	}
}

func BenchmarkStringCI(b *testing.B) {
	parser := parser.StringCI("Hello")
	s := state.NewState("hElLo world", state.Position{Offset: 0, Line: 1, Column: 1})
	for i := 0; i < b.N; i++ {
		_, _ = parser.Run(&s)
	}
}
