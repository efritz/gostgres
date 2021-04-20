package expressions

import "github.com/efritz/gostgres/internal/shared"

func NewConcat(left, right Expression) Expression {
	return newBinaryExpression(left, right, "||", func(left, right Expression, row shared.Row) (interface{}, error) {
		lVal, err := shared.EnsureString(left.ValueFrom(row))
		if err != nil {
			return nil, err
		}

		rVal, err := shared.EnsureString(right.ValueFrom(row))
		if err != nil {
			return nil, err
		}

		return lVal + rVal, nil
	})
}

func NewLike(left, right Expression) Expression {
	panic("NewLike unimplemented") // TODO
}

func NewILike(left, right Expression) Expression {
	panic("NewILike unimplemented") // TODO
}
