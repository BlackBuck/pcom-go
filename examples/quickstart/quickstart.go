package main

import (
	"fmt"
	"github.com/BlackBuck/pcom-go/parser"
	"github.com/BlackBuck/pcom-go/state"
)

func main() {
	// A simple parser that parses the string "hello"
	helloParser := parser.StringParser("hello parser", "hello")

	input := "hello"
	s := state.NewState(input, state.Position{Offset: 0, Line: 1, Column: 1})

	result, err := helloParser.Run(&s)
	if err.HasError() {
		fmt.Println("Parse error:", err.FullTrace())
		return
	}

	fmt.Printf("Parsed value: %s\n", result.Value)
}
