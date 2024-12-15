package ast

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

func ResolveExpression(
	ctx *impls.NodeResolutionContext,
	expr impls.Expression,
	projection *projection.Projection,
	allowAggregateFunctions bool,
) (impls.Expression, error) {
	if expr == nil {
		return nil, nil
	}

	mappedExpr, err := expr.Map(func(expr impls.Expression) (impls.Expression, error) {
		if named, ok := expr.(expressions.NamedExpression); ok {
			field, err := ctx.Lookup(named.Field().RelationName(), named.Field().Name())
			if err != nil {
				return nil, err
			}

			return expressions.NewNamed(field), nil
		}

		return expr, nil
	})
	if err != nil {
		return nil, err
	}

	if projection != nil {
		mappedExpr, err = mappedExpr.Map(func(expr impls.Expression) (impls.Expression, error) {
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
			return nil, err
		}
	}

	if err := mappedExpr.Resolve(ctx.ExpressionResolutionContext(allowAggregateFunctions)); err != nil {
		return nil, err
	}

	return mappedExpr, nil
}

func ResolveProjection(
	ctx *impls.NodeResolutionContext,
	name string,
	fields []fields.Field,
	projectionExpressions []projection.ProjectionExpression,
	tableAliases []projection.AliasedTable,
) (*projection.Projection, error) {
	projectedExpressions, err := projection.ExpandProjection(fields, projectionExpressions, tableAliases)
	if err != nil {
		return nil, err
	}

	for i, expr := range projectedExpressions {
		resolved, err := ResolveExpression(ctx, expr.Expression, nil, true)
		if err != nil {
			return nil, err
		}

		projectedExpressions[i].Expression = resolved
	}

	return projection.NewProjection(name, projectedExpressions)
}
