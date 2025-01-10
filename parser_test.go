package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCharParserPass(t *testing.T) {
	c := byte('c')
	CharParser := CharParser(c)
	curState := state{
		"chillin",
		0,
	}

	res, err := CharParser(curState)

	exp := result{
		"c",
		state{
			"chillin",
			1,	
		},
	}

	if err != nil {
		t.Fatalf("char parsing test failed. Expected %v got the following error:\n %s", exp, err.Error())
	}

	if !reflect.DeepEqual(exp, res) {
		t.Fatalf("char parsing test failed. Expected %v but got %v.", exp, res)
	}
}

func TestCharParserFail(t *testing.T) {
	c := byte('d')
	CharParser := CharParser(c)
	curState := state{
		"chillin",
		0,
	}

	res, err := CharParser(curState)

	exp := result{
		nil,
		curState,
	}

	if err == nil {
		t.Fatalf("char parsing test failed. Expected err got %v", exp)
	}

	if !reflect.DeepEqual(exp, res) {
		t.Fatalf("char parsing test failed. Expected %v but got %v.", exp, res)
	}
}

func TestOrCombinatorPass(t *testing.T) {
	t1 := byte('c')
	t2 := byte('1')
	curState := state{
		"c23",
		0,
	}
	charOrDigit := Or(CharParser(t1), CharParser(t2))
	res, err := charOrDigit(curState)

	exp := result{
		"c",
		state{
			"c23", 
			1,
		},
	}
	
	if err != nil {
		t.Fatalf("Or combinator test failed. Expected %v got the following error:\n %s.", exp, err.Error())
	}

	if !reflect.DeepEqual(exp, res) {
		t.Fatalf("Or combinator test failed. Expected %v got %v", exp, res)
	}
}

func TestOrCombinatorFail(t *testing.T) {
	t1 := byte('c')
	t2 := byte('1')
	curState := state{
		"d23",
		0,
	}
	charOrDigit := Or(CharParser(t1), CharParser(t2))
	res, err := charOrDigit(curState)

	exp := result{
		nil,
		curState,
	}
	
	if err == nil {
		t.Fatalf("Or combinator test failed. Expected error got %v.", exp)
	}

	if !reflect.DeepEqual(exp, res) {
		t.Fatalf("Or combinator test failed. Expected %v got %v", exp, res)
	}
}

func TestAndCombinatorPass(t *testing.T) {
	t1 := byte('c')
	t2 := byte('1')
	curState := state{
		"c13",
		0,
	}
	charAndDigit := And(CharParser(t1), CharParser(t2))
	res, err := charAndDigit(curState)

	// This is a work-around for the DeepEqual function because, 
	// somehow, the array I initialised wasn't interned (string interning)
	// and the string pointers were different and they weren't deeply equal :<
	exp := result{
		res.parsedResult, 
		state{
			"c13", 
			2,
		},
	}
	
	if err != nil {
		t.Fatalf("Or combinator test failed. Expected %v got the following error:\n %s.", exp, err.Error())
	}

	if !reflect.DeepEqual(exp, res) {
		t.Fatalf("Or combinator test failed. Expected %v got %v", exp, res)
	}
}

func TestAndCombinatorFail(t *testing.T) {
	t1 := byte('c')
	t2 := byte('1')
	curState := state{
		"d23",
		0,
	}
	charAndDigit := And(CharParser(t1), CharParser(t2))
	res, err := charAndDigit(curState)

	exp := result{
		nil,
		curState,
	}
	
	if err == nil {
		t.Fatalf("Or combinator test failed. Expected %v got error.", exp)
	}

	if !reflect.DeepEqual(exp, res) {
		t.Fatalf("Or combinator test failed. Expected %v got %v", exp, res)
	}
}

func TestMapCombinatorPass(t *testing.T) {
	ch := CharParser('1')
	mapfunc := func(ch string) string{
		return fmt.Sprintf("Mapped (%s)", ch)
	}
	curState := state{
		"123",
		0,
	}

	mapcomb := Map(ch, mapfunc)
	res, err := mapcomb(curState)

	exp := result{
		"Mapped (1)",
		state{
			"123",
			1,
		},
	}

	if err != nil {
		t.Fatalf("Map test failed. Expected %v but received %v and the following error:\n %s", exp, res, err.Error())
	}

	if !reflect.DeepEqual(exp, res) {
		t.Fatalf("Map combinator test failed. Expected %v but received %v", exp, res)
	}
}

func TestMapCombinatorFail(t *testing.T) {
	ch := CharParser('2')
	mapfunc := func(ch string) string{
		return fmt.Sprintf("Mapped (%s)", ch)
	}
	curState := state{
		"123",
		0,
	}

	mapcomb := Map(ch, mapfunc)
	res, err := mapcomb(curState)

	exp := result{
		nil,
		curState,
	}	

	if err == nil {
		t.Fatalf("Map test failed. Expected error but received %v", exp)
	}

	if !reflect.DeepEqual(exp, res) {
		t.Fatalf("Map combinator test failed. Expected %v but received %v", exp, res)
	}
}

