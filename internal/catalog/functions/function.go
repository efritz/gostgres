package functions

import (
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/types"
)

type functionImpl struct {
	impls.Callable
	invoke func(ctx impls.ExecutionContext, args []any) (any, error)
}

var _ impls.Function = functionImpl{}

func newFunctionImpl(
	name string,
	paramTypes []types.Type,
	returnType types.Type,
	invoke func(ctx impls.ExecutionContext, args []any) (any, error),
) impls.Function {
	return functionImpl{
		Callable: impls.NewCallable(name, paramTypes, returnType),
		invoke:   invoke,
	}
}

func (f functionImpl) Invoke(ctx impls.ExecutionContext, args []any) (any, error) {
	refinedArgs, err := f.Callable.RefineArgValues(args)
	if err != nil {
		return nil, err
	}

	return f.invoke(ctx, refinedArgs)
}
