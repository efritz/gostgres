package nodes

import (
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

func copyFields(fields []shared.Field) []shared.Field {
	c := make([]shared.Field, len(fields))
	copy(c, fields)
	return c
}

func copyValues(values []interface{}) []interface{} {
	c := make([]interface{}, len(values))
	copy(c, values)
	return c
}

func updateRelationName(fields []shared.Field, relationName string) []shared.Field {
	fields = copyFields(fields)
	for i := range fields {
		fields[i].RelationName = relationName
	}

	return fields
}

func indent(level int) string {
	return strings.Repeat(" ", level*4)
}

func unionFilters(filters ...expressions.Expression) expressions.Expression {
	var conjunctions []expressions.Expression
	for _, expression := range filters {
		if expression == nil {
			continue
		}

		conjunctions = append(conjunctions, expression.Conjunctions()...)
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

		filter = expressions.NewAnd(filter, conjunction)
	}

	return filter
}

func filterIntersection(filter, childFilter expressions.Expression) expressions.Expression {
	return combineFilters(filter, childFilter, func(conjunctions, childConjunctions []expressions.Expression) {
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

func filterDifference(filter, childFilter expressions.Expression) expressions.Expression {
	return combineFilters(filter, childFilter, func(conjunctions, childConjunctions []expressions.Expression) {
	outer:
		for i, f1 := range conjunctions {
			for _, f2 := range childConjunctions {
				if f1.Equal(f2) {
					conjunctions[i] = nil
					continue outer
				}
			}
		}
	})
}

func combineFilters(filter, childFilter expressions.Expression, filterConjunctions func(conjunctions, childConjunctions []expressions.Expression)) expressions.Expression {
	if filter == nil {
		return nil
	}
	if childFilter == nil {
		return filter
	}

	conjunctions := filter.Conjunctions()
	childConjunctions := childFilter.Conjunctions()

	filterConjunctions(conjunctions, childConjunctions)

	temp := conjunctions
	conjunctions = conjunctions[:]
	for _, v := range temp {
		if v == nil {
			continue
		}

		conjunctions = append(conjunctions, v)
	}

	return unionFilters(conjunctions...)
}

func lowerFilter(filter expressions.Expression, nodes ...Node) {
	for _, expression := range filter.Conjunctions() {
		missing := make([]bool, len(nodes))
		for _, field := range expression.Fields() {
			for i, node := range nodes {
				if _, err := shared.FindMatchingFieldIndex(field, node.Fields()); err != nil {
					missing[i] = true
				}
			}
		}

		for i, missing := range missing {
			if !missing {
				nodes[i].AddFilter(expression)
			}
		}
	}
}

func lowerOrder(order OrderExpression, nodes ...Node) {
	expressions := order.Expressions()

	for _, node := range nodes {
		filteredExpressions := make([]FieldExpression, 0, len(expressions))
	exprLoop:
		for _, expression := range expressions {
			for _, field := range expression.Expression.Fields() {
				if _, err := shared.FindMatchingFieldIndex(field, node.Fields()); err != nil {
					continue exprLoop
				}
			}

			filteredExpressions = append(filteredExpressions, expression)
		}

		if len(filteredExpressions) != 0 {
			node.AddOrder(NewOrderExpression(filteredExpressions))
		}
	}
}

func mapOrderExpressions(order OrderExpression, f func(expressions.Expression) expressions.Expression) OrderExpression {
	fieldExpressions := order.Expressions()
	aliasedExpressions := make([]FieldExpression, 0, len(fieldExpressions))

	for _, fieldExpression := range fieldExpressions {
		aliasedExpressions = append(aliasedExpressions, FieldExpression{
			Expression: f(fieldExpression.Expression),
			Reverse:    fieldExpression.Reverse,
		})
	}

	return NewOrderExpression(aliasedExpressions)
}
