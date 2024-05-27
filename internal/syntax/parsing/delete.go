package parsing

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// deleteTail := `FROM` table deleteUsing where returning
func (p *parser) parseDelete(token tokens.Token) (queries.Node, error) {
	if _, err := p.mustAdvance(isType(tokens.TokenTypeFrom)); err != nil {
		return nil, err
	}

	table, name, aliasName, err := p.parseTable()
	if err != nil {
		return nil, err
	}

	usingExpressions, err := p.parseDeleteUsing()
	if err != nil {
		return nil, err
	}

	whereExpression, _, err := p.parseWhere()
	if err != nil {
		return nil, err
	}

	returningExpressions, err := p.parseReturning(name)
	if err != nil {
		return nil, err
	}

	builder := &DeleteBuilder{
		tableDescription: TableDescription{
			table:     table,
			name:      name,
			aliasName: aliasName,
		},
		usingExpressions:     usingExpressions,
		whereExpression:      whereExpression,
		returningExpressions: returningExpressions,
	}

	return builder.Build()
}

// deleteUsing := `USING` tableExpressions
func (p *parser) parseDeleteUsing() ([]queries.Node, error) {
	if !p.advanceIf(isType(tokens.TokenTypeUsing)) {
		return nil, nil
	}

	return p.parseTableExpressions()
}
