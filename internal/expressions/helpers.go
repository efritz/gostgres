package expressions

import "github.com/efritz/gostgres/internal/shared"

func tryEvaluate(expression Expression) Expression {
	if value, err := expression.ValueFrom(EmptyContext, shared.Row{}); err == nil {
		return NewConstant(value)
	}

	return expression
}
