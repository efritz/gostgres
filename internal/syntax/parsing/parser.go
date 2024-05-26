package parsing

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

type parser struct {
	tokens                  []tokens.Token
	cursor                  int
	tables                  TableGetter
	ddlParsers              ddlParsers
	createParsers           createParsers
	alterParsers            alterParsers
	addConstraintParsers    addConstraintParsers
	columnConstraintParsers columnConstraintParsers
	explainableParsers      explainableParsers
	prefixParsers           prefixParsers
	infixParsers            infixParsers
}

type TableGetter interface {
	Get(name string) (impls.Table, bool)
}

type tokenFilterFunc func(token tokens.Token) bool
type prefixParserFunc func(token tokens.Token) (impls.Expression, error)
type infixParserFunc func(left impls.Expression, token tokens.Token) (impls.Expression, error)

type ddlParsers map[tokens.TokenType]func(token tokens.Token) (queries.Query, error)
type createParsers map[tokens.TokenType]func() (queries.Query, error)
type alterParsers map[tokens.TokenType]func() (queries.Query, error)
type addConstraintParsers map[tokens.TokenType]func(name, tableName string) (queries.Query, error)
type columnConstraintParsers map[tokens.TokenType]func(columnName, tableName string, description *columnDescription) error
type explainableParsers map[tokens.TokenType]func(token tokens.Token) (queries.Node, error)
type prefixParsers map[tokens.TokenType]prefixParserFunc
type infixParsers map[tokens.TokenType]infixParserFunc

func newParser(tokenStream []tokens.Token, tables TableGetter) *parser {
	p := &parser{
		tokens: tokenStream,
		tables: tables,
	}

	p.initAlterParsers()
	p.initColumnConstraintParsers()
	p.initConstraintParsers()
	p.initCreateParsers()
	p.initDDLParsers()
	p.initExpressionInfixParsers()
	p.initExpressionPrefixParsers()
	p.initStatementParsers()
	return p
}

func (p *parser) current() tokens.Token {
	return p.peek(0)
}

func (p *parser) peek(n int) tokens.Token {
	if p.cursor+n >= len(p.tokens) {
		return tokens.InvalidToken
	}

	return p.tokens[p.cursor+n]
}

func (p *parser) advance() tokens.Token {
	r := p.current()
	p.cursor++
	return r
}

func (p *parser) advanceIf(filters ...tokenFilterFunc) bool {
	start := p.cursor
	for _, filter := range filters {
		if !filter(p.current()) {
			p.cursor = start
			return false
		}

		p.cursor++
	}

	return true
}

func (p *parser) mustAdvance(filter tokenFilterFunc) (tokens.Token, error) {
	current := p.advance()
	if !filter(current) {
		return tokens.InvalidToken, fmt.Errorf("unexpected token (near %s)", current.Text)
	}

	return current, nil
}

func isType(tokenType tokens.TokenType) tokenFilterFunc {
	return func(t tokens.Token) bool {
		return t.Type == tokenType
	}
}

func isIdent(text string) tokenFilterFunc {
	return func(t tokens.Token) bool {
		return t.Type == tokens.TokenTypeIdent && t.Text == text
	}
}
