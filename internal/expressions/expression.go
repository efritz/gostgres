package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/functions"
	"github.com/efritz/gostgres/internal/sequence"
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
	ValueFrom(context ExpressionContext, row shared.Row) (any, error)
}

type ExpressionContext interface {
	GetFunction(name string) (functions.Function, bool)
	GetSequence(name string) (*sequence.Sequence, bool)
}

type noopContext struct{}

var EmptyContext = noopContext{}

func (ctx noopContext) GetFunction(name string) (functions.Function, bool) {
	return nil, false
}

func (ctx noopContext) GetSequence(name string) (*sequence.Sequence, bool) {
	return nil, false
}
