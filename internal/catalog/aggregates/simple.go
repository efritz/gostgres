package aggregates

import "github.com/efritz/gostgres/internal/types"

type simpleAggregateFunc func(state any, args []any) (any, error)

var _ types.Aggregate = simpleAggregateFunc(nil)

func (f simpleAggregateFunc) Step(state any, args []any) (any, error) { return f(state, args) }
func (f simpleAggregateFunc) Done(state any) (any, error)             { return state, nil }
