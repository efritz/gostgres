package ast

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/impls"
)

func resolveExpression(ctx *impls.NodeResolutionContext, expr impls.Expression) (impls.Expression, error) {
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

	if err := expr.Resolve(ctx.ExpressionResolutionContext()); err != nil {
		return nil, err
	}

	return mappedExpr, nil
}
