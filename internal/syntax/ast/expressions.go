package ast

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	projectionHelpers "github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/shared/impls"
)

func resolveExpression(
	ctx *impls.NodeResolutionContext,
	expr impls.Expression,
	projection *projectionHelpers.Projection,
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
		mappedExpr, err = expr.Map(func(expr impls.Expression) (impls.Expression, error) {
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
