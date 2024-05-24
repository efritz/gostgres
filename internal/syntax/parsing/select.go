package parsing

import (
	"fmt"
	"strconv"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/aggregate"
	"github.com/efritz/gostgres/internal/execution/queries/combination"
	"github.com/efritz/gostgres/internal/execution/queries/filter"
	"github.com/efritz/gostgres/internal/execution/queries/limit"
	"github.com/efritz/gostgres/internal/execution/queries/order"
	"github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// selectTail := simpleSelect orderBy limitOffset
func (p *parser) parseSelect(token tokens.Token) (queries.Node, error) {
	node, selectExpressions, err := p.parseSimpleSelect()
	if err != nil {
		return nil, err
	}

	if orderExpression, hasOrder, err := p.parseOrderBy(); err != nil {
		return nil, err
	} else if hasOrder {
		node = order.NewOrder(node, orderExpression)
	}

	node, err = p.parseLimitOffset(node)
	if err != nil {
		return nil, err
	}

	if selectExpressions != nil {
		return projection.NewProjection(node, selectExpressions)
	}

	return node, nil
}

// simpleSelect := selectExpressions from where groupBy combinedQuery
func (p *parser) parseSimpleSelect() (queries.Node, []projection.ProjectionExpression, error) {
	selectExpressions, err := p.parseSelectExpressions()
	if err != nil {
		return nil, nil, err
	}

	// TODO - make from optional
	node, err := p.parseFrom()
	if err != nil {
		return nil, nil, err
	}

	whereExpression, hasWhere, err := p.parseWhere()
	if err != nil {
		return nil, nil, err
	}
	if hasWhere {
		node = filter.NewFilter(node, whereExpression)
	}

	groupings, hasGroupings, err := p.parseGroupBy()
	if err != nil {
		return nil, nil, err
	}
	if hasGroupings {
	selectLoop:
		for _, selectExpression := range selectExpressions {
			expression, alias, ok := projection.UnwrapAlias(selectExpression)
			if !ok {
				return nil, nil, fmt.Errorf("cannot unwrap alias %q", selectExpression)
			}

			if fields := expressions.Fields(expression); len(fields) > 0 {
				for _, grouping := range groupings {
					if grouping.Equal(expression) || grouping.Equal(expressions.NewNamed(shared.NewField("", alias, shared.TypeAny))) {
						continue selectLoop
					}
				}

				// TODO - more lenient validation
				// return nil, nil, fmt.Errorf("%q not in group by", expression)
			}
		}

		node = aggregate.NewHashAggregate(node, groupings, selectExpressions)
		selectExpressions = nil
	}

	return p.parseCombinedQuery(node, selectExpressions)
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

// groupBy := [ `GROUP BY` expression [, ...] ]
func (p *parser) parseGroupBy() ([]expressions.Expression, bool, error) {
	// TODO - make this a combined token?
	if !p.advanceIf(isType(tokens.TokenTypeGroup), isType(tokens.TokenTypeBy)) {
		return nil, false, nil
	}

	groupingExpressions, err := parseCommaSeparatedList(p, p.parseRootExpression)
	if err != nil {
		return nil, false, err
	}

	return groupingExpressions, true, nil
}

// combinedQuery := [ ( ( `UNION` | `INTERSECT` | `EXCEPT` ) [ ( `ALL` | `DISTINCT` ) ] combinationTarget ) [, ...] ]
func (p *parser) parseCombinedQuery(node queries.Node, selectExpressions []projection.ProjectionExpression) (queries.Node, []projection.ProjectionExpression, error) {
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
			return nil, nil, err
		}

		if selectExpressions != nil {
			node, err = projection.NewProjection(node, selectExpressions)
			if err != nil {
				return nil, nil, err
			}

			selectExpressions = nil
		}

		node, err = factory(node, unionTarget, distinct)
		if err != nil {
			return nil, nil, err
		}

	}

	return node, selectExpressions, nil
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

			node, selectExpressions, err := p.parseSimpleSelect()
			if err != nil {
				return nil, err
			}

			return projection.NewProjection(node, selectExpressions)
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

// limitOffset := limit offset
func (p *parser) parseLimitOffset(node queries.Node) (queries.Node, error) {
	limitValue, hasLimit, err := p.parseLimit()
	if err != nil {
		return nil, err
	}
	offsetValue, hasOffset, err := p.parseOffset()
	if err != nil {
		return nil, err
	}

	if hasOffset {
		node = limit.NewOffset(node, offsetValue)
	}
	if hasLimit {
		node = limit.NewLimit(node, limitValue)
	}

	return node, nil
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
