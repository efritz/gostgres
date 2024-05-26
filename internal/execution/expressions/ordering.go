package expressions

import (
	"strings"

	"github.com/efritz/gostgres/internal/shared/impls"
)

type orderExpression struct {
	expressions []impls.ExpressionWithDirection
}

func NewOrderExpression(expressions []impls.ExpressionWithDirection) impls.OrderExpression {
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

func (e orderExpression) Expressions() []impls.ExpressionWithDirection {
	return e.expressions
}

func (e orderExpression) Fold() impls.OrderExpression {
	expressions := make([]impls.ExpressionWithDirection, 0, len(e.expressions))
	for _, expression := range e.expressions {
		expressions = append(expressions, expression.Fold())
	}

	return orderExpression{expressions: expressions}
}

func (e orderExpression) Map(f func(impls.Expression) impls.Expression) impls.OrderExpression {
	expressions := make([]impls.ExpressionWithDirection, 0, len(e.expressions))
	for _, expression := range e.expressions {
		expressions = append(expressions, impls.ExpressionWithDirection{
			Expression: f(expression.Expression),
			Reverse:    expression.Reverse,
		})
	}

	return orderExpression{expressions: expressions}
}
