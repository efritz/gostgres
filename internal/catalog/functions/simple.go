package functions

import "github.com/efritz/gostgres/internal/shared/impls"

type simpleFunction func(ctx impls.Context, args []any) (any, error)

var _ impls.Function = simpleFunction(nil)

func (f simpleFunction) Invoke(ctx impls.Context, args []any) (any, error) {
	return f(ctx, args)
}
