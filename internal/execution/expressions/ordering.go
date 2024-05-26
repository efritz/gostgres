package expressions

import (
	"strings"

	"github.com/efritz/gostgres/internal/types"
)

type orderExpression struct {
	expressions []types.ExpressionWithDirection
}

func NewOrderExpression(expressions []types.ExpressionWithDirection) types.OrderExpression {
	return orderExpression{
		expressions: expressions,
	}
}

func (e orderExpression) String() string {
	parts := make([]string, 0, len(e.expressions))
	for _, expression := range e.expressions {
		part := expression.Expression.String()
		if expression.Reverse {
			part += " desc"
		}

		parts = append(parts, part)
	}

	return strings.Join(parts, ", ")
}

func (e orderExpression) Expressions() []types.ExpressionWithDirection {
	return e.expressions
}

func (e orderExpression) Fold() types.OrderExpression {
	expressions := make([]types.ExpressionWithDirection, 0, len(e.expressions))
	for _, expression := range e.expressions {
		expressions = append(expressions, expression.Fold())
	}

	return orderExpression{expressions: expressions}
}

func (e orderExpression) Map(f func(types.Expression) types.Expression) types.OrderExpression {
	expressions := make([]types.ExpressionWithDirection, 0, len(e.expressions))
	for _, expression := range e.expressions {
		expressions = append(expressions, types.ExpressionWithDirection{
			Expression: f(expression.Expression),
			Reverse:    expression.Reverse,
		})
	}

	return orderExpression{expressions: expressions}
}
