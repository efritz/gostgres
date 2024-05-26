package functions

import "github.com/efritz/gostgres/internal/types"

type function func(ctx types.Context, args []any) (any, error)

var _ types.Function = function(nil)

func (f function) Invoke(ctx types.Context, args []any) (any, error) {
	return f(ctx, args)
}
