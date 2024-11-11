package projector

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

var _ ProjectionExpression1 = &ProjectedExpression{}
var _ ProjectionExpression2 = &ProjectedExpression{}

func NewProjectedExpression(expression impls.Expression, alias string) ProjectionExpression {
	return ProjectedExpression{
		Expression: expression,
		Alias:      alias,
	}
}

func NewProjectedExpressionFromField(field fields.Field) ProjectionExpression {
	return NewProjectedExpression(expressions.NewNamed(field), field.Name())
}

func (p ProjectedExpression) String() string {
	return fmt.Sprintf("%s as %s", p.Expression, p.Alias)
}

func (p ProjectedExpression) Dealias(name string, fields []fields.Field, alias string) ProjectionExpression {
	expression := p.Expression
	for _, field := range fields {
		expression = Alias(expression, field.WithRelationName(name), expressions.NewNamed(field))
	}

	return ProjectedExpression{
		Expression: expression,
		Alias:      p.Alias,
	}
}

func (p ProjectedExpression) Expand(fields []fields.Field) ([]ProjectedExpression, error) {
	return []ProjectedExpression{p}, nil
}

//
//

func Alias(e impls.Expression, field fields.Field, target impls.Expression) impls.Expression {
	mapped, _ := e.Map(func(e impls.Expression) (impls.Expression, error) {
		if named, ok := e.(expressions.NamedExpression); ok {
			if field.RelationName() == "" || named.Field().RelationName() == field.RelationName() {
				if named.Field().Name() == field.Name() {
					return target, nil
				}
			}
		}

		return e, nil
	})

	return mapped
}
