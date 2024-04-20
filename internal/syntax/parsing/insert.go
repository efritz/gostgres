package parsing

import (
	"github.com/efritz/gostgres/internal/nodes"
	"github.com/efritz/gostgres/internal/queries/mutation"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// insert := `INTO` table columnNames selectOrValues returning
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

	// TODO - support `DEFAULT VALUES`
	node, err := p.parseSelectOrValues()
	if err != nil {
		return nil, err
	}

	returningExpressions, err := p.parseReturning(name)
	if err != nil {
		return nil, err
	}

	return mutation.NewInsert(node, table, name, alias, columnNames, returningExpressions)
}
