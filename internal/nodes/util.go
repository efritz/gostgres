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

func combineConjunctions(conjunctions []expressions.Expression) expressions.Expression {
	if len(conjunctions) == 0 {
		return nil
	}

	expression := conjunctions[0]
	for _, conjunction := range conjunctions[1:] {
		expression = expressions.NewAnd(expression, conjunction)
	}

	return expression
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
