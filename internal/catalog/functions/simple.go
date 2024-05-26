package functions

import "github.com/efritz/gostgres/internal/types"

type simpleFunction func(ctx types.Context, args []any) (any, error)

var _ types.Function = simpleFunction(nil)

func (f simpleFunction) Invoke(ctx types.Context, args []any) (any, error) {
	return f(ctx, args)
}
