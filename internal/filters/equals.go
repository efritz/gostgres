package filters

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type equalsFilter struct {
	left  expressions.Expression
	right expressions.Expression
}

var _ Filter = &equalsFilter{}

func NewEquals(left, right expressions.Expression) Filter {
	return &equalsFilter{
		left:  left,
		right: right,
	}
}

func (f equalsFilter) Test(row shared.Row) (bool, error) {
	lVal, err := f.left.ValueFrom(row)
	if err != nil {
		return false, err
	}

	rVal, err := f.right.ValueFrom(row)
	if err != nil {
		return false, err
	}

	return lVal == rVal, nil
}
