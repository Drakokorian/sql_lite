package pkg

import (
	"fmt"
	"strconv"
	"strings"
)

// NodeType represents the type of an AST node.
type NodeType int

const (
	ProgramNode NodeType = iota
	SelectStatementNode
	InsertStatementNode
	CreateStatementNode
	ExpressionNode
	IdentifierNode
	IntegerLiteralNode
	StringLiteralNode
	BinaryExpressionNode
	TableConstraintNode
	ColumnDefinitionNode
)

// Node is the interface that all AST nodes implement.
type Node interface {
	TokenLiteral() string
	String() string
	NodeType() NodeType
}

// Statement is the interface that all statement nodes implement.
type Statement interface {
	Node
	statementNode()
}

// Expression is the interface that all expression nodes implement.
type Expression interface {
	Node	
	expressionNode()
}

// Program is the root node of every AST.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string { return "" }
func (p *Program) String() string {
	var out strings.Builder
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}
func (p *Program) NodeType() NodeType { return ProgramNode }

// SelectStatement represents a SELECT statement.
type SelectStatement struct {
	Token      Token // The SELECT token
	Columns    []Expression
	From       *Identifier
	Where      Expression
	Limit      Expression
	Offset     Expression
	OrderBy    []*OrderByClause
}

func (s *SelectStatement) statementNode()       {}
func (s *SelectStatement) TokenLiteral() string { return s.Token.Literal }
func (s *SelectStatement) String() string {
	var out strings.Builder
	out.WriteString("SELECT ")
	columns := []string{}
	for _, c := range s.Columns {
		columns = append(columns, c.String())
	}
	out.WriteString(strings.Join(columns, ", "))
	if s.From != nil {
		out.WriteString(" FROM " + s.From.String())
	}
	if s.Where != nil {
		out.WriteString(" WHERE " + s.Where.String())
	}
	if len(s.OrderBy) > 0 {
		out.WriteString(" ORDER BY ")
		orderByClauses := []string{}
		for _, ob := range s.OrderBy {
			orderByClauses = append(orderByClauses, ob.String())
		}
		out.WriteString(strings.Join(orderByClauses, ", "))
	}
	if s.Limit != nil {
		out.WriteString(" LIMIT " + s.Limit.String())
	}
	if s.Offset != nil {
		out.WriteString(" OFFSET " + s.Offset.String())
	}
	out.WriteString(";")
	return out.String()
}
func (s *SelectStatement) NodeType() NodeType { return SelectStatementNode }

// InsertStatement represents an INSERT statement.
type InsertStatement struct {
	Token   Token // The INSERT token
	Table   *Identifier
	Columns []*Identifier
	Values  []Expression
}

func (is *InsertStatement) statementNode()       {}
func (is *InsertStatement) TokenLiteral() string { return is.Token.Literal }
func (is *InsertStatement) String() string {
	var out strings.Builder
	out.WriteString("INSERT INTO " + is.Table.String())
	if len(is.Columns) > 0 {
		columns := []string{}
		for _, c := range is.Columns {
			columns = append(columns, c.String())
		}
		out.WriteString(" (" + strings.Join(columns, ", ") + ")")
	}
	out.WriteString(" VALUES (")
	values := []string{}
	for _, v := range is.Values {
		values = append(values, v.String())
	}
	out.WriteString(strings.Join(values, ", ") + ")")
	out.WriteString(";")
	return out.String()
}
func (is *InsertStatement) NodeType() NodeType { return InsertStatementNode }

// CreateStatement represents a CREATE TABLE statement.
type CreateStatement struct {
	Token   Token // The CREATE token
	Table   *Identifier
	Columns []*ColumnDefinition
	Constraints []*TableConstraint
}

func (cs *CreateStatement) statementNode()       {}
func (cs *CreateStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *CreateStatement) String() string {
	var out strings.Builder
	out.WriteString("CREATE TABLE " + cs.Table.String() + " (")
	parts := []string{}
	for _, col := range cs.Columns {
		parts = append(parts, col.String())
	}
	for _, cons := range cs.Constraints {
		parts = append(parts, cons.String())
	}
	out.WriteString(strings.Join(parts, ", ") + ")")
	out.WriteString(";")
	return out.String()
}
func (cs *CreateStatement) NodeType() NodeType { return CreateStatementNode }

