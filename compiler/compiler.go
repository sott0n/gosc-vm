package compiler

import (
	"fmt"
	"gosc-vm/lexer"
	"gosc-vm/opcode"
	"gosc-vm/token"
	"os"
	"strconv"
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

// getRegister converts a register string "#2" to an integer 2.
func (p *Compiler) getRegister(input string) byte {
	num := strings.TrimPrefix(input, "#")

	i, err := strconv.Atoi(num)
	if err != nil {
		panic(err)
	}
	return byte(i)
}

// Compile process the stream of tokens from the lexer and
// builds up the bytecode program.
func (p *Compiler) Compiler() {

	// Until we get the end of our stream we'll process each token
	// in turn, generating bytecode as we go.
	for p.curToken.Type != token.EOF {

		// Now handle the various tokens
		switch p.curToken.Type {
		case token.LABEL:
			// Remove the ":" prefix from the label
			label := strings.TrimPrefix(p.curToken.Literal, ":")
			// The label points to the current point in our bytecode
			p.labels[label] = len(p.bytecode)

		case token.EXIT:
			p.exitOp()

		case token.INC:
			p.incOp()

		case token.DEC:
			p.decOp()

		case token.RANDOM:
			p.randOp()

		case token.RET:
			p.retOp()

		case token.CALL:
			p.callOp()

		case token.IS_INTEGER:
			p.isIntOp()

		case token.IS_STRING:
			p.isStrOp()

		case token.STRING2INT:
			p.str2IntOp()

		case token.INT2STRING:
			p.int2StrOp()

		case token.SYSTEM:
			p.systemOp()

		case token.CMP:
			p.cmpOp()

		case token.CONCAT:
			p.concatOp()

		case token.DB:
			p.dataOp()

		case token.DATA:
			p.dataOp()

		case token.GOTO:
			p.jumpOp(opcode.JUMP_TO)

		case token.JMP:
			p.jumpOp(opcode.JUMP_TO)

		case token.JMPZ:
			p.jumpOp(opcode.JUMP_Z)

		case token.JMPNZ:
			p.jumpOp(opcode.JUMP_NZ)

		case token.MEMCPY:
			p.memcpyOp()

		case token.NOP:
			p.nopOp()

		case token.PEEK:
			p.peekOp()

		case token.POKE:
			p.pokeOp()

		case token.PUSH:
			p.pushOp()

		case token.POP:
			p.popOp()

		case token.STORE:
			p.storeOp()

		case token.PRINT_INT:
			p.printInt()

		case token.PRINT_STR:
			p.printString()

		case token.ADD:
			p.mathOperation(opcode.ADD_OP)

		case token.SUB:
			p.mathOperation(opcode.SUB_OP)

		case token.MUL:
			p.mathOperation(opcode.MUL_OP)

		case token.DIV:
			p.mathOperation(opcode.DIV_OP)

		default:
			fmt.Println("Unhandled token: ", p.curToken)
		}
		p.nextToken()
	}

	// Now fixup any label-names we've got to patch into place.
	for addr, name := range p.fixups {
		value := p.labels[name]
		if value == 0 {
			fmt.Printf("Use of undefined label '%s\n'", name)
			os.Exit(1)
		}

		p1 := value % 256
		p2 := (value - p1) / 256

		p.bytecode[addr] = byte(p1)
		p.bytecode[addr+1] = byte(p2)
	}
}

// nopOp does nothing
func (p *Compiler) nopOp() {
	p.bytecode = append(p.bytecode, byte(opcode.NOP_OP))
}

// peekOp reads the contents of a memory address, and stores in a register
func (p *Compiler) peekOp() {
	// looking for an identifier next.
	if !p.expectPeek(token.IDENT) {
		return
	}

	res := p.getRegister(p.curToken.Literal)

	// now we have a comma
	if !p.expectPeek(token.COMMA) {
		return
	}
	p.nextToken()

	// and a literal
	if p.curToken.Type != token.IDENT {
		return
	}
	addr := p.getRegister(p.curToken.Literal)

	p.bytecode = append(p.bytecode, byte(opcode.PEEK))
	p.bytecode = append(p.bytecode, byte(res))
	p.bytecode = append(p.bytecode, byte(addr))
}

// pokeOp writes to memory
func (p *Compiler) pokeOp() {
	// looking for an identifier next.
	if !p.expectPeek(token.IDENT) {
		return
	}

	val := p.getRegister(p.curToken.Literal)

	// we have a comma
	if !p.expectPeek(token.COMMA) {
		return
	}
	p.nextToken()

	// and a literal
	if p.curToken.Type != token.IDENT {
		return
	}
	addr := getRegister(p.curToken.Literal)

	p.bytecode = append(p.bytecode, byte(opcode.POKE))
	p.bytecode = append(p.bytecode, byte(val))
	p.bytecode = append(p.bytecode, byte(addr))
}

// pushOp stores a stack-push
func (p *Compiler) pushOp() {
	// looking for an identifier next.
	if !p.expectPeek(token.IDENT) {
		return
	}

	// Save the register we're storing to.
	reg := p.getRegister(p.curToken.Literal)

	p.bytecode = append(p.bytecode, byte(opcode.STACK_PUSH))
	p.bytecode = append(p.bytecode, byte(reg))
}

// pophOp stores a stack-push
func (p *Compiler) popOp() {
	// looking for an identifier next.
	if !p.expectPeek(token.IDENT) {
		return
	}

	// Save the register we're storing to.
	reg := p.getRegister(p.curToken.Literal)

	p.bytecode = append(p.bytecode, byte(opcode.STACK_POP))
	p.bytecode = append(p.bytecode, byte(reg))
}

// exitOp terminates our interpreter
func (p *Compiler) exitOp() {
	p.bytecode = append(p.bytecode, byte(opcode.EXIT))
}

// incOp increments the contents of the given register
func (p *Compiler) incOp() {
	// looking for an identifier next.
	if !p.expectPeek(token.IDENT) {
		return
	}

	// Save the register we're storing to.
	reg := p.getRegister(p.curToken.Literal)

	p.bytecode = append(p.bytecode, opcode.INC_OP)
	p.bytecode = append(p.bytecode, byte(reg))
}

// decOp decrements the contents of the given register
func (p *Compiler) decOp() {
	// looking for an identifier next.
	if !p.expectPeek(token.IDENT) {
		return
	}

	// Save the register we're storing to.
	reg := p.getRegister(p.curToken.Literal)

	p.bytecode = append(p.bytecode, opcode.DEC_OP)
	p.bytecode = append(p.bytecode, byte(reg))
}

// randOp returns a random value
func (p *Compiler) randOp() {
	// looking for an identifier next.
	if !p.expectPeek(token.IDENT) {
		return
	}

	// Save the register we're storing to.
	reg := p.getRegister(p.curToken.Literal)

	p.bytecode = append(p.bytecode, opcode.INT_RANDOM)
	p.bytecode = append(p.bytecode, byte(reg))
}

// retOp returns from a call
func (p *Compiler) retOp() {
	p.bytecode = append(p.bytecode, byte(opcode.STACK_RET))
}

// isStrOp tests if a register contains a string
func (p *Compiler) isStrOp() {
	// looking for an identifier next.
	if !p.expectPeek(token.IDENT) {
		return
	}

	// Save the register we're storing to.
	reg := p.getRegister(p.curToken.Literal)

	p.bytecode = append(p.bytecode, opcode.IS_STRING)
	p.bytecode = append(p.bytecode, byte(reg))
}

// str2IntOp converts the given string-register to an int.
func (p *Compiler) str2IntOp() {
	// looking for an identifier next.
	if !p.expectPeek(token.IDENT) {
		return
	}

	// Save the register we're storing to.
	reg := p.getRegister(p.curToken.Literal)

	p.bytecode = append(p.bytecode, opcode.STRING_TOINT)
	p.bytecode = append(p.bytecode, byte(reg))
}

// int2StrOp converts the given int-register to a string.
func (p *Compiler) int2StrOp() {
	// looking for an identifier next.
	if !p.expectPeek(token.IDENT) {
		return
	}

	// Save the register we're storing to.
	reg := p.getRegister(p.curToken.Literal)

	p.bytecode = append(p.bytecode, opcode.INT_TOSTRING)
	p.bytecode = append(p.bytecode, byte(reg))
}
