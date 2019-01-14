package token

// TokenType is a string
type TokenType string

// Token struct represent the lexer token
type Token struct {
	Type    TokenType
	Literal string
}

// pre-defined TokenType
const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"
	IDENT   = "IDENT"
	LABEL   = "LABEL"
	INT     = "INT"
	STRING  = "STRING"
	COMMA   = ","

	// math
	ADD = "ADD"
	SUB = "SUB"
	MUL = "MUL"
	DIV = "DIV"
	INC = "INC"
	DEC = "DEC"

	// control-flow
	CALL  = "CALL"
	GOTO  = "GOTO"
	JMP   = "JMP"
	JMPNZ = "JMPNZ"
	JMPZ  = "JMPZ"
	RET   = "RET"

	// stack
	PUSH = "PUSH"
	POP  = "POP"

	// types
	IS_STRING  = "IS_STRING"
	IS_INTEGER = "IS_INTEGER"
	STRING2INT = "STRING2INT"
	INT2STRING = "INT2STRING"

	// compare
	CMP = "CMP"

	// store
	STORE = "STORE"

	// print
	PRINT_INT = "PRINT_INT"
	PRINT_STR = "PRINT_STR"

	// memory
	PEEK = "PEEK"
	POKE = "POKE"

	//Misc
	CONCAT = "CONCAT"
	DATA   = "DATA"
	DB     = "DB"
	EXIT   = "EXIT"
	MEMCPY = "MEMCPY"
	NOP    = "NOP"
	RANDOM = "RANDOM"
	SYSTEM = "SYSTEM"
)

// reversed keywords
var keywords = map[string]TokenType{

	// compare
	"cmp": CMP,

	// types
	"is_integer": IS_INTEGER,
	"is_string":  IS_STRING,
	"int2string": INT2STRING,
	"string2int": STRING2INT,

	// store
	"store": STORE,

	// print
	"print_int": PRINT_INT,
	"print_str": PRINT_STR,

	// math
	"add": ADD,
	"sub": SUB,
	"mul": MUL,
	"div": DIV,
	"inc": INC,
	"dec": DEC,

	// control-flow
	"call":  CALL,
	"goto":  GOTO,
	"jmp":   JMP,
	"jmpnz": JMPNZ,
	"jmpz":  JMPZ,
	"ret":   RET,

	// stack
	"push": PUSH,
	"pop":  POP,

	// memory
	"peek": PEEK,
	"poke": POKE,

	// misc
	"exit":   EXIT,
	"concat": CONCAT,
	"DATA":   DATA,
	"DB":     DB,
	"memcpy": MEMCPY,
	"nop":    NOP,
	"random": RANDOM,
	"system": SYSTEM,
}

// LookupIdentifier used to determine whether identifier is keyword nor not
func LookupIdentifier(identifier string) TokenType {
	if tok, ok := keywords[identifier]; ok {
		return tok
	}
	return IDENT
}
