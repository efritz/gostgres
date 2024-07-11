package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

func NewAnd(left, right impls.Expression) impls.Expression {
	return newConditionalExpression(left, right, "and", func(a, b *bool) (any, error) {
		if (a != nil && !*a) || (b != nil && !*b) {
			return false, nil
		}
		if a != nil && b != nil && *a && *b {
			return true, nil
		}
		return nil, nil
	}, simplifyConditional(NewAnd, func(value *bool) (impls.Expression, bool) {
		if value != nil && !*value {
			return NewConstant(false), true
		}

		return nil, false
	}), true)
}

func NewOr(left, right impls.Expression) impls.Expression {
	return newConditionalExpression(left, right, "or", func(a, b *bool) (any, error) {
		if (a != nil && *a) || (b != nil && *b) {
			return true, nil
		}
		if a != nil && b != nil && !*a && !*b {
			return false, nil
		}
		return nil, nil
	}, simplifyConditional(NewOr, func(value *bool) (impls.Expression, bool) {
		if value != nil && *value {
			return NewConstant(true), true
		}

		return nil, false
	}), false)
}

type conditionalExpression struct {
	left         impls.Expression
	right        impls.Expression
	operatorText string
	foldFunc     foldFunc
	valueFrom    conditionalValueFromFunc
	conjunctions bool
}

type foldFunc func(left, right impls.Expression) impls.Expression
type conditionalValueFromFunc func(a, b *bool) (any, error)

func newConditionalExpression(left, right impls.Expression, operatorText string, valueFrom conditionalValueFromFunc, foldFunc foldFunc, conjunctions bool) impls.Expression {
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

func (e conditionalExpression) Equal(other impls.Expression) bool {
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

func compareExpressionBags(as, bs []impls.Expression) bool {
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

func (e conditionalExpression) Children() []impls.Expression {
	return []impls.Expression{e.left, e.right}
}

func (e conditionalExpression) Fold() impls.Expression {
	return tryEvaluate(e.foldFunc(e.left.Fold(), e.right.Fold()))
}

func (e conditionalExpression) Map(f func(impls.Expression) (impls.Expression, error)) (impls.Expression, error) {
	left, err := e.left.Map(f)
	if err != nil {
		return nil, err
	}

	right, err := e.right.Map(f)
	if err != nil {
		return nil, err
	}

	return f(newConditionalExpression(left, right, e.operatorText, e.valueFrom, e.foldFunc, e.conjunctions))
}

func (e conditionalExpression) ValueFrom(ctx impls.Context, row rows.Row) (any, error) {
	lVal, err := types.ValueAs[bool](e.left.ValueFrom(ctx, row))
	if err != nil {
		return nil, err
	}

	rVal, err := types.ValueAs[bool](e.right.ValueFrom(ctx, row))
	if err != nil {
		return nil, err
	}

	return e.valueFrom(lVal, rVal)
}

func simplifyConditional(factory foldFunc, f func(value *bool) (impls.Expression, bool)) foldFunc {
	return func(left, right impls.Expression) impls.Expression {
		if value, err := types.ValueAs[bool](left.ValueFrom(impls.EmptyContext, rows.Row{})); err == nil {
			if expression, ok := f(value); ok {
				return expression
			}

			return right
		}

		if value, err := types.ValueAs[bool](right.ValueFrom(impls.EmptyContext, rows.Row{})); err == nil {
			if expression, ok := f(value); ok {
				return expression
			}

			return left
		}

		return factory(left, right)
	}
}
