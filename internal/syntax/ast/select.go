package ast

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	projectionHelpers "github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/aggregate"
	"github.com/efritz/gostgres/internal/execution/queries/combination"
	"github.com/efritz/gostgres/internal/execution/queries/filter"
	"github.com/efritz/gostgres/internal/execution/queries/limit"
	"github.com/efritz/gostgres/internal/execution/queries/order"
	projection "github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

type SelectBuilder struct {
	Select *SimpleSelectDescription
	Order  impls.OrderExpression
	Limit  *int
	Offset *int

	fields           []fields.Field
	projection       *projectionHelpers.Projection
	aggregateFactory impls.AggregateExpressionFactory
}

type SimpleSelectDescription struct {
	SelectExpressions []projectionHelpers.ProjectionExpression
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

func (b *SelectBuilder) Resolve(ctx *impls.NodeResolutionContext) error {
	if err := b.resolvePrimarySelect(ctx); err != nil {
		return err
	}

	if err := b.resolveCombinations(ctx); err != nil {
		return err
	}

	return nil
}

func (b *SelectBuilder) resolvePrimarySelect(ctx *impls.NodeResolutionContext) error {
	if err := ctx.WithScope(func() error {
		return b.Select.From.Resolve(ctx)
	}); err != nil {
		return err
	}

	fromFields := b.Select.From.TableFields()

	ctx.PushScope()
	defer ctx.PopScope()
	ctx.Bind(fromFields...)

	resolved, err := resolveExpression(ctx, b.Select.Where, nil, false)
	if err != nil {
		return err
	}
	b.Select.Where = resolved

	projectedExpressions, err := projectionHelpers.ExpandProjection(fromFields, b.Select.SelectExpressions)
	if err != nil {
		return err
	}
	for i, expr := range projectedExpressions {
		resolved, err := resolveExpression(ctx, expr.Expression, nil, true)
		if err != nil {
			return err
		}

		projectedExpressions[i].Expression = resolved
	}
	projection, err := projectionHelpers.NewProjectionFromProjectedExpressions("", projectedExpressions)
	if err != nil {
		return err
	}
	b.projection = projection
	b.fields = projection.Fields()

	ctx.PushScope()
	defer ctx.PopScope()
	ctx.Bind(b.fields...)

	for i, expr := range b.Select.Groupings {
		resolved, err := resolveExpression(ctx, expr, projection, false)
		if err != nil {
			return err
		}

		b.Select.Groupings[i] = resolved
	}

	var rawProjectedExpressions []impls.Expression
	for _, selectExpression := range projectedExpressions {
		rawProjectedExpressions = append(rawProjectedExpressions, selectExpression.Expression)
	}
	_, nonAggregatedFields, containsAggregate, err := expressions.PartitionAggregatedFieldReferences(
		ctx.ExpressionResolutionContext(true),
		rawProjectedExpressions,
		b.Select.Groupings,
	)
	if err != nil {
		return err
	}

	if len(b.Select.Groupings) == 0 && containsAggregate {
		b.Select.Groupings = []impls.Expression{expressions.NewConstant(nil)}
	}

	if len(b.Select.Groupings) > 0 {
	selectLoop:
		for _, field := range nonAggregatedFields {
			for _, grouping := range b.Select.Groupings {
				if grouping.Equal(expressions.NewNamed(field)) {
					continue selectLoop
				}
			}

			return fmt.Errorf("%q not in group by", field)
		}

		b.aggregateFactory = projectionHelpers.NewAggregateFactory(projectedExpressions)
	}

	if b.Order != nil {
		resolved, err := b.Order.Map(func(expr impls.Expression) (impls.Expression, error) {
			if len(b.Select.Groupings) > 0 {
				return resolveExpression(ctx, expr, nil, false)
			}

			return resolveExpression(ctx, expr, projection, false)
		})
		if err != nil {
			return err
		}

		b.Order = resolved
	}

	return nil
}

func (b *SelectBuilder) resolveCombinations(ctx *impls.NodeResolutionContext) error {
	for _, c := range b.Select.Combinations {
		if err := ctx.WithScope(func() error {
			return c.Select.Resolve(ctx)
		}); err != nil {
			return err
		}

		if len(c.Select.TableFields()) != len(b.fields) {
			return fmt.Errorf("selects in combination must have the same number of columns")
		}
	}

	return nil
}

func (b *SelectBuilder) TableFields() []fields.Field {
	return slices.Clone(b.fields)
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
		node = aggregate.NewHashAggregate(node, b.Select.Groupings, b.aggregateFactory)
	}

	if len(b.Select.Combinations) > 0 {
		if b.Select.Groupings == nil {
			node = projection.NewProjection(node, b.projection)
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

	if b.Select.Groupings == nil && len(b.Select.Combinations) == 0 {
		node = projection.NewProjection(node, b.projection)
	}

	return node, nil
}
