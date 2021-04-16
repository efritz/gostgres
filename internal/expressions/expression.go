package expressions

import (
	"github.com/efritz/gostgres/internal/shared"
)

type Expression interface {
	Fields() []shared.Field
	Fold() Expression
	Alias(field shared.Field, expression Expression) Expression
	Conjunctions() []Expression
	ValueFrom(row shared.Row) (interface{}, error)
}
