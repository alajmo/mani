package dao

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TOKEN_TAG TokenType = iota
	TOKEN_AND
	TOKEN_OR
	TOKEN_NOT
	TOKEN_LPAREN
	TOKEN_RPAREN
	TOKEN_EOF
)

type Position struct {
	line   int
	column int
}

type Token struct {
	Type     TokenType
	Value    string
	Position Position
}

type Lexer struct {
	input  string
	pos    int
	line   int
	column int
	tokens []Token
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
		tokens: make([]Token, 0),
	}
}

func (l *Lexer) Tokenize() error {
	if strings.TrimSpace(l.input) == "" {
		return fmt.Errorf("empty expression")
	}

	for l.pos < len(l.input) {
		char := l.current()
		switch {
		case char == ' ' || char == '\t':
			l.advance()
		case char == '\n':
			l.line++
			l.column = 1
			l.advance()
		case char == '(':
			l.addToken(TOKEN_LPAREN, "(")
			l.advance()
		case char == ')':
			l.addToken(TOKEN_RPAREN, ")")
			l.advance()
		case char == '!':
			l.addToken(TOKEN_NOT, "!")
			l.advance()
		case l.matchOperator("&&"):
			l.addToken(TOKEN_AND, "&&")
			l.advance()
			l.advance()
		case l.matchOperator("||"):
			l.addToken(TOKEN_OR, "||")
			l.advance()
			l.advance()
		case isValidTagStart(char):
			l.readTag()
		default:
			return fmt.Errorf("unexpected character: %c at line %d, column %d", char, l.line, l.column)
		}
	}

	l.addToken(TOKEN_EOF, "")
	return nil
}

func (l *Lexer) addToken(tokenType TokenType, value string) {
	l.tokens = append(l.tokens, Token{
		Type:     tokenType,
		Value:    value,
		Position: Position{line: l.line, column: l.column},
	})
}

func (l *Lexer) advance() {
	l.pos++
	l.column++
}

func (l *Lexer) current() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return rune(l.input[l.pos])
}

func (l *Lexer) matchOperator(op string) bool {
	if l.pos+len(op) > len(l.input) {
		return false
	}
	return l.input[l.pos:l.pos+len(op)] == op
}

func (l *Lexer) readTag() {
	startPos := l.pos
	startColumn := l.column

	// First character must be a letter
	if !isValidTagStart(l.current()) {
		return
	}
	l.advance()

	// Subsequent characters can be letters, numbers, hyphens, or underscores
	for l.pos < len(l.input) && isValidTagPart(l.current()) {
		l.advance()
	}

	value := l.input[startPos:l.pos]
	l.tokens = append(l.tokens, Token{
		Type:     TOKEN_TAG,
		Value:    value,
		Position: Position{line: l.line, column: startColumn},
	})
}

func isValidTagStart(r rune) bool {
	return unicode.IsLetter(r)
}

func isValidTagPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' || r == '_'
}

type Parser struct {
	tokens  []Token
	pos     int
	project *Project
}

func NewParser(tokens []Token, project *Project) *Parser {
	return &Parser{
		tokens:  tokens,
		pos:     0,
		project: project,
	}
}

func (p *Parser) Parse() (bool, error) {
	if len(p.tokens) <= 1 { // Only EOF token
		return false, fmt.Errorf("empty expression")
	}

	result, err := p.parseExpression()
	if err != nil {
		return false, err
	}

	// Check if we consumed all tokens
	if p.current().Type != TOKEN_EOF {
		pos := p.current().Position
		return false, fmt.Errorf("unexpected token at line %d, column %d", pos.line, pos.column)
	}

	return result, nil
}

func (p *Parser) parseExpression() (bool, error) {
	left, err := p.parseTerm()
	if err != nil {
		return false, err
	}

	for p.current().Type == TOKEN_OR {
		op := p.current()
		p.pos++

		// Check for missing right operand
		if p.current().Type == TOKEN_EOF {
			return false, fmt.Errorf("missing right operand for OR operator at line %d, column %d",
				op.Position.line, op.Position.column)
		}

		right, err := p.parseTerm()
		if err != nil {
			return false, err
		}
		left = left || right
	}

	return left, nil
}

func (p *Parser) parseTerm() (bool, error) {
	left, err := p.parseFactor()
	if err != nil {
		return false, err
	}

	for p.current().Type == TOKEN_AND {
		op := p.current()
		p.pos++

		// Check for missing right operand
		if p.current().Type == TOKEN_EOF {
			return false, fmt.Errorf("missing right operand for AND operator at line %d, column %d",
				op.Position.line, op.Position.column)
		}

		right, err := p.parseFactor()
		if err != nil {
			return false, err
		}
		left = left && right
	}

	return left, nil
}

func (p *Parser) parseFactor() (bool, error) {
	token := p.current()

	switch token.Type {
	case TOKEN_NOT:
		p.pos++
		if p.current().Type == TOKEN_EOF {
			return false, fmt.Errorf("missing operand after NOT at line %d, column %d",
				token.Position.line, token.Position.column)
		}
		val, err := p.parseFactor()
		if err != nil {
			return false, err
		}
		return !val, nil

	case TOKEN_LPAREN:
		p.pos++
		// Check for empty parentheses
		if p.current().Type == TOKEN_RPAREN {
			return false, fmt.Errorf("empty parentheses at line %d, column %d",
				token.Position.line, token.Position.column)
		}
		val, err := p.parseExpression()
		if err != nil {
			return false, err
		}
		if p.current().Type != TOKEN_RPAREN {
			return false, fmt.Errorf("missing closing parenthesis for opening parenthesis at line %d, column %d",
				token.Position.line, token.Position.column)
		}
		p.pos++
		return val, nil

	case TOKEN_TAG:
		p.pos++
		return p.project.hasTag(token.Value), nil

	default:
		return false, fmt.Errorf("unexpected token at line %d, column %d: %v",
			token.Position.line, token.Position.column, token.Value)
	}
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Project) hasTag(tag string) bool {
	for _, t := range p.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// evaluateExpression checks if a boolean tag expression evaluates to true for a given project.
// The function supports boolean operations on project tags with full operator precedence.
//
// Operators (in precedence order):
//  1. ()  - Parentheses for grouping
//  2. !   - NOT operator (logical negation)
//  3. &&  - AND operator (logical conjunction)
//  4. ||  - OR operator (logical disjunction)
//
// Tag Expression Example:
//
// Expression: (main && (dev || prod)) && !test
//
// Requirements:
// 1. Must have "main" tag            - Mandatory
// 2. Must have "dev" OR "prod" tag   - At least one required
// 3. Must NOT have "test" tag        - Excluded if present
//
// Matches tags:
//
//	["main", "dev"]
//	["main", "prod"]
//	["main", "dev", "prod"]
//
// Does NOT match tags:
//
//	["main"]                   - missing dev/prod
//	["main", "dev", "test"]    - has test tag
//	["dev", "prod"]            - missing main
func evaluateExpression(project *Project, expression string) (bool, error) {
	lexer := NewLexer(expression)
	err := lexer.Tokenize()
	if err != nil {
		return false, fmt.Errorf("lexer error: %v", err)
	}

	parser := NewParser(lexer.tokens, project)
	return parser.Parse()
}

func validateExpression(expression string) error {
	lexer := NewLexer(expression)
	err := lexer.Tokenize()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	project := &Project{Tags: []string{}}
	parser := NewParser(lexer.tokens, project)
	_, err = parser.Parse()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}
