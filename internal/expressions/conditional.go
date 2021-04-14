package expressions

import "github.com/efritz/gostgres/internal/shared"

type notExpression struct {
	expression BoolExpression
}

var _ Expression = &andExpression{}

func NewNot(expression Expression) Expression {
	return &notExpression{
		expression: Bool(expression),
	}
}

func (e notExpression) ValueFrom(row shared.Row) (interface{}, error) {
	val, err := e.expression.ValueFrom(row)
	if err != nil {
		return false, err
	}

	return !val, nil
}

type andExpression struct {
	left  BoolExpression
	right BoolExpression
}

var _ Expression = &andExpression{}

func NewAnd(left, right Expression) Expression {
	return &andExpression{
		left:  Bool(left),
		right: Bool(right),
	}
}

func (e andExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lVal, err := e.left.ValueFrom(row)
	if err != nil {
		return false, err
	}

	rVal, err := e.right.ValueFrom(row)
	if err != nil {
		return false, err
	}

	// TODO - short-circuit
	return lVal && rVal, nil
}

type orExpression struct {
	left  BoolExpression
	right BoolExpression
}

var _ Expression = &orExpression{}

func NewOr(left, right Expression) Expression {
	return &orExpression{
		left:  Bool(left),
		right: Bool(right),
	}
}

func (e orExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lVal, err := e.left.ValueFrom(row)
	if err != nil {
		return false, err
	}

	rVal, err := e.right.ValueFrom(row)
	if err != nil {
		return false, err
	}

	// TODO - short-circuit
	return lVal || rVal, nil
}
