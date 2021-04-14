package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type notExpression struct {
	expression Expression
}

var _ Expression = &andExpression{}

func NewNot(expression Expression) Expression {
	return &notExpression{
		expression: expression,
	}
}

func (e notExpression) String() string {
	return fmt.Sprintf("not %s", e.expression)
}

func (e notExpression) ValueFrom(row shared.Row) (interface{}, error) {
	val, err := Bool(e.expression).ValueFrom(row)
	if err != nil {
		return false, err
	}

	return !val, nil
}

type andExpression struct {
	left  Expression
	right Expression
}

var _ Expression = &andExpression{}

func NewAnd(left, right Expression) Expression {
	return &andExpression{
		left:  left,
		right: right,
	}
}

func (e andExpression) String() string {
	return fmt.Sprintf("%s and %s", e.left, e.right)
}

func (e andExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lVal, err := Bool(e.left).ValueFrom(row)
	if err != nil {
		return false, err
	}

	rVal, err := Bool(e.right).ValueFrom(row)
	if err != nil {
		return false, err
	}

	// TODO - short-circuit
	return lVal && rVal, nil
}

type orExpression struct {
	left  Expression
	right Expression
}

var _ Expression = &orExpression{}

func NewOr(left, right Expression) Expression {
	return &orExpression{
		left:  left,
		right: right,
	}
}

func (e orExpression) String() string {
	return fmt.Sprintf("%s or %s", e.left, e.right)
}

func (e orExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lVal, err := Bool(e.left).ValueFrom(row)
	if err != nil {
		return false, err
	}

	rVal, err := Bool(e.right).ValueFrom(row)
	if err != nil {
		return false, err
	}

	// TODO - short-circuit
	return lVal || rVal, nil
}
