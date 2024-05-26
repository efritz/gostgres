package expressions

import (
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
)

func Conjunctions(e impls.Expression) []impls.Expression {
	if c, ok := e.(conditionalExpression); ok && c.conjunctions {
		return append(Conjunctions(c.left), Conjunctions(c.right)...)
	}

	return []impls.Expression{e}
}

func Disjunctions(e impls.Expression) []impls.Expression {
	if c, ok := e.(conditionalExpression); ok && !c.conjunctions {
		return append(Disjunctions(c.left), Disjunctions(c.right)...)
	}

	return []impls.Expression{e}
}

func FilterDifference(filter, childFilter impls.Expression) impls.Expression {
	return combineFilters(filter, childFilter, func(conjunctions, childConjunctions []impls.Expression) {
		for i, f1 := range conjunctions {
			for _, f2 := range childConjunctions {
				if f1.Equal(f2) {
					conjunctions[i] = nil
					break
				}
			}
		}
	})
}

func FilterIntersection(filter, childFilter impls.Expression) impls.Expression {
	return combineFilters(filter, childFilter, func(conjunctions, childConjunctions []impls.Expression) {
	outer:
		for i, f1 := range conjunctions {
			for _, f2 := range childConjunctions {
				if f1.Equal(f2) {
					continue outer
				}
			}

			conjunctions[i] = nil
		}
	})
}

func combineFilters(filter, childFilter impls.Expression, filterConjunctions func(conjunctions, childConjunctions []impls.Expression)) impls.Expression {
	if filter == nil {
		return nil
	}
	if childFilter == nil {
		return filter
	}

	conjunctions := Conjunctions(filter)
	filterConjunctions(conjunctions, Conjunctions(childFilter))
	return UnionFilters(conjunctions...)
}

func UnionFilters(filters ...impls.Expression) impls.Expression {
	var conjunctions []impls.Expression
	for _, expression := range filters {
		if expression == nil {
			continue
		}

		conjunctions = append(conjunctions, Conjunctions(expression)...)
	}
	if len(conjunctions) == 0 {
		return nil
	}

	for i, c1 := range conjunctions {
		for j, c2 := range conjunctions {
			if c1 == nil || c2 == nil || j <= i {
				continue
			}

			if c1.Equal(c2) {
				conjunctions[j] = nil
			}
		}
	}

	filter := conjunctions[0]
	for _, conjunction := range conjunctions[1:] {
		if conjunction == nil {
			continue
		}

		filter = NewAnd(filter, conjunction)
	}

	return filter
}

func SubsumesOrder(a, b impls.OrderExpression) bool {
	if a == nil || b == nil {
		return false
	}

	aExpressions := a.Expressions()
	bExpressions := b.Expressions()
	if len(bExpressions) < len(aExpressions) {
		return false
	}

	for i, expression := range aExpressions {
		if expression.Reverse != bExpressions[i].Reverse {
			return false
		}

		if !expression.Expression.Equal(bExpressions[i].Expression) {
			return false
		}
	}

	return true
}

func tryEvaluate(expression impls.Expression) impls.Expression {
	if value, err := expression.ValueFrom(impls.EmptyContext, rows.Row{}); err == nil {
		return NewConstant(value)
	}

	return expression
}
