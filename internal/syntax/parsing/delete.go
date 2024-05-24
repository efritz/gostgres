package parsing

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/access"
	"github.com/efritz/gostgres/internal/execution/queries/alias"
	"github.com/efritz/gostgres/internal/execution/queries/filter"
	"github.com/efritz/gostgres/internal/execution/queries/mutation"
	"github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/shared"
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
	node := access.NewAccess(table)
	if aliasName != "" {
		node = alias.NewAlias(node, aliasName)
	}

	usingExpressions, err := p.parseDeleteUsing()
	if err != nil {
		return nil, err
	}
	if len(usingExpressions) > 0 {
		node = joinNodes(append([]queries.Node{node}, usingExpressions...))
	}

	whereExpression, hasWhere, err := p.parseWhere()
	if err != nil {
		return nil, err
	}
	if hasWhere {
		node = filter.NewFilter(node, whereExpression)
	}

	returningExpressions, err := p.parseReturning(name)
	if err != nil {
		return nil, err
	}

	relationName := name
	if aliasName != "" {
		relationName = aliasName
	}
	tidField := shared.NewField(relationName, shared.TIDName, shared.TypeBigInteger)

	node, err = projection.NewProjection(node, []projection.ProjectionExpression{
		projection.NewAliasProjectionExpression(expressions.NewNamed(tidField), shared.TIDName),
	})
	if err != nil {
		return nil, err
	}

	return mutation.NewDelete(node, table, aliasName, returningExpressions)
}

// deleteUsing := `USING` tableExpressions
func (p *parser) parseDeleteUsing() ([]queries.Node, error) {
	if !p.advanceIf(isType(tokens.TokenTypeUsing)) {
		return nil, nil
	}

	return p.parseTableExpressions()
}
