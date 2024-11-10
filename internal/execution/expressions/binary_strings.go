package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

func NewConcat(left, right impls.Expression) impls.Expression {
	typeChecker := func(left types.Type, right types.Type) (types.Type, error) {
		if left == types.TypeText && right == types.TypeText {
			return types.TypeText, nil
		}

		return types.TypeUnknown, fmt.Errorf("illegal operand types for concatenation: %s and %s", left, right)
	}

	valueFrom := func(ctx impls.ExecutionContext, left, right impls.Expression, row rows.Row) (any, error) {
		lVal, err := types.ValueAs[string](left.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}

		rVal, err := types.ValueAs[string](right.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}

		if lVal == nil || rVal == nil {
			return nil, nil
		}

		return *lVal + *rVal, nil
	}

	return newBinaryExpression(left, right, "||", typeChecker, valueFrom)
}

func NewLike(left, right impls.Expression) impls.Expression {
	panic("NewLike unimplemented") // TODO
}

func NewILike(left, right impls.Expression) impls.Expression {
	panic("NewILike unimplemented") // TODO
}
