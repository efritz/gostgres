package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/aggregate"
	"github.com/efritz/gostgres/internal/execution/queries/combination"
	"github.com/efritz/gostgres/internal/execution/queries/filter"
	"github.com/efritz/gostgres/internal/execution/queries/limit"
	"github.com/efritz/gostgres/internal/execution/queries/order"
	"github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/types"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

type SelectBuilder struct {
	SimpleSelect    *SimpleSelectDescription
	OrderExpression impls.OrderExpression
	Limit           *int
	Offset          *int
}

type SimpleSelectDescription struct {
	SelectExpressions []projection.ProjectionExpression
	From              TableExpressionDescription
	WhereExpression   impls.Expression
	Groupings         []impls.Expression
	Combinations      []*CombinationDescription
}

type CombinationDescription struct {
	Type                    tokens.TokenType
	Distinct                bool
	SimpleSelectDescription BaseTableExpressionDescription
}

func (b *SelectBuilder) Build(ctx BuildContext) (queries.Node, error) {
	return b.TableExpression(ctx)
}

func (b SelectBuilder) TableExpression(ctx BuildContext) (queries.Node, error) {
	node, err := b.SimpleSelect.From.TableExpression(ctx)
	if err != nil {
		return nil, err
	}

	if b.SimpleSelect.WhereExpression != nil {
		node = filter.NewFilter(node, b.SimpleSelect.WhereExpression)
	}

	if b.SimpleSelect.Groupings != nil {
	selectLoop:
		for _, selectExpression := range b.SimpleSelect.SelectExpressions {
			expression, alias, ok := projection.UnwrapAlias(selectExpression)
			if !ok {
				return nil, fmt.Errorf("cannot unwrap alias %q", selectExpression)
			}

			if len(expressions.Fields(expression)) > 0 {
				for _, grouping := range b.SimpleSelect.Groupings {
					if grouping.Equal(expression) || grouping.Equal(expressions.NewNamed(fields.NewField("", alias, types.TypeAny))) {
						continue selectLoop
					}
				}

				// TODO - more lenient validation
				// return nil,  fmt.Errorf("%q not in group by", expression)
			}
		}

		node = aggregate.NewHashAggregate(node, b.SimpleSelect.Groupings, b.SimpleSelect.SelectExpressions)
		b.SimpleSelect.SelectExpressions = nil
	}

	if len(b.SimpleSelect.Combinations) != 0 {
		if b.SimpleSelect.SelectExpressions != nil {
			newNode, err := projection.NewProjection(node, b.SimpleSelect.SelectExpressions)
			if err != nil {
				return nil, err
			}
			node = newNode
			b.SimpleSelect.SelectExpressions = nil
		}

		for _, c := range b.SimpleSelect.Combinations {
			var factory func(left, right queries.Node, distinct bool) (queries.Node, error)
			switch c.Type {
			case tokens.TokenTypeUnion:
				factory = combination.NewUnion
			case tokens.TokenTypeIntersect:
				factory = combination.NewIntersect
			case tokens.TokenTypeExcept:
				factory = combination.NewExcept
			}

			right, err := c.SimpleSelectDescription.Build(ctx)
			if err != nil {
				return nil, err
			}

			newNode, err := factory(node, right, c.Distinct)
			if err != nil {
				return nil, err
			}
			node = newNode
		}
	}

	if b.OrderExpression != nil {
		node = order.NewOrder(node, b.OrderExpression)
	}
	if b.Offset != nil {
		node = limit.NewOffset(node, *b.Offset)
	}
	if b.Limit != nil {
		node = limit.NewLimit(node, *b.Limit)
	}

	if b.SimpleSelect.SelectExpressions != nil {
		return projection.NewProjection(node, b.SimpleSelect.SelectExpressions)
	}
	return node, nil
}
