package parsing

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/nodes"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// delete := `FROM` table using where returning
func (p *parser) parseDelete(token tokens.Token) (nodes.Node, error) {
	if _, err := p.mustAdvance(isType(tokens.TokenTypeFrom)); err != nil {
		return nil, err
	}

	table, name, alias, err := p.parseTable()
	if err != nil {
		return nil, err
	}
	node := nodes.NewData(table)
	if alias != "" {
		node = nodes.NewAlias(node, alias)
	}

	usingExpressions, err := p.parseUsing()
	if err != nil {
		return nil, err
	}
	if len(usingExpressions) > 0 {
		node = joinNodes(append([]nodes.Node{node}, usingExpressions...))
	}

	whereExpression, hasWhere, err := p.parseWhereClause()
	if err != nil {
		return nil, err
	}
	if hasWhere {
		node = nodes.NewFilter(node, whereExpression)
	}

	returningExpressions, err := p.parseReturning(name)
	if err != nil {
		return nil, err
	}

	relationName := name
	if alias != "" {
		relationName = alias
	}
	tidField := shared.NewField(relationName, "tid", shared.TypeKindNumeric, false)

	node, err = nodes.NewProjection(node, []nodes.ProjectionExpression{
		nodes.NewAliasProjectionExpression(expressions.NewNamed(tidField), "tid"),
	})
	if err != nil {
		return nil, err
	}

	return nodes.NewDelete(node, table, alias, returningExpressions)
}

// using := `USING` tableExpressions
func (p *parser) parseUsing() ([]nodes.Node, error) {
	if _, err := p.mustAdvance(isType(tokens.TokenTypeUsing)); err != nil {
		return nil, err
	}

	return p.parseTableExpressions()
}
