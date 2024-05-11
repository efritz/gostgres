package expressions

import (
	"github.com/efritz/gostgres/internal/functions"
	"github.com/efritz/gostgres/internal/sequence"
)

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
