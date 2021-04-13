package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type lengthExpression struct {
	expression Expression
}

var _ Expression = &lengthExpression{}

func NewLength(expression Expression) Expression {
	return &lengthExpression{
		expression: expression,
	}
}

func (e lengthExpression) ValueFrom(row shared.Row) (interface{}, error) {
	raw, err := e.expression.ValueFrom(row)
	if err != nil {
		return nil, err
	}
	val, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("not a string")
	}

	return len(val), nil
}
