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

// updateTail := table `SET` ( setExpression [, ...] ) [ `FROM` tableExpressions ] where returning
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

	setExpressions, err := parseCommaSeparatedList(p, p.parseSetExpression)
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

// setExpression := ident `=` expression
func (p *parser) parseSetExpression() (mutation.SetExpression, error) {
	if p.advanceIf(isType(tokens.TokenTypeLeftParen)) {
		// TODO - support sub-select
		// TODO - support row values
		panic("Multi-column sets unimplemented")
	}

	name, err := p.parseIdent()
	if err != nil {
		return mutation.SetExpression{}, err
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeEquals)); err != nil {
		return mutation.SetExpression{}, err
	}

	expr, err := p.parseRootExpression()
	if err != nil {
		return mutation.SetExpression{}, err
	}

	return mutation.SetExpression{
		Name:       name,
		Expression: expr,
	}, nil
}
