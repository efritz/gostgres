package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type sumExpression struct {
	left  Expression
	right Expression
}

var _ Expression = &sumExpression{}

func NewSum(left, right Expression) Expression {
	return &sumExpression{
		left:  left,
		right: right,
	}
}

func (e sumExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lRaw, err := e.left.ValueFrom(row)
	if err != nil {
		return nil, err
	}
	lVal, ok := lRaw.(int)
	if !ok {
		return nil, fmt.Errorf("not a number")
	}

	rRaw, err := e.right.ValueFrom(row)
	if err != nil {
		return nil, err
	}
	rVal, ok := rRaw.(int)
	if !ok {
		return nil, fmt.Errorf("not a number")
	}

	return lVal + rVal, nil
}
