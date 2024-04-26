package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type Expression interface {
	fmt.Stringer

	Equal(other Expression) bool
	Fields() []shared.Field
	Named() (shared.Field, bool)
	Fold() Expression
	Alias(field shared.Field, expression Expression) Expression
	Conjunctions() []Expression
	ValueFrom(row shared.Row) (any, error)
}
