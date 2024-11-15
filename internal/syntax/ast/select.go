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
	"github.com/efritz/gostgres/internal/shared/types"
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

	resolved, err := resolveExpression(ctx, b.Select.Where)
	if err != nil {
		return err
	}
	b.Select.Where = resolved

	// TODO - map the resulting expressions?
	projection, err := projectionHelpers.NewProjection("", fromFields, b.Select.SelectExpressions)
	if err != nil {
		return err
	}

	if b.Select.Groupings != nil {
	selectLoop:
		for _, selectExpression := range projection.Aliases() {
			if len(expressions.Fields(selectExpression.Expression)) > 0 {
				alias := expressions.NewNamed(fields.NewField("", selectExpression.Alias, types.TypeAny, fields.NonInternalField))

				for _, grouping := range b.Select.Groupings {
					if grouping.Equal(selectExpression.Expression) || grouping.Equal(alias) {
						continue selectLoop
					}
				}

				// TODO - more lenient validation
				// return nil,  fmt.Errorf("%q not in group by", expression)
			}
		}
	}

	b.fields = projection.Fields()

	ctx.PushScope()
	defer ctx.PopScope()
	ctx.Bind(b.fields...)

	for i, expr := range b.Select.Groupings {
		resolved, err := resolveExpression(ctx, expr)
		if err != nil {
			return err
		}

		resolved, err = resolved.Map(func(expr impls.Expression) (impls.Expression, error) {
			if named, ok := expr.(expressions.NamedExpression); ok {
				for _, pair := range projection.Aliases() {
					if pair.Alias == named.Field().Name() {
						return pair.Expression, nil
					}
				}
			}

			return expr, nil
		})
		if err != nil {
			return err
		}

		b.Select.Groupings[i] = resolved
	}

	if b.Order != nil {
		resolved, err := b.Order.Map(func(expr impls.Expression) (impls.Expression, error) {
			return resolveExpression(ctx, expr)
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
		node = aggregate.NewHashAggregate(node, b.Select.Groupings, b.Select.SelectExpressions)
	}

	if len(b.Select.Combinations) > 0 {
		if b.Select.Groupings == nil {
			node, err = projection.NewProjection(node, b.Select.SelectExpressions)
			if err != nil {
				return nil, err
			}
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
		node, err = projection.NewProjection(node, b.Select.SelectExpressions)
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}
