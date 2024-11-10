package expressions

import (
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

func NewIsNull(expression impls.Expression) impls.Expression {
	typeChecker := func(expression types.Type) (types.Type, error) {
		return types.TypeBool, nil
	}

	valueFrom := func(ctx impls.ExecutionContext, expression impls.Expression, row rows.Row) (any, error) {
		val, err := expression.ValueFrom(ctx, row)
		if err != nil {
			return nil, err
		}
		return val == nil, nil
	}

	return newUnaryExpression(expression, "is null", typeChecker, valueFrom)
}

func NewIsTrue(expression impls.Expression) impls.Expression {
	typeChecker := func(expression types.Type) (types.Type, error) {
		return types.TypeBool, nil
	}

	valueFrom := func(ctx impls.ExecutionContext, expression impls.Expression, row rows.Row) (any, error) {
		val, err := types.ValueAs[bool](expression.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}
		return val != nil && *val, nil
	}

	return newUnaryExpression(expression, "is true", typeChecker, valueFrom)
}

func NewIsFalse(expression impls.Expression) impls.Expression {
	typeChecker := func(expression types.Type) (types.Type, error) {
		return types.TypeBool, nil
	}

	valueFrom := func(ctx impls.ExecutionContext, expression impls.Expression, row rows.Row) (any, error) {
		val, err := types.ValueAs[bool](expression.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}
		return val != nil && !*val, nil
	}

	return newUnaryExpression(expression, "is false", typeChecker, valueFrom)
}

func NewIsUnknown(expression impls.Expression) impls.Expression {
	typeChecker := func(expression types.Type) (types.Type, error) {
		return types.TypeBool, nil
	}

	valueFrom := func(ctx impls.ExecutionContext, expression impls.Expression, row rows.Row) (any, error) {
		val, err := types.ValueAs[bool](expression.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}
		return val == nil, nil
	}

	return newUnaryExpression(expression, "is unknown", typeChecker, valueFrom)
}
