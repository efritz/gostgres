package parsing

import (
	"github.com/efritz/gostgres/internal/nodes"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// insert := `INTO` table columnNames selectOrValues [`RETURNING` selectExpressions]
func (p *parser) parseInsert(token tokens.Token) (nodes.Node, error) {
	if _, err := p.mustAdvance(isType(tokens.TokenTypeInto)); err != nil {
		return nil, err
	}

	table, name, alias, err := p.parseTable()
	if err != nil {
		return nil, err
	}

	columnNames, err := p.parseColumnNames()
	if err != nil {
		return nil, err
	}

	node, err := p.parseSelectOrValues()
	if err != nil {
		return nil, err
	}

	returningExpressions, err := p.parseReturning(name)
	if err != nil {
		return nil, err
	}

	return nodes.NewInsert(node, table, name, alias, columnNames, returningExpressions)
}

// columnNames := [ `(` ident [, ...] `)` ]
func (p *parser) parseColumnNames() ([]string, error) {
	if p.current().Type != tokens.TokenTypeLeftParen || p.peek(1).Type != tokens.TokenTypeIdent {
		return nil, nil
	}

	p.advance()

	var columnNames []string
	for {
		nameToken, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
		if err != nil {
			return nil, err
		}

		columnNames = append(columnNames, nameToken.Text)

		if !p.advanceIf(isType(tokens.TokenTypeComma)) {
			break
		}
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeRightParen)); err != nil {
		return nil, err
	}

	return columnNames, nil
}
