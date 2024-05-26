package order

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

func LowerOrder(order impls.OrderExpression, nodes ...queries.Node) {
	orderExpressions := order.Expressions()

	for _, node := range nodes {
		filteredExpressions := make([]impls.ExpressionWithDirection, 0, len(orderExpressions))
	exprLoop:
		for _, expression := range orderExpressions {
			for _, field := range expressions.Fields(expression.Expression) {
				if _, err := fields.FindMatchingFieldIndex(field, node.Fields()); err != nil {
					continue exprLoop
				}
			}

			filteredExpressions = append(filteredExpressions, expression)
		}

		if len(filteredExpressions) != 0 {
			node.AddOrder(expressions.NewOrderExpression(filteredExpressions))
		}
	}
}
