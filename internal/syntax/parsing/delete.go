package parsing

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/queries/access"
	"github.com/efritz/gostgres/internal/queries/alias"
	"github.com/efritz/gostgres/internal/queries/filter"
	"github.com/efritz/gostgres/internal/queries/mutation"
	"github.com/efritz/gostgres/internal/queries/projection"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// delete := `FROM` table using where returning
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

	usingExpressions, err := p.parseUsing()
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
	tidField := shared.NewField(relationName, shared.TIDName, shared.TypeKindNumeric)

	node, err = projection.NewProjection(node, []projection.ProjectionExpression{
		projection.NewAliasProjectionExpression(expressions.NewNamed(tidField), shared.TIDName),
	})
	if err != nil {
		return nil, err
	}

	return mutation.NewDelete(node, table, aliasName, returningExpressions)
}

// using := `USING` tableExpressions
func (p *parser) parseUsing() ([]queries.Node, error) {
	if !p.advanceIf(isType(tokens.TokenTypeUsing)) {
		return nil, nil
	}

	return p.parseTableExpressions()
}
