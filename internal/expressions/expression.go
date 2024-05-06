package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/functions"
	"github.com/efritz/gostgres/internal/shared"
)

type Expression interface {
	fmt.Stringer

	Equal(other Expression) bool
	Fields() []shared.Field
	Named() (shared.Field, bool)
	Fold() Expression
	Map(f func(Expression) Expression) Expression
	Conjunctions() []Expression
	ValueFrom(context Context, row shared.Row) (any, error)
}

type Context struct {
	Functions *functions.Functionspace
}

var EmptyContext = NewContext(functions.NewFunctionspace())

func NewContext(functions *functions.Functionspace) Context {
	return Context{Functions: functions}
}
