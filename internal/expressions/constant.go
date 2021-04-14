package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type constantExpression struct {
	value interface{}
}

var _ Expression = &constantExpression{}

func NewConstant(value interface{}) Expression {
	return &constantExpression{
		value: value,
	}
}

func (e constantExpression) String() string {
	return fmt.Sprintf("%v", e.value)
}

func (e constantExpression) ValueFrom(row shared.Row) (interface{}, error) {
	return e.value, nil
}