func TestMany0CombinatorPass(t *testing.T) {
	ch := CharParser('d')
	m0 := Many0(ch)
	curState := state{
		"dddd",
		0,
	}

	_, err := m0(curState)

	// TODO: Find an alternative to reflect.DeepEqual()!!!
	exp := result{
		[]string{"d", "d", "d", "d"},
		state{
			"dddd",
			4,
		},
	}
	if err != nil{
		t.Fatalf("Many0 combinator test failed. Expected %v but got the following error:\n %s.", exp, err.Error())
	}
}

func TestMany1CombinatorPass(t *testing.T) {
	ch := CharParser('d')
	m1 := Many1(ch)
	curState := state{
		"dddd",
		0,
	}

	_, err := m1(curState)

	exp := result{
		[]string{"d", "d", "d", "d"},
		state{
			"dddd",
			4,
		},
	}
	if err != nil{
		t.Fatalf("Many0 combinator test failed. Expected %v but got the following error:\n %s.", exp, err.Error())
	}
}

func TestMany1CombinatorFail(t *testing.T) {
	ch := CharParser('c')
	m1 := Many1(ch)
	curState := state{
		"dddd",
		0,
	}

	_, err := m1(curState)

	exp := result{
		nil,
		curState,
	}

	if err == nil{
		t.Fatalf("Many0 combinator test failed. Expected error but got %v.", exp)
	}
}

func TestSeqCombinatorPass(t *testing.T) {
	seq := Seq(
		CharParser('h'),
		CharParser('i'),
		Many0(CharParser('!')),
	)

	curState := state{
		"hi!!?",
		0,
	}
	res, err := seq(curState)

	exp := result{
		[]string{"h", "i", "!", "!"},
		state{
			"hi!!?",
			3,
		},
	}
	if err != nil {
		t.Fatalf("Seq combinator test failed. expected %v but received %v and the following error:\n %s", exp, res, err.Error())
	}
}

func TestOptionalCombinatorPass(t *testing.T) {
	q := CharParser('?')	
	op := Optional(q)
	curState := state{
		"?!!!",
		0,
	}

	_, err := op(curState)

	exp := state{
		"?!!!",
		1,
	}

	if err != nil {
		t.Fatalf("Optional Combinator test failed. Expected %v but got the following error:\n %s.", exp, err.Error())	
	}
}

func TestBetweenCombinatorPass(t *testing.T) {
	opening := CharParser('{')
	closing := CharParser('}')
	ab := Many0(Or(CharParser('a'), CharParser('b')))	
	op := Between(opening, ab, closing)
	curState := state{
		"{abba}",
		0,
	}
	_, err := op(curState)

	exp := result{
		[]string{"a", "b", "b", "a"},
		state{
			"{abba}",
			6,
		},
	}

	if err != nil {
		t.Fatalf("Between Combinator test failed. Expected %v but got the following error:\n %s.", exp, err.Error())	
	}
}

func TestBetweenCombinatorFail(t *testing.T) {
	opening := CharParser('{')
	closing := CharParser('}')
	ab := Many0(Or(CharParser('a'), CharParser('b')))	
	op := Between(opening, ab, closing)
	curState := state{
		"{abba",
		0,
	}
	_, err := op(curState)

	exp := result{
		nil,
		curState,
	}

	if err == nil {
		t.Fatalf("Between Combinator test failed. Expected error but got %v.", exp)	
	}

}

func TestLazyCombinatorPass(t *testing.T) {
	var expr Parser
	expr = Lazy(func() Parser {
		return Or(
			CharParser('a'),
			Seq(
				CharParser('('),
				expr,  
				CharParser(')'),
			),
		)
	})

	curState := state{
		"a(a(a))",
		0,
	}

	_, err := expr(curState)
	exp := result{
		[]string{"a", "(", "a", "(", "a", ")", ")"},
		state{
			"a(a(a))",
			7,
		},
	}	

	if err != nil {
		t.Fatalf("Lazy Combinator test failed. Expected %v but got the following error:\n %s.", exp, err.Error())	
	}
}

func TestLazyCombinatorFail(t *testing.T) {
	var expr Parser
	expr = Lazy(func() Parser {
		return Or(
			CharParser('a'),
			Seq(
				CharParser('('),
				expr,  // Recursive reference
				CharParser(')'),
			),
		)
	})

	curState := state{
		"(a(a)",
		0,
	}

	res, err := expr(curState)
	exp := result{
		nil,
		curState,
	}	

	if err == nil {
		t.Fatalf("Lazy Combinator test failed. Expected error but received %v.", res)	
	}

	if !reflect.DeepEqual(exp, res) {
		t.Fatalf("Lazy Combinator test failed. Expected %v but received %v.", exp, res)
	}
}