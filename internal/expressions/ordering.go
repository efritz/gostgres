package expressions

import (
	"fmt"
	"strings"
)

type OrderExpression interface {
	Fold() OrderExpression
	Map(f func(e Expression) Expression) OrderExpression
	Expressions() []ExpressionWithDirection
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
		part := fmt.Sprintf("%s", expression.Expression)
		if expression.Reverse {
			part += " desc"
		}

		parts = append(parts, part)
	}

	return strings.Join(parts, ", ")
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

func (e orderExpression) Expressions() []ExpressionWithDirection {
	return e.expressions
}

func SubsumesOrder(a, b OrderExpression) bool {
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
