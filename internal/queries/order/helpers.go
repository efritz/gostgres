package order

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/shared"
)

func LowerOrder(order expressions.OrderExpression, nodes ...queries.Node) {
	orderExpressions := order.Expressions()

	for _, node := range nodes {
		filteredExpressions := make([]expressions.ExpressionWithDirection, 0, len(orderExpressions))
	exprLoop:
		for _, expression := range orderExpressions {
			for _, field := range expressions.Fields(expression.Expression) {
				if _, err := shared.FindMatchingFieldIndex(field, node.Fields()); err != nil {
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
