package aggregates

import "github.com/efritz/gostgres/internal/types"

type aggregatorFuncPair struct {
	step func(state any, args []any) (any, error)
	done func(state any) (any, error)
}

var _ types.Aggregate = aggregatorFuncPair{}

func (p aggregatorFuncPair) Step(state any, args []any) (any, error) { return p.step(state, args) }
func (p aggregatorFuncPair) Done(state any) (any, error)             { return p.done(state) }

type simpleAggregateFunc func(state any, args []any) (any, error)

var _ types.Aggregate = simpleAggregateFunc(nil)

func (f simpleAggregateFunc) Step(state any, args []any) (any, error) { return f(state, args) }
func (f simpleAggregateFunc) Done(state any) (any, error)             { return state, nil }
