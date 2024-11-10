package aggregates

import (
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/types"
)

type aggregateImpl struct {
	impls.Callable
	step stepFunc
	done doneFunc
}

type stepFunc func(ctx impls.ExecutionContext, state any, args []any) (any, error)
type doneFunc func(ctx impls.ExecutionContext, state any) (any, error)

var _ impls.Aggregate = aggregateImpl{}

func newAggregateImpl(
	name string,
	paramTypes []types.Type,
	returnType types.Type,
	step stepFunc,
	done doneFunc,
) impls.Aggregate {
	return aggregateImpl{
		Callable: impls.NewCallable(name, paramTypes, returnType),
		step:     step,
		done:     done,
	}
}

func (a aggregateImpl) Step(ctx impls.ExecutionContext, state any, args []any) (any, error) {
	refinedArgs, err := a.Callable.RefineArgValues(args)
	if err != nil {
		return nil, err
	}

	return a.step(ctx, state, refinedArgs)
}

func (a aggregateImpl) Done(ctx impls.ExecutionContext, state any) (any, error) {
	return a.done(ctx, state)
}
