package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type intExpressionFunc func(row shared.Row) (int, error)
type boolExpressionFunc func(row shared.Row) (bool, error)

func (f intExpressionFunc) ValueFrom(row shared.Row) (int, error)   { return f(row) }
func (f boolExpressionFunc) ValueFrom(row shared.Row) (bool, error) { return f(row) }

type IntExpression interface {
	ValueFrom(row shared.Row) (int, error)
}

type BoolExpression interface {
	ValueFrom(row shared.Row) (bool, error)
}

func Int(expression Expression) IntExpression {
	return intExpressionFunc(typedExpression{expression}.valueFromInt)
}

func Bool(expression Expression) BoolExpression {
	return boolExpressionFunc(typedExpression{expression}.valueFromBool)
}

type typedExpression struct {
	expression Expression
}

func (f typedExpression) valueFromInt(row shared.Row) (int, error) {
	val, err := f.expression.ValueFrom(row)
	if err != nil {
		return 0, err
	}

	typedVal, ok := val.(int)
	if !ok {
		return 0, fmt.Errorf("unexpected type")
	}
	return typedVal, nil
}

func (f typedExpression) valueFromBool(row shared.Row) (bool, error) {
	val, err := f.expression.ValueFrom(row)
	if err != nil {
		return false, err
	}

	typedVal, ok := val.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected type")
	}
	return typedVal, nil
}
