package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projector"
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
	Select *SimpleSelectDescription
	Order  impls.OrderExpression
	Limit  *int
	Offset *int
}

func (b SelectBuilder) TableExpression() {}

func (b *SelectBuilder) Resolve(ctx ResolveContext) ([]fields.Field, error) {
	fields, err := b.Select.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	// TODO
	return fields, nil
}

func (b *SelectBuilder) Build() (queries.Node, error) {
	node, err := b.Select.Build()
	if err != nil {
		return nil, err
	}

	if b.Order != nil {
		node = order.NewOrder(node, b.Order)
	}
	if b.Offset != nil {
		node = limit.NewOffset(node, *b.Offset)
	}
	if b.Limit != nil {
		node = limit.NewLimit(node, *b.Limit)
	}

	// TODO - must come last? Can order come first?
	if b.Select.SelectExpressions != nil {
		return projection.NewProjection(node, b.Select.SelectExpressions)
	}
	return node, nil
}

//
//

type SimpleSelectDescription struct {
	SelectExpressions []projector.ProjectionExpression
	From              TableExpression
	Where             impls.Expression
	Groupings         []impls.Expression
	Combinations      []*CombinationDescription
}

type CombinationDescription struct {
	Type     tokens.TokenType
	Distinct bool
	Select   TableReferenceOrExpression
}

func (b *SimpleSelectDescription) Resolve(ctx ResolveContext) ([]fields.Field, error) {
	fields, err := b.From.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	for _, t := range b.Combinations {
		fs, err := t.Select.Resolve(ctx)
		if err != nil {
			return nil, err
		}

		_ = fs // TODO - compare types
	}

	return fields, nil
}

func (b *SimpleSelectDescription) Build() (queries.Node, error) {
	node, err := b.From.Build()
	if err != nil {
		return nil, err
	}

	if b.Where != nil {
		node = filter.NewFilter(node, b.Where)
	}

	if b.Groupings != nil {
	selectLoop:
		for _, selectExpression := range b.SelectExpressions {
			expression, alias, ok := projector.UnwrapAlias(selectExpression)
			if !ok {
				return nil, fmt.Errorf("cannot unwrap alias %q", selectExpression)
			}

			if len(expressions.Fields(expression)) > 0 {
				for _, grouping := range b.Groupings {
					if grouping.Equal(expression) || grouping.Equal(expressions.NewNamed(fields.NewField("", alias, types.TypeAny))) {
						continue selectLoop
					}
				}

				// TODO - more lenient validation
				// return nil,  fmt.Errorf("%q not in group by", expression)
			}
		}

		node = aggregate.NewHashAggregate(node, b.Groupings, b.SelectExpressions)
		b.SelectExpressions = nil
	}

	if len(b.Combinations) != 0 {
		if b.SelectExpressions != nil {
			newNode, err := projection.NewProjection(node, b.SelectExpressions)
			if err != nil {
				return nil, err
			}
			node = newNode
			b.SelectExpressions = nil
		}

		for _, c := range b.Combinations {
			var factory func(left, right queries.Node, distinct bool) (queries.Node, error)
			switch c.Type {
			case tokens.TokenTypeUnion:
				factory = combination.NewUnion
			case tokens.TokenTypeIntersect:
				factory = combination.NewIntersect
			case tokens.TokenTypeExcept:
				factory = combination.NewExcept
			}

			right, err := c.Select.Build()
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

	return node, nil
}
