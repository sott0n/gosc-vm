package lexer

import (
	"gosc-vm/token"
	"testing"
)

func TestNextTokenTrivial(t *testing.T) {
	input := `,`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.COMMA, ","},
		{token.EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, but got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, but got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextTokenReal(t *testing.T) {
	input := `
	store #1, 10
	store #2, 20
	add #0, #1, #2
	print_int #0
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.STORE, "store"},
		{token.IDENT, "#1"},
		{token.COMMA, ","},
		{token.INT, "10"},

		{token.STORE, "store"},
		{token.IDENT, "#2"},
		{token.COMMA, ","},
		{token.INT, "20"},

		{token.ADD, "add"},
		{token.IDENT, "#0"},
		{token.COMMA, ","},
		{token.IDENT, "#1"},
		{token.COMMA, ","},
		{token.IDENT, "#2"},

		{token.PRINT_INT, "print_int"},
		{token.IDENT, "#0"},

		{token.EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, but got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, but got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestSimpleComment(t *testing.T) {
	input := `# This is a comment
	# This is still a comment
	print_int #3
	# This is a final
	print_int #21
	# comment on two-lines
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.PRINT_INT, "print_int"},
		{token.IDENT, "#3"},
		{token.PRINT_INT, "print_int"},
		{token.IDENT, "#21"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, but got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, but got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
