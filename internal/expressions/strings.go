package expressions

import "github.com/efritz/gostgres/internal/shared"

func NewConcat(left, right Expression) Expression {
	return newBinaryExpression(left, right, "||", func(left, right Expression, row shared.Row) (any, error) {
		lVal, err := shared.ValueAs[string](left.ValueFrom(row))
		if err != nil {
			return nil, err
		}

		rVal, err := shared.ValueAs[string](right.ValueFrom(row))
		if err != nil {
			return nil, err
		}

		if lVal == nil || rVal == nil {
			return nil, nil
		}

		return *lVal + *rVal, nil
	})
}

func NewLike(left, right Expression) Expression {
	panic("NewLike unimplemented") // TODO
}

func NewILike(left, right Expression) Expression {
	panic("NewILike unimplemented") // TODO
}