// Identifier represents an identifier (e.g., column name, table name).
type Identifier struct {
	Token Token // The IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }
func (i *Identifier) NodeType() NodeType { return IdentifierNode }

// IntegerLiteral represents an integer literal.
type IntegerLiteral struct {
	Token Token // The INT token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }
func (il *IntegerLiteral) NodeType() NodeType { return IntegerLiteralNode }

// StringLiteral represents a string literal.
type StringLiteral struct {
	Token Token // The STRING token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "'" + sl.Token.Literal + "'" }
func (sl *StringLiteral) NodeType() NodeType { return StringLiteralNode }

// BinaryExpression represents a binary operation (e.g., 1 + 2, a > b).
type BinaryExpression struct {
	Token    Token // The operator token (e.g., EQ, PLUS)
	Left     Expression
	Operator string
	Right    Expression
}

func (be *BinaryExpression) expressionNode()      {}
func (be *BinaryExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BinaryExpression) String() string {
	var out strings.Builder
	out.WriteString("(")
	out.WriteString(be.Left.String())
	out.WriteString(" " + be.Operator + " ")
	out.WriteString(be.Right.String())
	out.WriteString(")")
	return out.String()
}
func (be *BinaryExpression) NodeType() NodeType { return BinaryExpressionNode }

// OrderByClause represents an ORDER BY clause.
type OrderByClause struct {
	Column    *Identifier
	Direction Token // ASC or DESC
}

func (ob *OrderByClause) String() string {
	return fmt.Sprintf("%s %s", ob.Column.String(), ob.Direction.Literal)
}

// ColumnDefinition represents a column definition in a CREATE TABLE statement.
type ColumnDefinition struct {
	Name    *Identifier
	DataType Token // TEXT, INTEGER
	Constraints []*ColumnConstraint
}

func (cd *ColumnDefinition) String() string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("%s %s", cd.Name.String(), cd.DataType.Literal))
	for _, cons := range cd.Constraints {
		out.WriteString(" " + cons.String())
	}
	return out.String()
}

// ColumnConstraint represents a column constraint (e.g., PRIMARY KEY, NOT NULL).
type ColumnConstraint struct {
	Type Token // PRIMARY, NOT, NULL
}

func (cc *ColumnConstraint) String() string {
	// Simplified for now. Will need more logic for composite constraints.
	switch cc.Type.Type {
	case PRIMARY:
		return "PRIMARY KEY"
	case NOT:
		return "NOT NULL"
	case NULL:
		return "NULL"
	default:
		return ""
	}
}

// TableConstraint represents a table constraint (e.g., PRIMARY KEY (col1, col2)).
type TableConstraint struct {
	Type Token // PRIMARY
	Columns []*Identifier
}

func (tc *TableConstraint) String() string {
	// Simplified for now. Will need more logic for composite constraints.
	switch tc.Type.Type {
	case PRIMARY:
		cols := []string{}
		for _, c := range tc.Columns {
			cols = append(cols, c.String())
		}
		return fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(cols, ", "))
	default:
		return ""
	}
}

// Parser parses a stream of tokens into an AST.
type Parser struct {
	l *Tokenizer
	currentToken Token
	peekToken    Token
	errors       []string
}

// NewParser creates a new Parser instance.
func NewParser(l *Tokenizer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	// Read two tokens, so currentToken and peekToken are both set.
	p.nextToken()
	p.nextToken()
	return p
}

// Errors returns the parser errors.
func (p *Parser) Errors() []string {
	return p.errors
}

// peekError adds an error if the peek token is not of the expected type.
func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
			t.String(), p.peekToken.Type.String())
	p.errors = append(p.errors, msg)
}

