package expressions

import (
	"github.com/efritz/gostgres/internal/shared"
)

type equalsExpression struct {
	left  Expression
	right Expression
}

var _ Expression = &equalsExpression{}

func NewEquals(left, right Expression) Expression {
	return &equalsExpression{
		left:  left,
		right: right,
	}
}

func (e equalsExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lVal, err := e.left.ValueFrom(row)
	if err != nil {
		return false, err
	}

	rVal, err := e.right.ValueFrom(row)
	if err != nil {
		return false, err
	}

	return lVal == rVal, nil
}
