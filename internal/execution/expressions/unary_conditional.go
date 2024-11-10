package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

func NewNot(expression impls.Expression) impls.Expression {
	typeChecker := func(expression types.Type) (types.Type, error) {
		if expression == types.TypeBool {
			return types.TypeBool, nil
		}

		return types.TypeUnknown, fmt.Errorf("illegal operand types for not: %s", expression)
	}

	valueFrom := func(ctx impls.ExecutionContext, expression impls.Expression, row rows.Row) (any, error) {
		val, err := types.ValueAs[bool](expression.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}
		if val == nil {
			return nil, nil
		}
		return !*val, nil
	}

	return newUnaryExpression(expression, "not", typeChecker, valueFrom)
}
