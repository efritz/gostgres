package expressions

import (
	"github.com/efritz/gostgres/internal/aggregates"
	"github.com/efritz/gostgres/internal/functions"
	"github.com/efritz/gostgres/internal/sequence"
)

type ExpressionContext interface {
	functions.FunctionContext
	GetFunction(name string) (functions.Function, bool)
	GetAggregate(name string) (aggregates.Aggregate, bool)
}

type noopContext struct{}

var EmptyContext = noopContext{}

func (ctx noopContext) GetSequence(name string) (*sequence.Sequence, bool)    { return nil, false }
func (ctx noopContext) GetFunction(name string) (functions.Function, bool)    { return nil, false }
func (ctx noopContext) GetAggregate(name string) (aggregates.Aggregate, bool) { return nil, false }
