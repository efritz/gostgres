package expressions

import (
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type NamedExpression interface {
	Field() fields.Field
}

type CompositeExpression interface {
	Children() []impls.Expression
}

func Fields(expr impls.Expression) []fields.Field {
	return gatherFields(expr, nil)
}

func gatherFields(expr impls.Expression, fields []fields.Field) []fields.Field {
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
