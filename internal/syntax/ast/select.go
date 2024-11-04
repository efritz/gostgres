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
	"github.com/efritz/gostgres/internal/syntax/ast/context"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

type SelectBuilder struct {
	Select *SimpleSelectDescription
	Order  impls.OrderExpression
	Limit  *int
	Offset *int

	fields []fields.Field
}

type SimpleSelectDescription struct {
	SelectExpressions []projector.ProjectionExpression
	From              *TableExpression
	Where             impls.Expression
	Groupings         []impls.Expression
	Combinations      []*CombinationDescription
}

type CombinationDescription struct {
	Type     tokens.TokenType
	Distinct bool
	Select   TableReferenceOrExpression
}

func (b *SelectBuilder) Resolve(ctx *context.ResolveContext) error {
	ctx.PushScope()
	err := b.Select.From.Resolve(ctx)
	ctx.PopScope()
	if err != nil {
		return err
	}

	fromFields := b.Select.From.TableFields()

	// TODO - figure out scoping rule here
	for _, c := range b.Select.Combinations {
		ctx.PushScope()
		err := c.Select.Resolve(ctx)
		ctx.PopScope()
		if err != nil {
			return err
		}

		if len(c.Select.TableFields()) != len(fromFields) {
			return fmt.Errorf("union tables have different number of columns")
		}
	}

	ctx.PushScope()
	defer ctx.PopScope()

	ctx.Bind(fromFields...)

	if b.Select.Where != nil {
		e, err := ctx.ResolveExpression(b.Select.Where)
		if err != nil {
			return err
		}
		b.Select.Where = e
	}

	aliases, err := projector.ExpandProjection(fromFields, b.Select.SelectExpressions)
	if err != nil {
		return err
	}
	for _, a := range aliases {
		e, err := a.Map(ctx.ResolveExpression)
		if err != nil {
			return err
		}
		_ = e // TODO
	}
	// TODO - store aliases?
	b.fields = projector.FieldsFromProjection("", aliases)

	ctx.PushScope()
	defer ctx.PopScope()
	ctx.Bind(b.fields...)

	for i, g := range b.Select.Groupings {
		e, err := ctx.ResolveExpression(g)
		if err != nil {
			return err
		}

		b.Select.Groupings[i] = e
	}

	if b.Order != nil {
		os, err := b.Order.Map(ctx.ResolveExpression)
		if err != nil {
			return err
		}
		b.Order = os
	}

	return nil
}

func (b SelectBuilder) TableFields() []fields.Field {
	return b.fields
}

func (b *SelectBuilder) Build() (queries.Node, error) {
	node, err := b.Select.From.Build()
	if err != nil {
		return nil, err
	}

	if b.Select.Where != nil {
		node = filter.NewFilter(node, b.Select.Where)
	}

	if b.Select.Groupings != nil {
	selectLoop:
		for _, selectExpression := range b.Select.SelectExpressions {
			expression, alias, ok := projector.UnwrapAlias(selectExpression)
			if !ok {
				return nil, fmt.Errorf("cannot unwrap alias %q", selectExpression)
			}

			if len(expressions.Fields(expression)) > 0 {
				for _, grouping := range b.Select.Groupings {
					if grouping.Equal(expression) || grouping.Equal(expressions.NewNamed(fields.NewField("", alias, types.TypeAny))) {
						continue selectLoop
					}
				}

				// TODO - more lenient validation
				// return nil,  fmt.Errorf("%q not in group by", expression)
			}
		}

		node = aggregate.NewHashAggregate(node, b.Select.Groupings, b.Select.SelectExpressions)
		b.Select.SelectExpressions = nil
	}

	if len(b.Select.Combinations) != 0 {
		if b.Select.SelectExpressions != nil {
			newNode, err := projection.NewProjection(node, b.Select.SelectExpressions)
			if err != nil {
				return nil, err
			}
			node = newNode
			b.Select.SelectExpressions = nil
		}

		for _, c := range b.Select.Combinations {
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

	if b.Order != nil {
		node = order.NewOrder(node, b.Order)
	}
	if b.Offset != nil {
		node = limit.NewOffset(node, *b.Offset)
	}
	if b.Limit != nil {
		node = limit.NewLimit(node, *b.Limit)
	}

	if b.Select.SelectExpressions != nil {
		return projection.NewProjection(node, b.Select.SelectExpressions)
	}
	return node, nil
}
