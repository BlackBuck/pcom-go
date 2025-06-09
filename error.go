package parser

type Error struct {
	Message string
	Line    int
	Column  int
}
