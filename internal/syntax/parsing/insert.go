package parsing

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// insertTail := `INTO` table [ `(` ident [, ...] `)` ] selectOrValues returning
func (p *parser) parseInsert(token tokens.Token) (queries.Node, error) {
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

	returningExpressions, err := p.parseReturning(tableDescription.name)
	if err != nil {
		return nil, err
	}

	builder := &InsertBuilder{
		tableDescription:     tableDescription,
		columnNames:          columnNames,
		node:                 node,
		returningExpressions: returningExpressions,
	}

	return builder.Build()
}
