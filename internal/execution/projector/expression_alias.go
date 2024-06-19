package projector

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type aliasProjectionExpression struct {
	expression impls.Expression
	alias      string
}

func NewAliasProjectionExpression(expression impls.Expression, alias string) ProjectionExpression {
	return aliasProjectionExpression{
		expression: expression,
		alias:      alias,
	}
}

func (p aliasProjectionExpression) String() string {
	return fmt.Sprintf("%s as %s", p.expression, p.alias)
}

func (p aliasProjectionExpression) Dealias(name string, fields []fields.Field, alias string) ProjectionExpression {
	expression := p.expression
	for _, field := range fields {
		expression = Alias(expression, field.WithRelationName(name), expressions.NewNamed(field))
	}

	return aliasProjectionExpression{
		expression: expression,
		alias:      p.alias,
	}
}

func (p aliasProjectionExpression) Expand(fields []fields.Field) ([]aliasProjectionExpression, error) {
	return []aliasProjectionExpression{p}, nil
}

//
//

func Alias(e impls.Expression, field fields.Field, target impls.Expression) impls.Expression {
	return e.Map(func(e impls.Expression) impls.Expression {
		if named, ok := e.(expressions.NamedExpression); ok {
			if field.RelationName() == "" || named.Field().RelationName() == field.RelationName() {
				if named.Field().Name() == field.Name() {
					return target
				}
			}
		}

		return e
	})
}

func UnwrapAlias(e ProjectionExpression) (impls.Expression, string, bool) {
	if alias, ok := e.(aliasProjectionExpression); ok {
		return alias.expression, alias.alias, true
	}

	return nil, "", false
}
