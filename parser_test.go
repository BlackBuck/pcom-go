package parser

import (
	"fmt"
	"testing"
)

func TestCharParser(t *testing.T) {
	c := byte('c')
	charParser := charParser(c)
	_, err := charParser("chillin")

	expectedResult := result{
		"c",
		"hillin",
	}

	if err != nil {
		t.Fatalf("char parsing test failed. Expected %v got err", expectedResult)
	}

}

func TestOrCombinator(t *testing.T) {
	t1 := byte('c')
	t2 := byte('1')

	charOrDigit := or(charParser(t1), charParser(t2))
	_, err1 := charOrDigit("c23")
	_, err2 := charOrDigit("1d3")

	exp1 := result{
		"c",
		"23",
	}

	exp2 := result{
		"1",
		"d3",
	}

	if err1 != nil || err2 != nil{
		t.Fatalf("or parser test failed. Expected %v and %v received error.", exp1, exp2)
	}
}

func TestAndCombinator(t *testing.T) {
	t1 := byte('c')
	t2 := byte('1')

	charAndDigit := and(charParser(t1), charParser(t2))
	res, err1 := charAndDigit("c13")

	exp := result{
		[]string{"c", "1"},
		"3",
	}

	if err1 != nil{
		t.Fatalf("or parser test failed. Expected %v received %v and error.", exp, res)
	}
}

func TestMapCombinator(t *testing.T) {
	ch := charParser('1')
	mapfunc := func(ch string) string{
		return fmt.Sprintf("Mapped (%s)", ch)
	}

	mapcomb := Map(ch, mapfunc)
	res, err := mapcomb("123")

	exp := result{
		"Mapped (1)",
		"23",
	}
	if err != nil {
		t.Fatalf("map test failed. Expected %v but received %v and error", exp, res)
	}
}

func TestMany0Combinator(t *testing.T) {
	ch := charParser('c')
	m0 := Many0(ch)
	res, err := m0("dddd")

	exp := result{
		[]string{},
		"dddd",
	}
	if err != nil{
		t.Fatalf("many0 combinator test failed. expected %v received %v and error.", exp, res)
	}

}

func TestMany1Combinator(t *testing.T) {
	ch := charParser('c')
	m0 := Many1(ch)
	res, err := m0("cdddd")

	exp := result{
		[]string{"c", "c", "c", "c"},
		"dddd",
	}
	if err != nil{
		t.Fatalf("many0 combinator test failed. expected %v received %v and error.", exp, res)
	}
}

func TestSeqCombinator(t *testing.T) {
	seq := Seq(
		charParser('h'),
		charParser('i'),
		Many0(charParser('!')),
	)

	res, err := seq("hi!!?")

	exp := result{
		[]string{"h", "i", "!", "!"},
		"?",
	}
	if err != nil {
		t.Fatalf("seq combinator test failed. expected %v but received %v and error.", exp, res)
	}
}