package projection

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type ProjectedExpression struct {
	Expression impls.Expression
	Alias      string
}

func NewProjectedExpression(expression impls.Expression, alias string) ProjectedExpression {
	return ProjectedExpression{
		Expression: expression,
		Alias:      alias,
	}
}

func NewProjectedExpressionFromField(field fields.Field) ProjectedExpression {
	return NewProjectedExpression(expressions.NewNamed(field), field.Name())
}

func (p ProjectedExpression) String() string {
	return fmt.Sprintf("%s as %s", p.Expression, p.Alias)
}
