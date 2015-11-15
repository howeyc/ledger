package calc

type TokenType int

type Token struct {
	Type  TokenType
	Value string
}

var eof = rune(0)

const (
	NUMBER TokenType = iota
	LPAREN
	RPAREN
	CONSTANT
	FUNCTION
	OPERATOR
	WHITESPACE
	ERROR
	EOF
)