// nextToken advances the current and peek tokens.
func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// ParseProgram parses the entire SQL program.
func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	program.Statements = []Statement{}

	// Collect tokenizer errors first
	p.errors = append(p.errors, p.l.Errors()...)

	for p.currentToken.Type != EOF {
		smt := p.parseStatement()
		if smt != nil {
			program.Statements = append(program.Statements, smt)
		}
		// Advance token even if statement parsing failed to avoid infinite loop
		p.nextToken()
	}
	return program
}

// parseStatement parses a single SQL statement.
func (p *Parser) parseStatement() Statement {
	switch p.currentToken.Type {
	case SELECT:
		return p.parseSelectStatement()
	case INSERT:
		return p.parseInsertStatement()
	case CREATE:
		return p.parseCreateStatement()
	default:
		p.errors = append(p.errors, fmt.Sprintf("unknown statement type: %s", p.currentToken.Literal))
		return nil
	}
}

// parseSelectStatement parses a SELECT statement.
func (p *Parser) parseSelectStatement() *SelectStatement {
	smt := &SelectStatement{Token: p.currentToken}

	// Parse columns
	p.nextToken() // Consume SELECT
	smt.Columns = p.parseExpressionList(FROM) // Read until FROM

	// Check for FROM clause
	if p.peekTokenIs(FROM) {
		p.nextToken() // Consume FROM
		p.nextToken() // Consume table name
		smt.From = &Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
	}

	// Check for WHERE clause
	if p.peekTokenIs(WHERE) {
		p.nextToken() // Consume WHERE
		p.nextToken() // Consume expression start
		smt.Where = p.parseExpression(0) // Parse expression with lowest precedence
	}

	// Check for ORDER BY clause
	if p.peekTokenIs(ORDER) {
		p.nextToken() // Consume ORDER
		if !p.expectPeek(BY) {
			return nil
		}
		p.nextToken() // Consume BY
		smt.OrderBy = p.parseOrderByClauses()
	}

	// Check for LIMIT clause
	if p.peekTokenIs(LIMIT) {
		p.nextToken() // Consume LIMIT
		p.nextToken() // Consume limit value
		smt.Limit = &IntegerLiteral{Token: p.currentToken, Value: p.parseIntegerLiteral().Value}
	}

	// Check for OFFSET clause
	if p.peekTokenIs(OFFSET) {
		p.nextToken() // Consume OFFSET
		p.nextToken() // Consume offset value
		smt.Offset = &IntegerLiteral{Token: p.currentToken, Value: p.parseIntegerLiteral().Value}
	}

	// Consume semicolon if present
	if p.peekTokenIs(SEMICOLON) {
		p.nextToken()
	}

	return smt
}

// parseInsertStatement parses an INSERT statement.
func (p *Parser) parseInsertStatement() *InsertStatement {
	smt := &InsertStatement{Token: p.currentToken}

	if !p.expectPeek(INTO) {
		return nil
	}

	if !p.expectPeek(IDENT) {
		return nil
	}
	smt.Table = &Identifier{Token: p.currentToken, Value: p.currentToken.Literal}

	// Parse optional columns
	if p.peekTokenIs(LPAREN) {
		p.nextToken() // Consume LPAREN
		smt.Columns = p.parseIdentifierList(RPAREN)
		if !p.expectPeek(RPAREN) {
			return nil
		}
	}

	if !p.expectPeek(VALUES) {
		return nil
	}

	if !p.expectPeek(LPAREN) {
		return nil
	}

	smt.Values = p.parseExpressionList(RPAREN)

	if !p.expectPeek(RPAREN) {
		return nil
	}

	// Consume semicolon if present
	if p.peekTokenIs(SEMICOLON) {
		p.nextToken()
	}

	return smt
}

