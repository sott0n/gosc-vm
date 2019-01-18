package compiler

import (
	"gosc-vm/lexer"
	"gosc-vm/token"
	"strings"
)

type Compiler struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	bytecode  []byte

	labels map[string]int
	fixups map[int]string
}

// New is our costructor
func New(l *lexer.Lexer) *Compiler {
	p := &Compiler{l: l}
	p.labels = make(map[string]int)
	p.fixups = make(map[int]string)

	p.nextToken()
	p.nextToken()
	return p
}

// nextToken gets the next token from our lexer-stream
func (p *Compiler) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// isRegister returns true if the given string has a register ID
func (p *Compiler) isRegister(input string) bool {
	if strings.HasPrefix(input, "#") {
		return true
	}
	return false
}

