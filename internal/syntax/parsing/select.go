package parsing

import (
	"strconv"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/queries/combination"
	"github.com/efritz/gostgres/internal/queries/filter"
	"github.com/efritz/gostgres/internal/queries/limit"
	"github.com/efritz/gostgres/internal/queries/order"
	"github.com/efritz/gostgres/internal/queries/projection"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// selectTail := simpleSelect orderBy limit offset
func (p *parser) parseSelect(token tokens.Token) (queries.Node, error) {
	selectNode, err := p.parseSimpleSelect()
	if err != nil {
		return nil, err
	}

	orderExpression, hasOrder, err := p.parseOrderBy()
	if err != nil {
		return nil, err
	}
	limitValue, hasLimit, err := p.parseLimit()
	if err != nil {
		return nil, err
	}
	offsetValue, hasOffset, err := p.parseOffset()
	if err != nil {
		return nil, err
	}

	node := selectNode.node
	if hasOrder {
		node = order.NewOrder(node, orderExpression)
	}
	if hasOffset {
		node = limit.NewOffset(node, offsetValue)
	}
	if hasLimit {
		node = limit.NewLimit(node, limitValue)
	}

	node, err = projection.NewProjection(node, selectNode.selectExpressions)
	if err != nil {
		return nil, err
	}

	return node, nil
}

type selectNode struct {
	node              queries.Node
	selectExpressions []projection.ProjectionExpression
}

// simpleSelect := selectExpressions from where combinedQuery
func (p *parser) parseSimpleSelect() (selectNode, error) {
	selectExpressions, err := p.parseSelectExpressions()
	if err != nil {
		return selectNode{}, err
	}

	// TODO - make from optional
	node, err := p.parseFrom()
	if err != nil {
		return selectNode{}, err
	}

	whereExpression, hasWhere, err := p.parseWhere()
	if err != nil {
		return selectNode{}, err
	}
	if hasWhere {
		node = filter.NewFilter(node, whereExpression)
	}

	if p.current().Type != tokens.TokenTypeUnion && p.current().Type != tokens.TokenTypeIntersect && p.current().Type != tokens.TokenTypeExcept {
		return selectNode{
			node:              node,
			selectExpressions: selectExpressions,
		}, nil
	}

	node, err = projection.NewProjection(node, selectExpressions)
	if err != nil {
		return selectNode{}, err
	}
	node, err = p.parseCombinedQuery(node)
	if err != nil {
		return selectNode{}, err
	}

	return selectNode{
		node:              node,
		selectExpressions: []projection.ProjectionExpression{projection.NewWildcardProjectionExpression()},
	}, nil
}

// selectExpressions := `*` | ( selectExpression [, ...] )
func (p *parser) parseSelectExpressions() (aliasedExpressions []projection.ProjectionExpression, _ error) {
	if p.advanceIf(isType(tokens.TokenTypeAsterisk)) {
		return []projection.ProjectionExpression{projection.NewWildcardProjectionExpression()}, nil
	}

	return parseCommaSeparatedList(p, p.parseSelectExpression)
}

// selectExpression := ( ident `.` `*` ) | ( expression alias )
func (p *parser) parseSelectExpression() (projection.ProjectionExpression, error) {
	nameToken := p.current()
	if p.advanceIf(isType(tokens.TokenTypeIdent), isType(tokens.TokenTypeDot), isType(tokens.TokenTypeAsterisk)) {
		return projection.NewTableWildcardProjectionExpression(nameToken.Text), nil
	}

	expression, err := p.parseRootExpression()
	if err != nil {
		return nil, err
	}

	type named interface {
		Name() string
	}
	var alias string
	if value, ok, err := p.parseAlias(); err != nil {
		return nil, err
	} else if ok {
		alias = value
	} else if named, ok := expression.(named); ok {
		alias = named.Name()
	} else {
		alias = "?column?"
	}

	return projection.NewAliasProjectionExpression(expression, alias), nil
}

// combinedQuery := [ ( ( `UNION` | `INTERSECT` | `EXCEPT` ) [ ( `ALL` | `DISTINCT` ) ] combinationTarget ) [, ...] ]
func (p *parser) parseCombinedQuery(node queries.Node) (queries.Node, error) {
	for {
		var factory func(left, right queries.Node, distinct bool) (queries.Node, error)
		if p.advanceIf(isType(tokens.TokenTypeUnion)) {
			factory = combination.NewUnion
		} else if p.advanceIf(isType(tokens.TokenTypeIntersect)) {
			factory = combination.NewIntersect
		} else if p.advanceIf(isType(tokens.TokenTypeExcept)) {
			factory = combination.NewExcept
		} else {
			break
		}

		distinct := true
		if p.advanceIf(isType(tokens.TokenTypeDistinct)) {
			// token is explicitly supplying the default
		} else if p.advanceIf(isType(tokens.TokenTypeAll)) {
			distinct = false
		}

		unionTarget, err := p.parseCombinationTarget()
		if err != nil {
			return nil, err
		}

		node, err = factory(node, unionTarget, distinct)
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

// combinationTarget := simpleSelect | ( `(` selectOrValues `)` )
func (p *parser) parseCombinationTarget() (queries.Node, error) {
	expectParen := false
	var parseFunc func() (queries.Node, error)

	if p.advanceIf(isType(tokens.TokenTypeLeftParen)) {
		expectParen = true
		parseFunc = p.parseSelectOrValues
	} else {
		parseFunc = func() (queries.Node, error) {
			if _, err := p.mustAdvance(isType(tokens.TokenTypeSelect)); err != nil {
				return nil, err
			}

			selectNode, err := p.parseSimpleSelect()
			if err != nil {
				return nil, err
			}

			return projection.NewProjection(selectNode.node, selectNode.selectExpressions)
		}
	}

	node, err := parseFunc()
	if err != nil {
		return nil, err
	}

	if expectParen {
		if _, err := p.mustAdvance(isType(tokens.TokenTypeRightParen)); err != nil {
			return nil, err
		}
	}

	return node, nil
}

// orderBy := [ `ORDER BY` ( expressionWithDirection [, ...] ) ]
func (p *parser) parseOrderBy() (expressions.OrderExpression, bool, error) {
	// TODO - make this a combined token?
	if !p.advanceIf(isType(tokens.TokenTypeOrder), isType(tokens.TokenTypeBy)) {
		return nil, false, nil
	}

	orderExpressions, err := parseCommaSeparatedList(p, p.parseExpressionWithDirection)
	if err != nil {
		return nil, false, err
	}

	// TODO - support `USING` operator
	// TODO - support [`NULLS` ( `FIRST` | `LAST` )]
	return expressions.NewOrderExpression(orderExpressions), true, nil
}

// limit := [ `LIMIT` expression ]
func (p *parser) parseLimit() (int, bool, error) {
	if !p.advanceIf(isType(tokens.TokenTypeLimit)) {
		return 0, false, nil
	}

	limitToken, err := p.mustAdvance(isType(tokens.TokenTypeNumber))
	if err != nil {
		return 0, false, err
	}

	limitValue, err := strconv.Atoi(limitToken.Text)
	return limitValue, true, err
}

// offset := [ `OFFSET` expression ]
func (p *parser) parseOffset() (int, bool, error) {
	if !p.advanceIf(isType(tokens.TokenTypeOffset)) {
		return 0, false, nil
	}

	limitToken, err := p.mustAdvance(isType(tokens.TokenTypeNumber))
	if err != nil {
		return 0, false, err
	}

	limitValue, err := strconv.Atoi(limitToken.Text)
	return limitValue, true, err
}
