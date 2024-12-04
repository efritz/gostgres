package ast

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

type SelectBuilder struct {
	SelectExpressions []projection.ProjectionExpression
	From              *TableExpression
	Where             impls.Expression
	Groupings         []impls.Expression
	Combinations      []*CombinationDescription
	Order             impls.OrderExpression
	Limit             *int
	Offset            *int

	fields     []fields.Field
	projection *projection.Projection
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
		return b.From.Resolve(ctx)
	}); err != nil {
		return err
	}

	fromFields := b.From.TableFields()

	ctx.PushScope()
	defer ctx.PopScope()
	ctx.Bind(fromFields...)

	resolved, err := ResolveExpression(ctx, b.Where, nil, false)
	if err != nil {
		return err
	}
	b.Where = resolved

	projectedExpressions, err := projection.ExpandProjection(fromFields, b.SelectExpressions)
	if err != nil {
		return err
	}
	for i, expr := range projectedExpressions {
		resolved, err := ResolveExpression(ctx, expr.Expression, nil, true)
		if err != nil {
			return err
		}

		projectedExpressions[i].Expression = resolved
	}
	projection, err := projection.NewProjectionFromProjectedExpressions("", projectedExpressions)
	if err != nil {
		return err
	}
	b.projection = projection
	b.fields = projection.Fields()

	ctx.PushScope()
	defer ctx.PopScope()
	ctx.Bind(b.fields...)

	for i, expr := range b.Groupings {
		resolved, err := ResolveExpression(ctx, expr, projection, false)
		if err != nil {
			return err
		}

		b.Groupings[i] = resolved
	}

	var rawProjectedExpressions []impls.Expression
	for _, selectExpression := range projectedExpressions {
		rawProjectedExpressions = append(rawProjectedExpressions, selectExpression.Expression)
	}
	_, nonAggregatedFields, containsAggregate, err := expressions.PartitionAggregatedFieldReferences(
		ctx.ExpressionResolutionContext(true),
		rawProjectedExpressions,
		b.Groupings,
	)
	if err != nil {
		return err
	}

	if len(b.Groupings) == 0 && containsAggregate {
		b.Groupings = []impls.Expression{expressions.NewConstant(nil)}
	}

	if len(b.Groupings) > 0 {
	selectLoop:
		for _, field := range nonAggregatedFields {
			for _, grouping := range b.Groupings {
				if grouping.Equal(expressions.NewNamed(field)) {
					continue selectLoop
				}
			}

			return fmt.Errorf("%q not in group by", field)
		}
	}

	if b.Order != nil {
		resolved, err := b.Order.Map(func(expr impls.Expression) (impls.Expression, error) {
			if len(b.Groupings) > 0 {
				return ResolveExpression(ctx, expr, nil, false)
			}

			return ResolveExpression(ctx, expr, projection, false)
		})
		if err != nil {
			return err
		}

		b.Order = resolved
	}

	return nil
}

func (b *SelectBuilder) resolveCombinations(ctx *impls.NodeResolutionContext) error {
	for _, c := range b.Combinations {
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

func (b *SelectBuilder) Build() (plan.LogicalNode, error) {
	node, err := b.From.Build()
	if err != nil {
		return nil, err
	}

	if len(b.Combinations) > 0 {
		node = plan.NewSelect(
			node,
			b.projection,
			b.Groupings,
			b.Where,
			nil,
			nil,
			nil,
		)

		for _, c := range b.Combinations {
			var factory func(left, right plan.LogicalNode, distinct bool) (plan.LogicalNode, error)
			switch c.Type {
			case tokens.TokenTypeUnion:
				factory = plan.NewUnion
			case tokens.TokenTypeIntersect:
				factory = plan.NewIntersect
			case tokens.TokenTypeExcept:
				factory = plan.NewExcept
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

		return plan.NewSelect(
			node,
			nil,
			nil,
			nil,
			b.Order,
			b.Limit,
			b.Offset,
		), nil
	} else {
		node = plan.NewSelect(
			node,
			b.projection,
			b.Groupings,
			b.Where,
			b.Order,
			b.Limit,
			b.Offset,
		)

		return node, nil
	}
}
