package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BlackBuck/pcom-go/parser"
	"github.com/BlackBuck/pcom-go/state"
)

// AST node types for arithmetic expressions
type Expr interface {
	Eval() int
}

type Number struct {
	Value int
}

func (n Number) Eval() int {
	return n.Value
}

type BinaryOp struct {
	Left  Expr
	Op    string
	Right Expr
}

func (b BinaryOp) Eval() int {
	left := b.Left.Eval()
	right := b.Right.Eval()

	switch b.Op {
	case "+":
		return left + right
	case "-":
		return left - right
	case "*":
		return left * right
	case "/":
		return left / right
	default:
		panic("Unknown operator: " + b.Op)
	}
}

// Parse a positive integer
func parseInteger() parser.Parser[int] {
	digit := parser.Digit()
	digits := parser.Many1("digits", digit)

	return parser.Map("integer", digits, func(chars []rune) int {
		numStr := strings.Join(func() []string {
			strs := make([]string, len(chars))
			for i, c := range chars {
				strs[i] = string(c)
			}
			return strs
		}(), "")

		num, err := strconv.Atoi(numStr)
		if err != nil {
			panic("Failed to convert to integer: " + numStr)
		}
		return num
	})
}

// Parse a number with optional whitespace
func parseNumber() parser.Parser[Expr] {
	return parser.Map(
		"number expression",
		parser.Lexeme(parseInteger()),
		func(n int) Expr {
			return Number{Value: n}
		},
	)
}

// Parse operators
func parseOperator() parser.Parser[string] {
	plus := parser.StringParser("plus", "+")
	minus := parser.StringParser("minus", "-")
	multiply := parser.StringParser("multiply", "*")
	divide := parser.StringParser("divide", "/")

	return parser.Or("operator", plus, minus, multiply, divide)
}

// Forward declaration for recursive parsing
var parseExpression func() parser.Parser[Expr]

// Parse a parenthesized expression
func parseParens() parser.Parser[Expr] {
	return parser.Parser[Expr]{
		Run: func(curState *state.State) (parser.Result[Expr], parser.Error) {
			// Parse opening parenthesis (with Lexeme)
			openParen := parser.Lexeme(parser.StringParser("open paren", "("))
			_, err := openParen.Run(curState)
			if err.HasError() {
				return parser.Result[Expr]{}, err
			}

			// Parse inner expression
			expr := parseExpression()
			result, err := expr.Run(curState)
			if err.HasError() {
				return parser.Result[Expr]{}, err
			}

			// Parse closing parenthesis (with Lexeme)
			closeParen := parser.Lexeme(parser.StringParser("close paren", ")"))
			_, err = closeParen.Run(curState)
			if err.HasError() {
				return parser.Result[Expr]{}, err
			}

			return result, parser.Error{}
		},
		Label: "parenthesized expression",
	}
}

// Parse a primary expression (number or parenthesized expression)
func parsePrimary() parser.Parser[Expr] {
	return parser.Or("primary", parseNumber(), parseParens())
}

// Parse a term (handles * and /)
func parseTerm() parser.Parser[Expr] {
	return parser.Parser[Expr]{
		Run: func(curState *state.State) (parser.Result[Expr], parser.Error) {
			// Parse first primary
			primary := parser.Lexeme(parsePrimary())
			left, err := primary.Run(curState)
			if err.HasError() {
				return parser.Result[Expr]{}, err
			}

			result := left.Value

			for {
				// Try to parse operator (wrapped in Lexeme)
				cp := curState.Save()
				op := parser.Lexeme(parseOperator())
				opResult, err := op.Run(curState)
				if err.HasError() {
					curState.Rollback(cp)
					break
				}

				// Only continue for * and /
				if opResult.Value != "*" && opResult.Value != "/" {
					curState.Rollback(cp)
					break
				}

				// Parse right primary (wrapped in Lexeme)
				right, err := primary.Run(curState)
				if err.HasError() {
					return parser.Result[Expr]{}, err
				}

				result = BinaryOp{
					Left:  result,
					Op:    opResult.Value,
					Right: right.Value,
				}
			}

			return parser.Result[Expr]{
				Value:     result,
				NextState: curState,
				Span:      left.Span,
			}, parser.Error{}
		},
		Label: "term",
	}
}

