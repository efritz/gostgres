package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type Expression interface {
	ValueFrom(row shared.Row) (interface{}, error)
}

type valueFromFunc func(row shared.Row) (interface{}, error)

type unaryExpression struct {
	expression   Expression
	operatorText string
	valueFrom    valueFromFunc
}

func newUnaryExpression(expression Expression, operatorText string, valueFrom valueFromFunc) Expression {
	return &unaryExpression{
		expression:   expression,
		operatorText: operatorText,
		valueFrom:    valueFrom,
	}
}

func (e unaryExpression) String() string {
	return fmt.Sprintf("%s %s", e.operatorText, e.expression)
}

func (e unaryExpression) ValueFrom(row shared.Row) (interface{}, error) {
	return e.valueFrom(row)
}

type binaryExpression struct {
	left         Expression
	right        Expression
	operatorText string
	valueFrom    valueFromFunc
}

func newBinaryExpression(left, right Expression, operatorText string, valueFrom valueFromFunc) Expression {
	return &binaryExpression{
		left:         left,
		right:        right,
		operatorText: operatorText,
		valueFrom:    valueFrom,
	}
}

func (e binaryExpression) String() string {
	return fmt.Sprintf("%s %s %s", e.left, e.operatorText, e.right)
}

func (e binaryExpression) ValueFrom(row shared.Row) (interface{}, error) {
	return e.valueFrom(row)
}
