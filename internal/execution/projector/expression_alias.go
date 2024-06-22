package projector

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type aliasProjectionExpression struct {
	expression   impls.Expression
	relationName string
	aliasName    string
}

func NewAliasProjectionExpression(expression impls.Expression, aliasName string) ProjectionExpression {
	return aliasProjectionExpression{
		expression: expression,
		aliasName:  aliasName,
	}
}

func (p aliasProjectionExpression) String() string {
	name := fmt.Sprintf("%q", p.aliasName)
	if p.relationName != "" {
		name = fmt.Sprintf("%q.%q", p.relationName, p.aliasName)
	}

	return fmt.Sprintf("%s as %s", p.expression, name)
}

func (p aliasProjectionExpression) Dealias(relationName string, fields []fields.Field, alias string) ProjectionExpression {
	expression := p.expression
	for _, field := range fields {
		expression = Alias(expression, field.WithRelationName(relationName), expressions.NewNamed(field))
	}

	return aliasProjectionExpression{
		expression:   expression,
		relationName: relationName,
		aliasName:    p.aliasName,
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

func UnwrapAlias(e ProjectionExpression) (impls.Expression, string, string, bool) {
	if alias, ok := e.(aliasProjectionExpression); ok {
		return alias.expression, alias.relationName, alias.aliasName, true
	}

	return nil, "", "", false
}
