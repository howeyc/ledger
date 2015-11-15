package calc

import (
	"fmt"
	"io"
)

type Parser struct {
	s   *Scanner
	buf struct {
		tok Token
		n   int
	}
}

func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

func (p *Parser) Scan() (tok Token) {
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok
	}

	tok = p.s.Scan()

	p.buf.tok = tok

	return
}

func (p *Parser) ScanIgnoreWhitespace() (tok Token) {
	tok = p.Scan()
	if tok.Type == WHITESPACE {
		tok = p.Scan()
	}
	return
}

func (p *Parser) Unscan() {
	p.buf.n = 1
}

func (p *Parser) Parse() (Stack, error) {
	stack := Stack{}
	for {
		tok := p.ScanIgnoreWhitespace()
		if tok.Type == ERROR {
			return Stack{}, fmt.Errorf("ERROR: %q", tok.Value)
		} else if tok.Type == EOF {
			break
		} else if tok.Type == OPERATOR && tok.Value == "-" {
			last_tok := stack.Peek()
			next_tok := p.ScanIgnoreWhitespace()
			if (last_tok.Type == OPERATOR || last_tok.Value == "" || last_tok.Type == LPAREN) && next_tok.Type == NUMBER {
				stack.Push(Token{NUMBER, "-" + next_tok.Value})
			} else {
				stack.Push(tok)
				p.Unscan()
			}
		} else {
			stack.Push(tok)
		}
	}
	return stack, nil
}
