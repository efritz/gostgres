package opt

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

func lowerOrder(ctx impls.OptimizationContext, order impls.OrderExpression, nodes ...LogicalNode) {
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
			node.AddOrder(ctx, expressions.NewOrderExpression(filteredExpressions))
		}
	}
}
