package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type Expression interface {
	fmt.Stringer

	Equal(other Expression) bool
	Fold() Expression
	Map(f func(Expression) Expression) Expression
	ValueFrom(context ExpressionContext, row shared.Row) (any, error)
}

type NamedExpression interface {
	Field() shared.Field
}

type CompositeExpression interface {
	Children() []Expression
}

func Fields(expr Expression) []shared.Field {
	return gatherFields(expr, nil)
}

func gatherFields(expr Expression, fields []shared.Field) []shared.Field {
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