// Parse an expression (handles + and -)
func parseExpr() parser.Parser[Expr] {
	return parser.Parser[Expr]{
		Run: func(curState *state.State) (parser.Result[Expr], parser.Error) {
			// Parse first term (wrapped in Lexeme)
			term := parser.Lexeme(parseTerm())
			left, err := term.Run(curState)
			if err.HasError() {
				return parser.Result[Expr]{}, err
			}

			result := left.Value

			for {
				// Try to parse operator (wrapped in Lexeme)
				cp := curState.Save()
				op := parser.Lexeme(parseOperator())
				opResult, err := op.Run(curState)
				if err.HasError() {
					curState.Rollback(cp)
					break
				}

				// Only continue for + and -
				if opResult.Value != "+" && opResult.Value != "-" {
					curState.Rollback(cp)
					break
				}

				// Parse right term (wrapped in Lexeme)
				right, err := term.Run(curState)
				if err.HasError() {
					return parser.Result[Expr]{}, err
				}

				result = BinaryOp{
					Left:  result,
					Op:    opResult.Value,
					Right: right.Value,
				}
			}

			return parser.Result[Expr]{
				Value:     result,
				NextState: curState,
				Span:      left.Span,
			}, parser.Error{}
		},
		Label: "expression",
	}
}

// Initialize the forward declaration
func init() {
	parseExpression = parseExpr
}

// Main parser that handles leading whitespace
func parseArithmeticExpression() parser.Parser[Expr] {
	return parser.Lexeme(parseExpression())
}

func main() {
	// Test cases for arithmetic expressions
	testCases := []string{
		"42",
		"1 + 2",
		"3 * 4",
		"10 - 5",
		"8 / 2",
		"2 + 3 * 4",         // Should be 14 (3*4 first, then +2)
		"(2 + 3) * 4",       // Should be 20 (2+3 first, then *4)
		"10 + 2 * 3 - 4",    // Should be 12 (2*3=6, 10+6=16, 16-4=12)
		"(1 + 2) * (3 + 4)", // Should be 21 (1+2=3, 3+4=7, 3*7=21)
	}

	parser := parseArithmeticExpression()

	fmt.Println("Arithmetic Expression Parser Demo")
	fmt.Println("=================================")

	for _, testCase := range testCases {
		fmt.Printf("\nParsing: %s\n", testCase)

		// Create parser state
		s := state.NewState(testCase, state.Position{Offset: 0, Line: 1, Column: 1})

		// Parse the expression
		result, err := parser.Run(&s)
		if err.HasError() {
			fmt.Printf("  Error: %s\n", err.FullTrace())
			continue
		}

		// Check if we consumed all input
		if s.Offset < len(s.Input) {
			remaining := s.Input[s.Offset:]
			fmt.Printf("  Warning: Unparsed input remaining: '%s'\n", remaining)
		}

		// Evaluate and display result
		value := result.Value.Eval()
		fmt.Printf("  Result: %d\n", value)
	}

	// Interactive example
	fmt.Println("\n\nTry your own expressions:")
	fmt.Println("Examples: '2 + 3', '(4 + 5) * 2', '10 - 3 * 2'")

	// Note: In a real application, you would add input reading here
	// For this example, we'll just show how to use the parser programmatically
	customExpression := "15 + (3 * 2) - 1"
	fmt.Printf("\nCustom example: %s\n", customExpression)

	s := state.NewState(customExpression, state.Position{Offset: 0, Line: 1, Column: 1})
	result, err := parser.Run(&s)
	if err.HasError() {
		fmt.Printf("Error: %s\n", err.FullTrace())
	} else {
		fmt.Printf("Result: %d\n", result.Value.Eval())
	}
}
