package expressions

import "github.com/efritz/gostgres/internal/shared"

type additionExpression struct {
	left  IntExpression
	right IntExpression
}

var _ Expression = &additionExpression{}

func NewAddition(left, right Expression) Expression {
	return &additionExpression{
		left:  Int(left),
		right: Int(right),
	}
}

func (e additionExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lVal, err := e.left.ValueFrom(row)
	if err != nil {
		return nil, err
	}

	rVal, err := e.right.ValueFrom(row)
	if err != nil {
		return nil, err
	}

	return lVal + rVal, nil
}

type subtractionExpression struct {
	left  IntExpression
	right IntExpression
}

var _ Expression = &subtractionExpression{}

func NewSubtraction(left, right Expression) Expression {
	return &subtractionExpression{
		left:  Int(left),
		right: Int(right),
	}
}

func (e subtractionExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lVal, err := e.left.ValueFrom(row)
	if err != nil {
		return nil, err
	}

	rVal, err := e.right.ValueFrom(row)
	if err != nil {
		return nil, err
	}

	return lVal - rVal, nil
}

type multiplicationExpression struct {
	left  IntExpression
	right IntExpression
}

var _ Expression = &multiplicationExpression{}

func NewMultiplication(left, right Expression) Expression {
	return &multiplicationExpression{
		left:  Int(left),
		right: Int(right),
	}
}

func (e multiplicationExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lVal, err := e.left.ValueFrom(row)
	if err != nil {
		return nil, err
	}

	rVal, err := e.right.ValueFrom(row)
	if err != nil {
		return nil, err
	}

	return lVal * rVal, nil
}

type divisionExpression struct {
	left  IntExpression
	right IntExpression
}

var _ Expression = &divisionExpression{}

func NewDivision(left, right Expression) Expression {
	return &divisionExpression{
		left:  Int(left),
		right: Int(right),
	}
}

func (e divisionExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lVal, err := e.left.ValueFrom(row)
	if err != nil {
		return nil, err
	}

	rVal, err := e.right.ValueFrom(row)
	if err != nil {
		return nil, err
	}

	// TODO - div by zero
	return lVal / rVal, nil
}

func NewUnaryPlus(expression Expression) Expression {
	return NewAddition(NewConstant(0), expression)
}

func NewUnaryMinus(expression Expression) Expression {
	return NewSubtraction(NewConstant(0), expression)
}
