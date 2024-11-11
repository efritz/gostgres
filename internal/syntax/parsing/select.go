package parsing

import (
	"strconv"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projector"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/ast"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

func (p *parser) parseSelectBuilder(_ tokens.Token) (ast.BuilderResolver, error) {
	return p.parseSelect()
}

// selectTail := simpleSelect orderBy limitOffset
func (p *parser) parseSelect() (ast.TableReferenceOrExpression, error) {
	simpleSelect, err := p.parseSimpleSelect()
	if err != nil {
		return nil, err
	}

	orderExpression, _, err := p.parseOrderBy()
	if err != nil {
		return nil, err
	}

	limit, offset, err := p.parseLimitOffset()
	if err != nil {
		return nil, err
	}

	builder := &ast.SelectBuilder{
		Select: simpleSelect,
		Order:  orderExpression,
		Limit:  limit,
		Offset: offset,
	}

	return builder, nil
}

// simpleSelect := selectExpressions from where groupBy combinedQuery
func (p *parser) parseSimpleSelect() (*ast.SimpleSelectDescription, error) {
	selectExpressions, err := p.parseSelectExpressions()
	if err != nil {
		return nil, err
	}

	// TODO - make from optional
	node, err := p.parseFrom()
	if err != nil {
		return nil, err
	}

	whereExpression, _, err := p.parseWhere()
	if err != nil {
		return nil, err
	}

	groupings, _, err := p.parseGroupBy()
	if err != nil {
		return nil, err
	}

	combinations, err := p.parseCombinedQuery()
	if err != nil {
		return nil, err
	}

	description := &ast.SimpleSelectDescription{
		SelectExpressions: selectExpressions,
		From:              node,
		Where:             whereExpression,
		Groupings:         groupings,
		Combinations:      combinations,
	}

	return description, nil
}

// selectExpressions := `*` | ( selectExpression [, ...] )
func (p *parser) parseSelectExpressions() (aliasedExpressions []projector.ProjectionExpression, _ error) {
	if p.advanceIf(isType(tokens.TokenTypeAsterisk)) {
		return []projector.ProjectionExpression{projector.NewWildcardProjectionExpression()}, nil
	}

	return parseCommaSeparatedList(p, p.parseSelectExpression)
}

// selectExpression := ( ident `.` `*` ) | ( expression alias )
func (p *parser) parseSelectExpression() (projector.ProjectionExpression, error) {
	nameToken := p.current()
	if p.advanceIf(isType(tokens.TokenTypeIdent), isType(tokens.TokenTypeDot), isType(tokens.TokenTypeAsterisk)) {
		return projector.NewTableWildcardProjectionExpression(nameToken.Text), nil
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

	return projector.NewProjectedExpression(expression, alias), nil
}

// groupBy := [ `GROUP BY` expression [, ...] ]
func (p *parser) parseGroupBy() ([]impls.Expression, bool, error) {
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
func (p *parser) parseCombinedQuery() ([]*ast.CombinationDescription, error) {
	var combinations []*ast.CombinationDescription
	for {
		typ := p.current().Type

		if p.advanceIf(isType(tokens.TokenTypeUnion)) {
		} else if p.advanceIf(isType(tokens.TokenTypeIntersect)) {
		} else if p.advanceIf(isType(tokens.TokenTypeExcept)) {
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

		description := &ast.CombinationDescription{
			Type:     typ,
			Distinct: distinct,
			Select:   unionTarget,
		}

		combinations = append(combinations, description)
	}

	return combinations, nil
}

// combinationTarget := simpleSelect | ( `(` selectOrValues `)` )
func (p *parser) parseCombinationTarget() (ast.TableReferenceOrExpression, error) {
	expectParen := false
	var parseFunc func() (ast.TableReferenceOrExpression, error)

	if p.advanceIf(isType(tokens.TokenTypeLeftParen)) {
		expectParen = true
		parseFunc = p.parseSelectOrValues
	} else {
		parseFunc = func() (ast.TableReferenceOrExpression, error) {
			if _, err := p.mustAdvance(isType(tokens.TokenTypeSelect)); err != nil {
				return nil, err
			}

			description, err := p.parseSimpleSelect()
			if err != nil {
				return nil, err
			}

			builder := &ast.SelectBuilder{
				Select: description,
			}

			return builder, nil
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
func (p *parser) parseOrderBy() (impls.OrderExpression, bool, error) {
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
func (p *parser) parseLimitOffset() (limit, offset *int, _ error) {
	if limitValue, hasLimit, err := p.parseLimit(); err != nil {
		return nil, nil, err
	} else if hasLimit {
		limit = &limitValue
	}

	if offsetValue, hasOffset, err := p.parseOffset(); err != nil {
		return nil, nil, err
	} else if hasOffset {
		offset = &offsetValue
	}

	return limit, offset, nil
}

// limit := [ `LIMIT` expression ]
func (p *parser) parseLimit() (int, bool, error) {
	if !p.advanceIf(isType(tokens.TokenTypeLimit)) {
		return 0, false, nil
	}

	// TODO - can be arbitrary expression
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

	// TODO - can be arbitrary expression
	limitToken, err := p.mustAdvance(isType(tokens.TokenTypeNumber))
	if err != nil {
		return 0, false, err
	}

	limitValue, err := strconv.Atoi(limitToken.Text)
	return limitValue, true, err
}
