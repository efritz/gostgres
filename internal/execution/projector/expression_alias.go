package projector

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type AliasProjectionExpression struct {
	Expression impls.Expression
	Alias      string
}

func NewAliasProjectionExpression(expression impls.Expression, alias string) ProjectionExpression {
	return AliasProjectionExpression{
		Expression: expression,
		Alias:      alias,
	}
}

func (p AliasProjectionExpression) String() string {
	return fmt.Sprintf("%s as %s", p.Expression, p.Alias)
}

func (p AliasProjectionExpression) Dealias(name string, fields []fields.Field, alias string) ProjectionExpression {
	expression := p.Expression
	for _, field := range fields {
		expression = Alias(expression, field.WithRelationName(name), expressions.NewNamed(field))
	}

	return AliasProjectionExpression{
		Expression: expression,
		Alias:      p.Alias,
	}
}

func (p AliasProjectionExpression) Expand(fields []fields.Field) ([]AliasProjectionExpression, error) {
	return []AliasProjectionExpression{p}, nil
}

func (a AliasProjectionExpression) Map(f func(impls.Expression) (impls.Expression, error)) (ProjectionExpression, error) {
	expression, err := f(a.Expression)
	if err != nil {
		return nil, err
	}

	return NewAliasProjectionExpression(expression, a.Alias), nil
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
	if alias, ok := e.(AliasProjectionExpression); ok {
		return alias.Expression, alias.Alias, true
	}

	return nil, "", false
}