// parseCreateStatement parses a CREATE TABLE statement.
func (p *Parser) parseCreateStatement() *CreateStatement {
	smt := &CreateStatement{Token: p.currentToken}

	if !p.expectPeek(TABLE) {
		return nil
	}

	if !p.expectPeek(IDENT) {
		return nil
	}
	smt.Table = &Identifier{Token: p.currentToken, Value: p.currentToken.Literal}

	if !p.expectPeek(LPAREN) {
		return nil
	}

	// Parse column definitions and table constraints
	for !p.peekTokenIs(RPAREN) && !p.peekTokenIs(EOF) {
		p.nextToken()
		if p.currentTokenIs(IDENT) {
			// Assume it's a column definition
			colDef := p.parseColumnDefinition()
			if colDef != nil {
				smt.Columns = append(smt.Columns, colDef)
			}
		} else if p.currentTokenIs(PRIMARY) {
			// Assume it's a table constraint
			tableCons := p.parseTableConstraint()
			if tableCons != nil {
				smt.Constraints = append(smt.Constraints, tableCons)
			}
		} else {
			p.errors = append(p.errors, fmt.Sprintf("unexpected token in CREATE TABLE: %s", p.currentToken.Literal))
			return nil
		}

		if p.peekTokenIs(COMMA) {
			p.nextToken() // Consume COMMA
		}
	}

	if !p.expectPeek(RPAREN) {
		return nil
	}

	// Consume semicolon if present
	if p.peekTokenIs(SEMICOLON) {
		p.nextToken()
	}

	return smt
}

// parseColumnDefinition parses a column definition in CREATE TABLE.
func (p *Parser) parseColumnDefinition() *ColumnDefinition {
	colDef := &ColumnDefinition{Name: &Identifier{Token: p.currentToken, Value: p.currentToken.Literal}}

	p.nextToken() // Consume column name
	if !p.currentTokenIs(TEXT) && !p.currentTokenIs(INTEGER) {
		p.errors = append(p.errors, fmt.Sprintf("expected data type, got %s", p.currentToken.Literal))
		return nil
	}
	colDef.DataType = p.currentToken

	// Parse constraints
	for p.peekTokenIs(PRIMARY) || p.peekTokenIs(NOT) || p.peekTokenIs(NULL) {
		p.nextToken()
		constraint := &ColumnConstraint{Type: p.currentToken}
		if p.currentTokenIs(PRIMARY) {
			if !p.expectPeek(KEY) {
				return nil
			}
			constraint.Type = p.currentToken // KEY token
		} else if p.currentTokenIs(NOT) {
			if !p.expectPeek(NULL) {
				return nil
			}
			constraint.Type = p.currentToken // NULL token
		}
		colDef.Constraints = append(colDef.Constraints, constraint)
	}

	return colDef
}

// parseTableConstraint parses a table constraint in CREATE TABLE.
func (p *Parser) parseTableConstraint() *TableConstraint {
	tableCons := &TableConstraint{Type: p.currentToken}

	if p.currentTokenIs(PRIMARY) {
		if !p.expectPeek(KEY) {
			return nil
		}
		if !p.expectPeek(LPAREN) {
			return nil
		}
		tableCons.Columns = p.parseIdentifierList(RPAREN)
		if !p.expectPeek(RPAREN) {
			return nil
		}
	}
	return tableCons
}

// parseExpressionList parses a comma-separated list of expressions until a stop token.
func (p *Parser) parseExpressionList(stop TokenType) []Expression {
	list := []Expression{}

	if p.peekTokenIs(stop) {
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(0))

	for p.peekTokenIs(COMMA) {
		p.nextToken() // Consume COMMA
		p.nextToken() // Consume next expression start
		list = append(list, p.parseExpression(0))
	}

	return list
}

// parseIdentifierList parses a comma-separated list of identifiers until a stop token.
func (p *Parser) parseIdentifierList(stop TokenType) []*Identifier {
	list := []*Identifier{}

	if p.peekTokenIs(stop) {
		return list
	}

	p.nextToken()
	list = append(list, &Identifier{Token: p.currentToken, Value: p.currentToken.Literal})

	for p.peekTokenIs(COMMA) {
		p.nextToken() // Consume COMMA
		p.nextToken() // Consume next identifier
		list = append(list, &Identifier{Token: p.currentToken, Value: p.currentToken.Literal})
	}

	return list
}

