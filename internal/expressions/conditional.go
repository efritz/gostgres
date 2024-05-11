package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

func NewAnd(left, right Expression) Expression {
	return newConditionalExpression(left, right, "and", func(a, b *bool) (any, error) {
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
	return newConditionalExpression(left, right, "or", func(a, b *bool) (any, error) {
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
type conditionalValueFromFunc func(a, b *bool) (any, error)

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

func (e conditionalExpression) Equal(other Expression) bool {
	if o, ok := other.(conditionalExpression); ok {
		if e.conjunctions && o.conjunctions {
			return compareExpressionBags(Conjunctions(e), Conjunctions(o))
		}

		if !e.conjunctions && !o.conjunctions {
			return compareExpressionBags(Disjunctions(e), Disjunctions(o))
		}
	}

	return false
}

func compareExpressionBags(as, bs []Expression) bool {
outer:
	for _, a := range as {
		for i, b := range bs {
			if a.Equal(b) {
				// Remove element i from bs
				n := len(bs) - 1
				bs[i] = bs[n]
				bs = bs[:n]

				continue outer
			}
		}

		return false
	}

	return len(bs) == 0
}

func (e conditionalExpression) Children() []Expression {
	return []Expression{e.left, e.right}
}

func (e conditionalExpression) Fold() Expression {
	return tryEvaluate(e.foldFunc(e.left.Fold(), e.right.Fold()))
}

func (e conditionalExpression) Map(f func(Expression) Expression) Expression {
	return f(newConditionalExpression(e.left.Map(f), e.right.Map(f), e.operatorText, e.valueFrom, e.foldFunc, e.conjunctions))
}

func (e conditionalExpression) ValueFrom(context ExpressionContext, row shared.Row) (any, error) {
	lVal, err := shared.ValueAs[bool](e.left.ValueFrom(context, row))
	if err != nil {
		return nil, err
	}

	rVal, err := shared.ValueAs[bool](e.right.ValueFrom(context, row))
	if err != nil {
		return nil, err
	}

	return e.valueFrom(lVal, rVal)
}

func simplifyConditional(factory foldFunc, f func(value *bool) (Expression, bool)) foldFunc {
	return func(left, right Expression) Expression {
		if value, err := shared.ValueAs[bool](left.ValueFrom(EmptyContext, shared.Row{})); err == nil {
			if expression, ok := f(value); ok {
				return expression
			}

			return right
		}

		if value, err := shared.ValueAs[bool](right.ValueFrom(EmptyContext, shared.Row{})); err == nil {
			if expression, ok := f(value); ok {
				return expression
			}

			return left
		}

		return factory(left, right)
	}
}
