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
	IsTID      bool
}

func NewProjectedExpression(expression impls.Expression, alias string, isTID bool) ProjectedExpression {
	return ProjectedExpression{
		Expression: expression,
		Alias:      alias,
		IsTID:      isTID,
	}
}

func NewProjectedExpressionFromField(field fields.Field) ProjectedExpression {
	return NewProjectedExpression(expressions.NewNamed(field), field.Name(), field.IsTID())
}

func (p ProjectedExpression) String() string {
	return fmt.Sprintf("%s as %s", p.Expression, p.Alias)
}

//
//

func fieldsFromProjectedExpressions(targetRelationName string, projectedExpressions []ProjectedExpression) []fields.Field {
	var projectedFields []fields.Field
	for _, proj := range projectedExpressions {
		relationName := targetRelationName
		if relationName == "" {
			if named, ok := proj.Expression.(expressions.NamedExpression); ok {
				relationName = named.Field().RelationName()
			}
		}

		internalType := fields.NonInternalField
		if proj.IsTID {
			internalType = fields.InternalFieldTid
		}

		projectedFields = append(projectedFields, fields.NewField(relationName, proj.Alias, proj.Expression.Type(), internalType))
	}

	return projectedFields
}

func serializeProjectedExpressions(projectedExpressions []ProjectedExpression) string {
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
