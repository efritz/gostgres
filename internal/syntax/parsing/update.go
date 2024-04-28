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

// update := table `SET` setExpressions [`FROM` tableExpressions] where returning
func (p *parser) parseUpdate(token tokens.Token) (queries.Node, error) {
	table, name, aliasName, err := p.parseTable()
	if err != nil {
		return nil, err
	}
	node := access.NewAccess(table)
	if aliasName != "" {
		node = alias.NewAlias(node, aliasName)
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
		node = joinNodes(append([]queries.Node{node}, fromExpressions...))
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
		projection.NewTableWildcardProjectionExpression(relationName),
	})
	if err != nil {
		return nil, err
	}

	node = alias.NewAlias(node, name)
	return mutation.NewUpdate(node, table, setExpressions, aliasName, returningExpressions)
}

// setExpressions := ( setExpression [, ...] )
func (p *parser) parseSetExpressions() ([]mutation.SetExpression, error) {
	var setExpressions []mutation.SetExpression
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

// TODO - support ident `=` `DEFAULT`
// TODO - support `(` ( ident [, ...] ) `)` = ( `(` sub-select `)` | [ `ROW` ] `(` ( expression | `DEFAULT` [, ...]) `)` )

// setExpression := ident `=` expression
func (p *parser) parseSetExpression() (mutation.SetExpression, error) {
	var name string
	if p.advanceIf(isType(tokens.TokenTypeLeftParen)) {
		panic("Multi-column sets unimplemented")
	} else {
		nameToken, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
		if err != nil {
			return mutation.SetExpression{}, err
		}

		name = nameToken.Text
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeEquals)); err != nil {
		return mutation.SetExpression{}, err
	}

	// TODO - or subselect, or values
	expr, err := p.parseExpression(0)
	if err != nil {
		return mutation.SetExpression{}, err
	}

	return mutation.SetExpression{
		Name:       name,
		Expression: expr,
	}, nil
}
