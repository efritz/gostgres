package projection

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

func MapFieldToExpression(e impls.Expression, field fields.Field, target impls.Expression) impls.Expression {
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

func MapExpressionToField(e impls.Expression, targetRelationName string, proj ProjectedExpression) impls.Expression {
	mapped, _ := e.Map(func(e impls.Expression) (impls.Expression, error) {
		if e.Equal(proj.Expression) {
			return expressions.NewNamed(fields.NewField(targetRelationName, proj.Alias, e.Type(), fields.NonInternalField)), nil
		}

		return e, nil
	})

	return mapped
}
