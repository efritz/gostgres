package projection

import (
	"fmt"
	"strings"

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

//
//

func SerializeProjectedExpressions(projectedExpressions []ProjectedExpression) string {
	relationNames := map[string]struct{}{}
	for _, expression := range projectedExpressions {
		if named, ok := expression.Expression.(expressions.NamedExpression); ok {
			relationNames[named.Field().RelationName()] = struct{}{}
		}
	}

	fields := make([]string, 0, len(projectedExpressions))
	for _, expression := range projectedExpressions {
		// TODO - simplify named expressions below top-level?
		if named, ok := expression.Expression.(expressions.NamedExpression); ok {
			name := named.Field().String()
			if len(relationNames) == 1 {
				name = named.Field().Name()
			}

			if named.Field().Name() == expression.Alias {
				fields = append(fields, name)
			} else {
				fields = append(fields, fmt.Sprintf("%s as %s", name, expression.Alias))
			}

			continue
		}

		fields = append(fields, expression.String())
	}

	return strings.Join(fields, ", ")
}

//
//

func FieldsFromProjectedExpressions(targetRelationName string, projectedExpressions []ProjectedExpression) []fields.Field {
	var projectedFields []fields.Field
	for _, field := range projectedExpressions {
		relationName := targetRelationName
		if relationName == "" {
			if named, ok := field.Expression.(expressions.NamedExpression); ok {
				relationName = named.Field().RelationName()
			}
		}

		projectedFields = append(projectedFields, fields.NewField(relationName, field.Alias, field.Expression.Type(), fields.NonInternalField))
	}

	return projectedFields
}
