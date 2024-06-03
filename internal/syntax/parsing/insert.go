package parsing

import (
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// insertTail := `INTO` table [ `(` ident [, ...] `)` ] selectOrValues returning
func (p *parser) parseInsert(token tokens.Token) (Builder, error) {
	if _, err := p.mustAdvance(isType(tokens.TokenTypeInto)); err != nil {
		return nil, err
	}

	tableDescription, err := p.parseTable()
	if err != nil {
		return nil, err
	}

	columnNames, err := parseParenthesizedCommaSeparatedList(p, true, false, p.parseIdent)
	if err != nil {
		return nil, err
	}

	// TODO - support `DEFAULT` expression and `DEFAULT VALUES`
	node, err := p.parseSelectOrValues()
	if err != nil {
		return nil, err
	}

	returningExpressions, err := p.parseReturning(tableDescription.Name)
	if err != nil {
		return nil, err
	}

	return &InsertBuilder{
		TableDescription:     tableDescription,
		ColumnNames:          columnNames,
		Node:                 node,
		ReturningExpressions: returningExpressions,
	}, nil
}
