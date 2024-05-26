package expressions

import (
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

type NamedExpression interface {
	Field() shared.Field
}

type CompositeExpression interface {
	Children() []types.Expression
}

func Fields(expr types.Expression) []shared.Field {
	return gatherFields(expr, nil)
}

func gatherFields(expr types.Expression, fields []shared.Field) []shared.Field {
	if named, ok := expr.(NamedExpression); ok {
		fields = append(fields, named.Field())
	}

	if c, ok := expr.(CompositeExpression); ok {
		for _, child := range c.Children() {
			fields = gatherFields(child, fields)
		}
	}

	return fields
}
