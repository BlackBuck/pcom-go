package main

import (
	"fmt"
	"strconv"

	"github.com/BlackBuck/pcom-go/parser"
	"github.com/BlackBuck/pcom-go/state"
)

// Expression is either a single number or a sum of expressions
type Expression interface {
	Eval() int
}

type Number struct {
	Value int
}

func (n Number) Eval() int {
	return n.Value
}

type Sum struct {
	Left  Expression
	Right Expression
}

func (s Sum) Eval() int {
	return s.Left.Eval() + s.Right.Eval()
}

// Parses a single digit and converts it to Number
func parseNumber() parser.Parser[Expression] {
	digit := parser.Digit()

	return parser.Map("number", digit, func(r rune) Expression {
		val, _ := strconv.Atoi(string(r))
		return Number{Value: val}
	})
}

// Parses expressions inside parentheses
func parseParens() parser.Parser[Expression] {
	return parser.Between(
		"parentheses",
		parser.RuneParser("open paren", '('),
		parser.Lazy("inner expression", parseExpr), // Lazy to allow recursion
		parser.RuneParser("close paren", ')'),
	)
}

// Parses either a number or a parenthesized expression
func parseTerm() parser.Parser[Expression] {
	return parser.Or("term",
		parseNumber(),
		parseParens(),
	)
}

// Parses addition: expr + expr + expr
func parseExpr() parser.Parser[Expression] {
	return parser.Map("expression",
		parser.Then(
			"sum",
			parseTerm(),
			parser.Many0("additional terms",
				parser.Then("plus and term",
					parser.Lexeme(parser.RuneParser("plus", '+')),
					parseTerm(),
				),
			),
		),
		func(pair parser.Pair[Expression, []parser.Pair[rune, Expression]]) Expression {
			expr := pair.Left
			for _, p := range pair.Right {
				expr = Sum{Left: expr, Right: p.Right}
			}
			return expr
		},
	)
}

func main() {
	input := "(1+2)+(3+(4+5))"

	s := state.NewState(input, state.Position{Offset: 0, Line: 1, Column: 1})
	exprParser := parseExpr()

	res, err := exprParser.Run(&s)
	if err.HasError() {
		fmt.Println(err.FullTrace())
		return
	}

	fmt.Printf("Parsed expression evaluates to: %d\n", res.Value.Eval())
}
