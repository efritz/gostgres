package projector

import (
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type ProjectionExpression interface {
	Dealias(name string, fields []fields.Field, alias string) ProjectionExpression
	Expand(fields []fields.Field) ([]AliasProjectionExpression, error)
	Map(f func(impls.Expression) (impls.Expression, error)) (ProjectionExpression, error)
}
