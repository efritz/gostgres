package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type IntExpression interface {
	ValueFrom(row shared.Row) (int, error)
}

type intExpression struct{ Expression }

func Int(expression Expression) IntExpression {
	return intExpression{expression}
}

func (e intExpression) String() string {
	return fmt.Sprintf("%s", e.Expression)
}

func (e intExpression) ValueFrom(row shared.Row) (int, error) {
	val, err := e.Expression.ValueFrom(row)
	if err != nil {
		return 0, err
	}

	typedVal, ok := val.(int)
	if !ok {
		return 0, fmt.Errorf("unexpected type (wanted int, have %v)", val)
	}
	return typedVal, nil
}

type BoolExpression interface {
	ValueFrom(row shared.Row) (bool, error)
}

type boolExpression struct{ Expression }

func Bool(expression Expression) BoolExpression {
	return boolExpression{expression}
}

func (e boolExpression) String() string {
	return fmt.Sprintf("%s", e.Expression)
}

func (e boolExpression) ValueFrom(row shared.Row) (bool, error) {
	val, err := e.Expression.ValueFrom(row)
	if err != nil {
		return false, err
	}

	typedVal, ok := val.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected type (wanted bool, have %v)", val)
	}
	return typedVal, nil
}
