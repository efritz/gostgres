package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

func NewNot(expression Expression) Expression {
	return newUnaryExpression(expression, "not", func(expression Expression, row shared.Row) (interface{}, error) {
		val, err := shared.EnsureNullableBool(expression.ValueFrom(row))
		if err != nil {
			return nil, err
		}
		if val == nil {
			return nil, nil
		}
		return !*val, nil
	})
}

func NewAnd(left, right Expression) Expression {
	return newConditionalExpression(left, right, "and", func(a, b *bool) (interface{}, error) {
		if (a != nil && !*a) || (b != nil && !*b) {
			return false, nil
		}
		if a != nil && b != nil && *a && *b {
			return true, nil
		}
		return nil, nil
	}, simplifyConditional(NewAnd, func(value *bool) (Expression, bool) {
		if value != nil && !*value {
			return NewConstant(false), true
		}

		return nil, false
	}), true)
}

func NewOr(left, right Expression) Expression {
	return newConditionalExpression(left, right, "or", func(a, b *bool) (interface{}, error) {
		if (a != nil && *a) || (b != nil && *b) {
			return true, nil
		}
		if a != nil && b != nil && !*a && !*b {
			return false, nil
		}
		return nil, nil
	}, simplifyConditional(NewOr, func(value *bool) (Expression, bool) {
		if value != nil && *value {
			return NewConstant(true), true
		}

		return nil, false
	}), false)
}

type conditionalExpression struct {
	left         Expression
	right        Expression
	operatorText string
	foldFunc     foldFunc
	valueFrom    conditionalValueFromFunc
	conjunctions bool
}

type foldFunc func(left, right Expression) Expression
type conditionalValueFromFunc func(a, b *bool) (interface{}, error)

func newConditionalExpression(left, right Expression, operatorText string, valueFrom conditionalValueFromFunc, foldFunc foldFunc, conjunctions bool) Expression {
	return conditionalExpression{
		left:         left,
		right:        right,
		operatorText: operatorText,
		valueFrom:    valueFrom,
		foldFunc:     foldFunc,
		conjunctions: conjunctions,
	}
}

func (e conditionalExpression) String() string {
	return fmt.Sprintf("%s %s %s", e.left, e.operatorText, e.right)
}

func (e conditionalExpression) Fields() []shared.Field {
	return append(e.left.Fields(), e.right.Fields()...)
}

func (e conditionalExpression) Fold() Expression {
	return tryEvaluate(e.foldFunc(e.left.Fold(), e.right.Fold()))
}

func (e conditionalExpression) Alias(field shared.Field, expression Expression) Expression {
	return newConditionalExpression(e.left.Alias(field, expression), e.right.Alias(field, expression), e.operatorText, e.valueFrom, e.foldFunc, e.conjunctions)
}

func (e conditionalExpression) Conjunctions() []Expression {
	if !e.conjunctions {
		return []Expression{e}
	}

	return append(e.left.Conjunctions(), e.right.Conjunctions()...)
}

func (e conditionalExpression) ValueFrom(row shared.Row) (interface{}, error) {
	lVal, err := shared.EnsureNullableBool(e.left.ValueFrom(row))
	if err != nil {
		return nil, err
	}

	rVal, err := shared.EnsureNullableBool(e.right.ValueFrom(row))
	if err != nil {
		return nil, err
	}

	return e.valueFrom(lVal, rVal)
}

func simplifyConditional(factory foldFunc, f func(value *bool) (Expression, bool)) func(left, right Expression) Expression {
	return func(left, right Expression) Expression {
		if value, err := shared.EnsureNullableBool(left.ValueFrom(shared.Row{})); err == nil {
			if expression, ok := f(value); ok {
				return expression
			}

			return right
		}

		if value, err := shared.EnsureNullableBool(right.ValueFrom(shared.Row{})); err == nil {
			if expression, ok := f(value); ok {
				return expression
			}

			return left
		}

		return factory(left, right)
	}
}
