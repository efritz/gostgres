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

func (b *SelectBuilder) Resolve(ctx impls.ResolutionContext) error {
	if err := b.resolvePrimarySelect(ctx); err != nil {
		return err
	}

	if err := b.resolveCombinations(ctx); err != nil {
		return err
	}

	return nil
}

func (b *SelectBuilder) resolvePrimarySelect(ctx impls.ResolutionContext) error {
	if err := ctx.WithScope(func() error {
		return b.Select.From.Resolve(ctx)
	}); err != nil {
		return err
	}

	fromFields := b.Select.From.TableFields()

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
	type mappedAlias struct {
		alias string
		expr  impls.Expression
	}
	var mappedAliases []mappedAlias
	for _, a := range aliases {
		e, err := a.Expression.Map(ctx.ResolveExpression)
		if err != nil {
			return err
		}

		mappedAliases = append(mappedAliases, mappedAlias{alias: a.Alias, expr: e})
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

		// TODO - vet
		e, err = e.Map(func(e impls.Expression) (impls.Expression, error) {
			if named, ok := e.(expressions.NamedExpression); ok {
				for _, a := range mappedAliases {
					if a.alias == named.Field().Name() {
						return a.expr, nil
					}
				}
			}

			return e, nil
		})
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

func (b *SelectBuilder) resolveCombinations(ctx *context.ResolveContext) error {
	for _, c := range b.Select.Combinations {
		if err := ctx.WithScope(func() error {
			return c.Select.Resolve(ctx)
		}); err != nil {
			return err
		}

		combinationFields := c.Select.TableFields()

		if len(combinationFields) != len(b.fields) {
			return fmt.Errorf("union tables have different number of columns")
		}
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
					if grouping.Equal(expression) || grouping.Equal(expressions.NewNamed(fields.NewField("", alias, types.TypeAny, fields.NonInternalField))) {
						continue selectLoop
					}
				}

				// TODO - more lenient validation
				// return nil, fmt.Errorf("%q not in group by", expression)
			}
		}

		node = aggregate.NewHashAggregate(node, b.Select.Groupings, b.Select.SelectExpressions)
	}

	if len(b.Select.Combinations) > 0 {
		node, err = projection.NewProjection(node, b.Select.SelectExpressions)
		if err != nil {
			return nil, err
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
