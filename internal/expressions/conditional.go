package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

func NewNot(expression Expression) Expression {
	return newUnaryExpression(expression, "not", func(context ExpressionContext, expression Expression, row shared.Row) (any, error) {
		val, err := shared.ValueAs[bool](expression.ValueFrom(context, row))
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
			return compareExpressionBags(e.Conjunctions(), o.Conjunctions())
		}

		if !e.conjunctions && !o.conjunctions {
			return compareExpressionBags(e.disjunctions(), o.disjunctions())
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

func (e conditionalExpression) Fields() []shared.Field {
	return append(e.left.Fields(), e.right.Fields()...)
}

func (e conditionalExpression) Named() (shared.Field, bool) {
	return shared.Field{}, false
}

func (e conditionalExpression) Fold() Expression {
	return tryEvaluate(e.foldFunc(e.left.Fold(), e.right.Fold()))
}

func (e conditionalExpression) Map(f func(Expression) Expression) Expression {
	return f(newConditionalExpression(e.left.Map(f), e.right.Map(f), e.operatorText, e.valueFrom, e.foldFunc, e.conjunctions))
}

func (e conditionalExpression) Conjunctions() []Expression {
	if !e.conjunctions {
		return []Expression{e}
	}

	return append(e.left.Conjunctions(), e.right.Conjunctions()...)
}

func (e conditionalExpression) disjunctions() (disjunctions []Expression) {
	if e.conjunctions {
		return []Expression{e}
	}

	if l, ok := e.left.(conditionalExpression); ok {
		disjunctions = append(disjunctions, l.disjunctions()...)
	} else {
		disjunctions = append(disjunctions, e.left)
	}

	if r, ok := e.right.(conditionalExpression); ok {
		disjunctions = append(disjunctions, r.disjunctions()...)
	} else {
		disjunctions = append(disjunctions, e.right)
	}

	return disjunctions
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
