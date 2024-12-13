package lexer

import (
	"unicode"
)

// TokenType represents the type of token
type TokenType string

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
}

// Token types
const (
	TokenIllegal      TokenType = "ILLEGAL"
	TokenEOF          TokenType = "EOF"
	TokenLeftBrace    TokenType = "{"
	TokenRightBrace   TokenType = "}"
	TokenLeftBracket  TokenType = "["
	TokenRightBracket TokenType = "]"
	TokenColon        TokenType = ":"
	TokenComma        TokenType = ","
	TokenString       TokenType = "STRING"
	TokenNumber       TokenType = "NUMBER"
	TokenTrue         TokenType = "TRUE"
	TokenFalse        TokenType = "FALSE"
	TokenNull         TokenType = "NULL"
)

// Lexer represents a JSON lexer.
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

// NewLexer initializes a new lexer with the given input.
func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// readChar reads the next character and advances the positions.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII code for NUL, signifies EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// NextToken retrieves the next token from the input.
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.ch {
	case '{':
		tok = Token{Type: TokenLeftBrace, Literal: "{"}
		l.readChar()
	case '}':
		tok = Token{Type: TokenRightBrace, Literal: "}"}
		l.readChar()
	case '[':
		tok = Token{Type: TokenLeftBracket, Literal: "["}
		l.readChar()
	case ']':
		tok = Token{Type: TokenRightBracket, Literal: "]"}
		l.readChar()
	case ':':
		tok = Token{Type: TokenColon, Literal: ":"}
		l.readChar()
	case ',':
		tok = Token{Type: TokenComma, Literal: ","}
		l.readChar()
	case '"':
		str, err := l.readString()
		if err != nil {
			tok = Token{Type: TokenIllegal, Literal: err.Error()}
		} else {
			tok = Token{Type: TokenString, Literal: str}
		}
	case 't':
		if l.peekWord("true") {
			tok = Token{Type: TokenTrue, Literal: "true"}
			l.advance(len("true"))
		} else {
			l.readIdentifier()
			tok = Token{Type: TokenIllegal, Literal: "t"}
		}
	case 'f':
		if l.peekWord("false") {
			tok = Token{Type: TokenFalse, Literal: "false"}
			l.advance(len("false"))
		} else {
			l.readIdentifier()
			tok = Token{Type: TokenIllegal, Literal: "f"}
		}
	case 'n':
		if l.peekWord("null") {
			tok = Token{Type: TokenNull, Literal: "null"}
			l.advance(len("null"))
		} else {
			l.readIdentifier()
			tok = Token{Type: TokenIllegal, Literal: "n"}
		}
	case 0:
		tok.Literal = ""
		tok.Type = TokenEOF
	default:
		if isDigit(l.ch) || l.ch == '-' {
			num := l.readNumber()
			tok = Token{Type: TokenNumber, Literal: num}
			return tok
		} else {
			tok = Token{Type: TokenIllegal, Literal: string(l.ch)}
			l.readChar()
		}
	}

	return tok
}

// skipWhitespace advances the lexer past any whitespace characters.
func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(rune(l.ch)) {
		l.readChar()
	}
}

// readString reads a string literal.
func (l *Lexer) readString() (string, error) {
	position := l.position + 1 // skip opening quote
	for {
		l.readChar()
		if l.ch == '"' {
			break
		}
		if l.ch == 0 {
			return "", ErrUnterminatedString
		}
	}
	str := l.input[position:l.position]
	l.readChar() // consume the closing quote
	return str, nil
}

// readNumber reads a number literal.
func (l *Lexer) readNumber() string {
	position := l.position
	if l.ch == '-' {
		l.readChar()
	}
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	return l.input[position:l.position]
}

// peekWord checks if the upcoming characters match the given word.
func (l *Lexer) peekWord(word string) bool {
	if l.position+len(word) > len(l.input) {
		return false
	}
	// Check if the word matches and is followed by a non-letter/digit
	matched := l.input[l.position:l.position+len(word)] == word
	if !matched {
		return false
	}
	// Check if there's more input after the word
	if l.position+len(word) < len(l.input) {
		nextChar := l.input[l.position+len(word)]
		// If next char is a letter or digit, this is not a complete word
		if isLetter(nextChar) || isDigit(nextChar) {
			return false
		}
	}
	return true
}

// advance moves the lexer forward by n characters.
func (l *Lexer) advance(n int) {
	for i := 0; i < n; i++ {
		l.readChar()
	}
}

// isDigit checks if a character is a digit.
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// Error definitions
var (
	ErrUnterminatedString = &LexerError{"unterminated string"}
)

// LexerError represents an error encountered by the lexer.
type LexerError struct {
	Message string
}

func (e *LexerError) Error() string {
	return e.Message
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}
