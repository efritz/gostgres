package projector

import "github.com/efritz/gostgres/internal/shared/fields"

type ProjectionExpression interface {
	Dealias(name string, fields []fields.Field, alias string) ProjectionExpression
	Expand(fields []fields.Field) ([]aliasProjectionExpression, error)
}