// parseExpression parses an expression with operator precedence.
func (p *Parser) parseExpression(precedence int) Expression {
	leftExp := p.parsePrefixExpression()

	for !p.peekTokenIs(SEMICOLON) && precedence < p.peekPrecedence() {
		p.nextToken()
		leftExp = p.parseInfixExpression(leftExp)
	}

	return leftExp
}

// parsePrefixExpression parses a prefix expression (e.g., NOT).
func (p *Parser) parsePrefixExpression() Expression {
	switch p.currentToken.Type {
	case IDENT:
		return &Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
	case INT:
		val, err := strconv.ParseInt(p.currentToken.Literal, 10, 64)
		if err != nil {
			p.errors = append(p.errors, fmt.Sprintf("could not parse %q as integer", p.currentToken.Literal))
			return nil
		}
		return &IntegerLiteral{Token: p.currentToken, Value: val}
	case STRING:
		return &StringLiteral{Token: p.currentToken, Value: p.currentToken.Literal}
	case LPAREN:
		p.nextToken() // Consume LPAREN
		exp := p.parseExpression(0)
		if !p.expectPeek(RPAREN) {
			return nil
		}
		return exp
	default:
		p.errors = append(p.errors, fmt.Sprintf("no prefix parse function for %s found", p.currentToken.Literal))
		return nil
	}
}

// parseInfixExpression parses an infix expression (e.g., +, -, =).
func (p *Parser) parseInfixExpression(left Expression) Expression {
	exp := &BinaryExpression{
		Token:    p.currentToken,
		Operator: p.currentToken.Literal,
		Left:     left,
	}

	precedence := p.currentPrecedence()
	p.nextToken()
	exp.Right = p.parseExpression(precedence)

	return exp
}

// Precedence levels for operators.
const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

// precedences maps token types to their precedence levels.
var precedences = map[TokenType]int{
	EQ:     EQUALS,
	NEQ:    EQUALS,
	LT:     LESSGREATER,
	GT:     LESSGREATER,
	LTE:    LESSGREATER,
	GTE:    LESSGREATER,
	PLUS:   SUM,
	MINUS:  SUM,
	ASTERISK: PRODUCT,
	SLASH:  PRODUCT,
}

// peekPrecedence returns the precedence of the peek token.
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// currentPrecedence returns the precedence of the current token.
func (p *Parser) currentPrecedence() int {
	if p, ok := precedences[p.currentToken.Type]; ok {
		return p
	}
	return LOWEST
}

// expectPeek checks if the peek token is of the expected type and advances if it is.
func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	
	p.peekError(t)
	return false
}

// peekTokenIs checks if the peek token is of the given type.
func (p *Parser) peekTokenIs(t TokenType) bool {
	return p.peekToken.Type == t
}

// currentTokenIs checks if the current token is of the given type.
func (p *Parser) currentTokenIs(t TokenType) bool {
	return p.currentToken.Type == t
}

// parseOrderByClauses parses a list of ORDER BY clauses.
func (p *Parser) parseOrderByClauses() []*OrderByClause {
	clauses := []*OrderByClause{}

	for p.currentTokenIs(IDENT) {
		clause := &OrderByClause{Column: &Identifier{Token: p.currentToken, Value: p.currentToken.Literal}}
		p.nextToken() // Consume column name

		if p.currentTokenIs(ASC) || p.currentTokenIs(DESC) {
			clause.Direction = p.currentToken
			p.nextToken() // Consume ASC/DESC
		} else {
			// Default to ASC if no direction specified
			clause.Direction = Token{Type: ASC, Literal: "ASC"}
		}
		clauses = append(clauses, clause)

		if p.currentTokenIs(COMMA) {
			p.nextToken() // Consume COMMA
		} else {
			break // End of ORDER BY clauses
		}
	}
	return clauses
}

// parseIntegerLiteral parses an integer literal.
func (p *Parser) parseIntegerLiteral() *IntegerLiteral {
	val, err := strconv.ParseInt(p.currentToken.Literal, 10, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("could not parse %q as integer", p.currentToken.Literal))
		return nil
	}
	return &IntegerLiteral{Token: p.currentToken, Value: val}
}