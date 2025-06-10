package parser

type Error struct {
	Message string
	Expected []string
	Got 	string
	Position Position
}
