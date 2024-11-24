package projection

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

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

func Dealias(e impls.Expression, field fields.Field, target impls.Expression) impls.Expression {
	mapped, _ := e.Map(func(e impls.Expression) (impls.Expression, error) {
		if named, ok := e.(expressions.NamedExpression); ok {
			if named.Field().RelationName() == "" || named.Field().RelationName() == field.RelationName() {
				if named.Field().Name() == field.Name() {
					return target, nil
				}
			}
		}

		return e, nil
	})

	return mapped
}
