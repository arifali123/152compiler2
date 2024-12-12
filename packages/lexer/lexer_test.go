package lexer

import (
	"testing"
)

func TestLexer_Tokenize_ValidJSON(t *testing.T) {
	input := `{
		"name": "John Doe",
		"age": 30.5,
		"active": true,
		"retired": false,
		"manager": null,
		"tags": ["golang", "json"]
	}`
	l := NewLexer(input)

	expectedTokens := []Token{
		{Type: TokenLeftBrace, Literal: "{"},
		{Type: TokenString, Literal: "name"},
		{Type: TokenColon, Literal: ":"},
		{Type: TokenString, Literal: "John Doe"},
		{Type: TokenComma, Literal: ","},
		{Type: TokenString, Literal: "age"},
		{Type: TokenColon, Literal: ":"},
		{Type: TokenNumber, Literal: "30.5"},
		{Type: TokenComma, Literal: ","},
		{Type: TokenString, Literal: "active"},
		{Type: TokenColon, Literal: ":"},
		{Type: TokenTrue, Literal: "true"},
		{Type: TokenComma, Literal: ","},
		{Type: TokenString, Literal: "retired"},
		{Type: TokenColon, Literal: ":"},
		{Type: TokenFalse, Literal: "false"},
		{Type: TokenComma, Literal: ","},
		{Type: TokenString, Literal: "manager"},
		{Type: TokenColon, Literal: ":"},
		{Type: TokenNull, Literal: "null"},
		{Type: TokenComma, Literal: ","},
		{Type: TokenString, Literal: "tags"},
		{Type: TokenColon, Literal: ":"},
		{Type: TokenLeftBracket, Literal: "["},
		{Type: TokenString, Literal: "golang"},
		{Type: TokenComma, Literal: ","},
		{Type: TokenString, Literal: "json"},
		{Type: TokenRightBracket, Literal: "]"},
		{Type: TokenRightBrace, Literal: "}"},
		{Type: TokenEOF, Literal: ""},
	}

	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}

	if len(tokens) != len(expectedTokens) {
		t.Fatalf("expected %d tokens, got %d", len(expectedTokens), len(tokens))
	}

	for i, tok := range tokens {
		expected := expectedTokens[i]
		if tok.Type != expected.Type || tok.Literal != expected.Literal {
			t.Errorf("token %d - expected (%v, %q), got (%v, %q)", i, expected.Type, expected.Literal, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_Tokenize_InvalidJSON(t *testing.T) {
	tests := []struct {
		input           string
		expectedIllegal string
		description     string
	}{
		{`{"name": "John Doe`, "unterminated string", "Unterminated string"},
		{`{"name": "John Doe", "active": tru}`, "t", "Invalid true literal"},
		{`{"name": "John Doe", "active": fal}`, "f", "Invalid false literal"},
		{`{"name": "John Doe", "active": nul}`, "n", "Invalid null literal"},
		{`{"name": "John Doe", "active": truexyz}`, "t", "Invalid true with extra chars"},
		{`{"name": "John Doe", "active": falsexyz}`, "f", "Invalid false with extra chars"},
		{`{"name": "John Doe", "active": nullxyz}`, "n", "Invalid null with extra chars"},
		{`{"name": @}`, "@", "Invalid character"},
		{`{"age": #123}`, "#", "Invalid character before number"},
		{`{"test": $}`, "$", "Invalid character as value"},
	}

	for i, tt := range tests {
		l := NewLexer(tt.input)
		var foundIllegal bool
		for {
			tok := l.NextToken()
			if tok.Type == TokenIllegal {
				if tok.Literal != tt.expectedIllegal {
					t.Errorf("test case %d (%s) - expected illegal token %q, got %q",
						i, tt.description, tt.expectedIllegal, tok.Literal)
				}
				foundIllegal = true
				break
			}
			if tok.Type == TokenEOF {
				break
			}
		}
		if !foundIllegal {
			t.Errorf("test case %d (%s) - expected to find an illegal token", i, tt.description)
		}
	}
}

func TestLexer_Numbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"42", "42"},
		{"-42", "-42"},
		{"3.14", "3.14"},
		{"-3.14", "-3.14"},
		{"0.123", "0.123"},
		{"-0.123", "-0.123"},
	}

	for i, tt := range tests {
		l := NewLexer(tt.input)
		tok := l.NextToken()
		if tok.Type != TokenNumber {
			t.Errorf("test case %d - expected token type NUMBER, got %v", i, tok.Type)
		}
		if tok.Literal != tt.expected {
			t.Errorf("test case %d - expected %q, got %q", i, tt.expected, tok.Literal)
		}
	}
}

func TestLexerError_Error(t *testing.T) {
	err := &LexerError{"test error message"}
	if err.Error() != "test error message" {
		t.Errorf("expected error message %q, got %q", "test error message", err.Error())
	}
}
