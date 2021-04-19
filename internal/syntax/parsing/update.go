package parsing

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/nodes"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// update := table `SET` setExpressions [`FROM` tableExpressions] [`RETURNING` selectExpressions]
func (p *parser) parseUpdate(token tokens.Token) (nodes.Node, error) {
	table, name, alias, err := p.parseTable()
	if err != nil {
		return nil, err
	}
	node := nodes.NewData(table)
	if alias != "" {
		node = nodes.NewAlias(node, alias)
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeSet)); err != nil {
		return nil, err
	}

	setExpressions, err := p.parseSetExpressions()
	if err != nil {
		return nil, err
	}

	if p.advanceIf(isType(tokens.TokenTypeFrom)) {
		fromExpressions, err := p.parseTableExpressions()
		if err != nil {
			return nil, err
		}
		node = joinNodes(append([]nodes.Node{node}, fromExpressions...))
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
		nodes.NewTableWildcardProjectionExpression(relationName),
	})
	if err != nil {
		return nil, err
	}

	node = nodes.NewAlias(node, name)
	return nodes.NewUpdate(node, table, setExpressions, alias, returningExpressions)
}

// setExpressions := ( setExpression [, ...] )
func (p *parser) parseSetExpressions() ([]nodes.SetExpression, error) {
	var setExpressions []nodes.SetExpression
	for {
		setExpression, err := p.parseSetExpression()
		if err != nil {
			return nil, err
		}

		setExpressions = append(setExpressions, setExpression)

		if !p.advanceIf(isType(tokens.TokenTypeComma)) {
			break
		}
	}

	return setExpressions, nil
}

// setExpression := ident `=` expression
func (p *parser) parseSetExpression() (nodes.SetExpression, error) {
	var name string
	if p.advanceIf(isType(tokens.TokenTypeLeftParen)) {
		// TODO - implement
		panic("Multi-column sets unimplemented")
	} else {
		nameToken, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
		if err != nil {
			return nodes.SetExpression{}, err
		}

		name = nameToken.Text
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeEquals)); err != nil {
		return nodes.SetExpression{}, err
	}

	// TODO - or subselect
	expr, err := p.parseExpression(0)
	if err != nil {
		return nodes.SetExpression{}, err
	}

	return nodes.SetExpression{
		Name:       name,
		Expression: expr,
	}, nil
}
