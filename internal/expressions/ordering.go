package expressions

import (
	"strings"
)

type OrderExpression interface {
	Expressions() []ExpressionWithDirection
	Fold() OrderExpression
	Map(f func(e Expression) Expression) OrderExpression
}

type ExpressionWithDirection struct {
	Expression Expression
	Reverse    bool
}

func (e ExpressionWithDirection) Fold() ExpressionWithDirection {
	return ExpressionWithDirection{
		Expression: e.Expression.Fold(),
		Reverse:    e.Reverse,
	}
}

type orderExpression struct {
	expressions []ExpressionWithDirection
}

func NewOrderExpression(expressions []ExpressionWithDirection) OrderExpression {
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

func (e orderExpression) Expressions() []ExpressionWithDirection {
	return e.expressions
}

func (e orderExpression) Fold() OrderExpression {
	expressions := make([]ExpressionWithDirection, 0, len(e.expressions))
	for _, expression := range e.expressions {
		expressions = append(expressions, expression.Fold())
	}

	return orderExpression{expressions: expressions}
}

func (e orderExpression) Map(f func(Expression) Expression) OrderExpression {
	expressions := make([]ExpressionWithDirection, 0, len(e.expressions))
	for _, expression := range e.expressions {
		expressions = append(expressions, ExpressionWithDirection{
			Expression: f(expression.Expression),
			Reverse:    expression.Reverse,
		})
	}

	return orderExpression{expressions: expressions}
}
