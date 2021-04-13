package expressions

import "github.com/efritz/gostgres/internal/shared"

type constantExpression struct {
	value interface{}
}

var _ Expression = &constantExpression{}

func NewConstant(value interface{}) Expression {
	return &constantExpression{
		value: value,
	}
}

func (e constantExpression) ValueFrom(row shared.Row) (interface{}, error) {
	return e.value, nil
}
