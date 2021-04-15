package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

func NewNot(expression Expression) Expression {
	return newUnaryExpression(expression, "not", func(expression Expression, row shared.Row) (interface{}, error) {
		val, err := EnsureBool(expression.ValueFrom(row))
		if err != nil {
			return nil, err
		}

		return !val, nil
	})
}

func NewAnd(left, right Expression) Expression {
	binaryExpression := newConditionalExpression(left, right, "and", func(a, b bool) (interface{}, error) {
		return a && b, nil
	})

	return andExpression{conditionalExpression{binaryExpression, simplifyAnd}}
}

func NewOr(left, right Expression) Expression {
	binaryExpression := newConditionalExpression(left, right, "or", func(a, b bool) (interface{}, error) {
		return a || b, nil
	})

	return conditionalExpression{binaryExpression, simplifyOr}
}

type andExpression struct {
	conditionalExpression
}

func (e andExpression) Fold() Expression {
	folded := e.conditionalExpression.Fold()
	if conditionalExpression, ok := folded.(conditionalExpression); ok {
		return andExpression{conditionalExpression}
	}

	return folded
}

func (e andExpression) Alias(from, to string) Expression {
	aliased := e.binaryExpression.Alias(from, to)

	if bin, ok := aliased.(binaryExpression); ok {
		return conditionalExpression{bin, e.foldFunc}
	}

	return aliased
}

func (e andExpression) Conjunctions() []Expression {
	return append(
		e.conditionalExpression.left.Conjunctions(),
		e.conditionalExpression.right.Conjunctions()...,
	)
}

type conditionalExpression struct {
	binaryExpression
	foldFunc foldFunc
}

type boolFunc func(a, b bool) (interface{}, error)

func (e conditionalExpression) String() string {
	return fmt.Sprintf("%s", e.binaryExpression)
}

func (e conditionalExpression) Fold() Expression {
	folded := e.binaryExpression.Fold()

	if bin, ok := folded.(binaryExpression); ok {
		return e.foldFunc(bin.left, bin.right)
	}

	return folded
}

func (e conditionalExpression) Alias(from, to string) Expression {
	aliased := e.binaryExpression.Alias(from, to)

	if bin, ok := aliased.(binaryExpression); ok {
		return conditionalExpression{bin, e.foldFunc}
	}

	return aliased
}

func newConditionalExpression(left, right Expression, operatorText string, f boolFunc) binaryExpression {
	return newBinaryExpression(left, right, operatorText, func(left, right Expression, row shared.Row) (interface{}, error) {
		lVal, err := EnsureBool(left.ValueFrom(row))
		if err != nil {
			return nil, err
		}

		rVal, err := EnsureBool(right.ValueFrom(row))
		if err != nil {
			return nil, err
		}

		return f(lVal, rVal)
	})
}

func simplifyAnd(left, right Expression) Expression {
	return simplifyConditional(left, right, NewAnd, func(value bool) (Expression, bool) {
		if value {
			return nil, false
		}

		return NewConstant(false), true
	})
}

func simplifyOr(left, right Expression) Expression {
	return simplifyConditional(left, right, NewOr, func(value bool) (Expression, bool) {
		if !value {
			return nil, false
		}

		return NewConstant(true), true
	})
}

type foldFunc func(left, right Expression) Expression

func simplifyConditional(left, right Expression, factory foldFunc, f func(value bool) (Expression, bool)) Expression {
	if value, err := EnsureBool(left.ValueFrom(shared.Row{})); err == nil {
		if expression, ok := f(value); ok {
			return expression
		}

		return right
	}

	if value, err := EnsureBool(right.ValueFrom(shared.Row{})); err == nil {
		if expression, ok := f(value); ok {
			return expression
		}

		return left
	}

	return factory(left, right)
}
