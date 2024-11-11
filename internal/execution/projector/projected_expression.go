package projector

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type ProjectedExpression struct {
	expression impls.Expression
	alias      string
}

func NewProjectedExpression(expression impls.Expression, alias string) ProjectionExpression {
	return ProjectedExpression{
		expression: expression,
		alias:      alias,
	}
}

func NewProjectedExpressionFromField(field fields.Field) ProjectionExpression {
	return NewProjectedExpression(expressions.NewNamed(field), field.Name())
}

func (p ProjectedExpression) String() string {
	return fmt.Sprintf("%s as %s", p.expression, p.alias)
}

func (p ProjectedExpression) Dealias(name string, fields []fields.Field, alias string) ProjectionExpression {
	expression := p.expression
	for _, field := range fields {
		expression = Alias(expression, field.WithRelationName(name), expressions.NewNamed(field))
	}

	return ProjectedExpression{
		expression: expression,
		alias:      p.alias,
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

func UnwrapAlias(e ProjectionExpression) (impls.Expression, string, bool) {
	if alias, ok := e.(ProjectedExpression); ok {
		return alias.expression, alias.alias, true
	}

	return nil, "", false
}
