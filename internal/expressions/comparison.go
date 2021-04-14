package expressions

import "github.com/efritz/gostgres/internal/shared"

type equalsExpression struct {
	left  Expression
	right Expression
}

var _ Expression = &equalsExpression{}

func NewEquals(left, right Expression) Expression {
	return &equalsExpression{
		left:  left,
		right: right,
	}
}

func (e equalsExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lVal, err := e.left.ValueFrom(row)
	if err != nil {
		return false, err
	}

	rVal, err := e.right.ValueFrom(row)
	if err != nil {
		return false, err
	}

	return lVal == rVal, nil
}

type lessThanExpression struct {
	left  IntExpression
	right IntExpression
}

var _ Expression = &lessThanExpression{}

func NewLessThan(left, right Expression) Expression {
	return &lessThanExpression{
		left:  Int(left),
		right: Int(right),
	}
}

func (e lessThanExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lVal, err := e.left.ValueFrom(row)
	if err != nil {
		return false, err
	}

	rVal, err := e.right.ValueFrom(row)
	if err != nil {
		return false, err
	}

	return lVal < rVal, nil
}

type lessThanEqualsExpression struct {
	left  IntExpression
	right IntExpression
}

var _ Expression = &lessThanEqualsExpression{}

func NewLessThanEquals(left, right Expression) Expression {
	return &lessThanEqualsExpression{
		left:  Int(left),
		right: Int(right),
	}
}

func (e lessThanEqualsExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lVal, err := e.left.ValueFrom(row)
	if err != nil {
		return false, err
	}

	rVal, err := e.right.ValueFrom(row)
	if err != nil {
		return false, err
	}

	return lVal <= rVal, nil
}

type greaterThanExpression struct {
	left  IntExpression
	right IntExpression
}

var _ Expression = &greaterThanExpression{}

func NewGreaterThan(left, right Expression) Expression {
	return &greaterThanExpression{
		left:  Int(left),
		right: Int(right),
	}
}

func (e greaterThanExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lVal, err := e.left.ValueFrom(row)
	if err != nil {
		return false, err
	}

	rVal, err := e.right.ValueFrom(row)
	if err != nil {
		return false, err
	}

	return lVal > rVal, nil
}

type greaterThanEqualsExpression struct {
	left  IntExpression
	right IntExpression
}

var _ Expression = &greaterThanEqualsExpression{}

func NewGreaterThanEquals(left, right Expression) Expression {
	return &greaterThanEqualsExpression{
		left:  Int(left),
		right: Int(right),
	}
}

func (e greaterThanEqualsExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lVal, err := e.left.ValueFrom(row)
	if err != nil {
		return false, err
	}

	rVal, err := e.right.ValueFrom(row)
	if err != nil {
		return false, err
	}

	return lVal >= rVal, nil
}
