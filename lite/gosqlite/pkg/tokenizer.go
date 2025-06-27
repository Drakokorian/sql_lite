package pkg

import (
	"fmt"
	"strings"
)

// TokenType represents the type of a token.
type TokenType int

const (
	ILLEGAL TokenType = iota
	EOF

	// Identifiers + literals
	IDENT  // field_name, table_name
	INT    // 12345
	STRING // "hello world"

	// Operators
	EQ     // =
	NEQ    // !=
	LT     // <
	GT     // >
	LTE    // <=
	GTE    // >=
	PLUS   // +
	MINUS  // -
	ASTERISK // *
	SLASH  // /

	// Delimiters
	COMMA     // ,
	SEMICOLON // ;
	LPAREN    // (
	RPAREN    // )

	// Keywords
	SELECT
	FROM
	WHERE
	INSERT
	INTO
	VALUES
	CREATE
	TABLE
	TEXT
	INTEGER
	PRIMARY
	KEY
	NULL
	NOT
	AND
	OR
	TRUE
	FALSE
	LIMIT
	OFFSET
	ORDER
	BY
	ASC
	DESC
)

// String returns the string representation of the TokenType.
func (t TokenType) String() string {
	switch t {
	case ILLEGAL: return "ILLEGAL"
	case EOF: return "EOF"
	case IDENT: return "IDENT"
	case INT: return "INT"
	case STRING: return "STRING"
	case EQ: return "EQ"
	case NEQ: return "NEQ"
	case LT: return "LT"
	case GT: return "GT"
	case LTE: return "LTE"
	case GTE: return "GTE"
	case PLUS: return "PLUS"
	case MINUS: return "MINUS"
	case ASTERISK: return "ASTERISK"
	case SLASH: return "SLASH"
	case COMMA: return "COMMA"
	case SEMICOLON: return "SEMICOLON"
	case LPAREN: return "LPAREN"
	case RPAREN: return "RPAREN"
	case SELECT: return "SELECT"
	case FROM: return "FROM"
	case WHERE: return "WHERE"
	case INSERT: return "INSERT"
	case INTO: return "INTO"
	case VALUES: return "VALUES"
	case CREATE: return "CREATE"
	case TABLE: return "TABLE"
	case TEXT: return "TEXT"
	case INTEGER: return "INTEGER"
	case PRIMARY: return "PRIMARY"
	case KEY: return "KEY"
	case NULL: return "NULL"
	case NOT: return "NOT"
	case AND: return "AND"
	case OR: return "OR"
	case TRUE: return "TRUE"
	case FALSE: return "FALSE"
	case LIMIT: return "LIMIT"
	case OFFSET: return "OFFSET"
	case ORDER: return "ORDER"
	case BY: return "BY"
	case ASC: return "ASC"
	case DESC: return "DESC"
	default: return "UNKNOWN"
	}
}

// Token represents a lexical token of the SQL language.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// keywords maps reserved words to their token types.
var keywords = map[string]TokenType{
	"select":  SELECT,
	"from":    FROM,
	"where":   WHERE,
	"insert":  INSERT,
	"into":    INTO,
	"values":  VALUES,
	"create":  CREATE,
	"table":   TABLE,
	"text":    TEXT,
	"integer": INTEGER,
	"primary": PRIMARY,
	"key":     KEY,
	"null":    NULL,
	"not":     NOT,
	"and":     AND,
	"or":      OR,
	"true":    TRUE,
	"false":   FALSE,
	"limit":   LIMIT,
	"offset":  OFFSET,
	"order":   ORDER,
	"by":      BY,
	"asc":     ASC,
	"desc":    DESC,
}

// LookupIdent checks if the given identifier is a keyword.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}

// Tokenizer breaks an SQL query string into a stream of tokens.
type Tokenizer struct {
	input        string
	position     int // current position in input (points to current char)
	readPosition int // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int
	column       int
	maxQueryLen  int // Maximum allowed query length
	errors       []string // Stores all errors encountered
}

// NewTokenizer creates a new Tokenizer instance.
func NewTokenizer(input string, maxQueryLen int) *Tokenizer {
	t := &Tokenizer{input: input, line: 1, column: 0, maxQueryLen: maxQueryLen, errors: []string{}}
	t.readChar()
	return t
}

