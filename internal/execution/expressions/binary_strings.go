package expressions

import (
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

func NewConcat(left, right impls.Expression) impls.Expression {
	return newBinaryExpression(left, right, "||", func(ctx impls.Context, left, right impls.Expression, row rows.Row) (any, error) {
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
	})
}

func NewLike(left, right impls.Expression) impls.Expression {
	panic("NewLike unimplemented") // TODO
}

func NewILike(left, right impls.Expression) impls.Expression {
	panic("NewILike unimplemented") // TODO
}
