package aggregates

import "github.com/efritz/gostgres/internal/types"

type aggregatorFuncPair struct {
	step func(state any, args []any) (any, error)
	done func(state any) (any, error)
}

var _ types.Aggregate = aggregatorFuncPair{}

func (p aggregatorFuncPair) Step(state any, args []any) (any, error) { return p.step(state, args) }
func (p aggregatorFuncPair) Done(state any) (any, error)             { return p.done(state) }