// readChar reads the next character and advances the positions.
func (t *Tokenizer) readChar() {
	if t.readPosition >= len(t.input) {
		t.ch = 0 // ASCII for "NUL" (EOF)
	} else {
		t.ch = t.input[t.readPosition]
	}
	t.position = t.readPosition
	t.readPosition++
	t.column++

	// Enforce max query length
	if t.position >= t.maxQueryLen {
		t.errors = append(t.errors, fmt.Sprintf("query exceeds maximum allowed length of %d characters", t.maxQueryLen))
		t.ch = 0 // Stop further processing
	}
}

// peekChar returns the next character without advancing positions.
func (t *Tokenizer) peekChar() byte {
	if t.readPosition >= len(t.input) {
		return 0
	}
	return t.input[t.readPosition]
}

// skipWhitespace skips over whitespace characters.
func (t *Tokenizer) skipWhitespace() {
	for t.ch == ' ' || t.ch == '\t' || t.ch == '\n' || t.ch == '\r' {
		if t.ch == '\n' {
			t.line++
			t.column = 0
		}
		t.readChar()
	}
}

// NextToken returns the next token from the input string.
func (t *Tokenizer) NextToken() Token {
	var tok Token

	t.skipWhitespace()

	tok.Line = t.line
	tok.Column = t.column

	switch t.ch {
	case '=':
		tok = newToken(EQ, t.ch)
	case '!':
		if t.peekChar() == '=' {
			ch := t.ch
			t.readChar()
			literal := string(ch) + string(t.ch)
			tok = Token{Type: NEQ, Literal: literal, Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(ILLEGAL, t.ch)
			t.errors = append(t.errors, fmt.Sprintf("unexpected character: %q at line %d, column %d", t.ch, tok.Line, tok.Column))
		}
	case ';':
		tok = newToken(SEMICOLON, t.ch)
	case ',':
		tok = newToken(COMMA, t.ch)
	case '(':
		tok = newToken(LPAREN, t.ch)
	case ')':
		tok = newToken(RPAREN, t.ch)
	case '>':
		if t.peekChar() == '=' {
			ch := t.ch
			t.readChar()
			literal := string(ch) + string(t.ch)
			tok = Token{Type: GTE, Literal: literal, Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(GT, t.ch)
		}
	case '<':
		if t.peekChar() == '=' {
			ch := t.ch
			t.readChar()
			literal := string(ch) + string(t.ch)
			tok = Token{Type: LTE, Literal: literal, Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(LT, t.ch)
		}
	case '+':
		tok = newToken(PLUS, t.ch)
	case '-':
		tok = newToken(MINUS, t.ch)
	case '*':
		tok = newToken(ASTERISK, t.ch)
	case '/':
		tok = newToken(SLASH, t.ch)
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if isLetter(t.ch) {
			tok.Literal = t.readIdentifier()
			tok.Type = LookupIdent(tok.Literal)
			return tok
		} else if isDigit(t.ch) {
			tok.Type = INT
			tok.Literal = t.readNumber()
			return tok
		} else if t.ch == '\'' {
			tok.Type = STRING
			tok.Literal = t.readString()
			return tok
		} else {
			tok = newToken(ILLEGAL, t.ch)
			t.errors = append(t.errors, fmt.Sprintf("unexpected character: %q at line %d, column %d", t.ch, tok.Line, tok.Column))
		}
	}

	t.readChar()
	return tok
}

// Errors returns the tokenizer errors.
func (t *Tokenizer) Errors() []string {
	return t.errors
}

// readIdentifier reads an identifier (letters and underscores).
func (t *Tokenizer) readIdentifier() string {
	position := t.position
	for isLetter(t.ch) || isDigit(t.ch) {
		t.readChar()
	}
	return t.input[position:t.position]
}

// readNumber reads a number (digits).
func (t *Tokenizer) readNumber() string {
	position := t.position
	for isDigit(t.ch) {
		t.readChar()
	}
	return t.input[position:t.position]
}

// readString reads a string literal (enclosed in single quotes).
func (t *Tokenizer) readString() string {
	position := t.position + 1 // Skip the opening quote
	for {
		t.readChar()
		if t.ch == '\'' || t.ch == 0 {
			break
		}
	}
	literal := t.input[position:t.position]
	t.readChar() // Consume the closing quote
	return literal
}

// newToken creates a new Token.
func newToken(tokenType TokenType, ch byte) Token {
	return Token{Type: tokenType, Literal: string(ch)}
}

// isLetter checks if a byte is a letter or underscore.
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit checks if a byte is a digit.
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
