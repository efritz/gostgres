package expressions

import (
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

func NewConcat(left, right types.Expression) types.Expression {
	return newBinaryExpression(left, right, "||", func(ctx types.Context, left, right types.Expression, row shared.Row) (any, error) {
		lVal, err := shared.ValueAs[string](left.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}

		rVal, err := shared.ValueAs[string](right.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}

		if lVal == nil || rVal == nil {
			return nil, nil
		}

		return *lVal + *rVal, nil
	})
}

func NewLike(left, right types.Expression) types.Expression {
	panic("NewLike unimplemented") // TODO
}

func NewILike(left, right types.Expression) types.Expression {
	panic("NewILike unimplemented") // TODO
}
