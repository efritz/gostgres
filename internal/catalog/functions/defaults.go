package functions

import (
	"fmt"
	"time"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/types"
)

func DefaultFunctions() map[string]impls.Function {
	m := map[string]impls.Function{}
	for _, f := range []impls.Function{
		now,
		length,
		currval,
		nextval,
		setval,
	} {
		m[f.Name()] = f
	}

	return m
}

var now = newFunctionImpl(
	"now",
	nil,
	types.TypeTimestampTz,
	func(ctx impls.ExecutionContext, args []any) (any, error) {
		return time.Now(), nil
	},
)

var length = newFunctionImpl(
	"length",
	[]types.Type{types.TypeText},
	types.TypeBigInteger,
	func(ctx impls.ExecutionContext, args []any) (any, error) {
		value := args[0].(string)
		return int64(len(value)), nil
	},
)

var currval = newFunctionImpl(
	"currval",
	[]types.Type{types.TypeText},
	types.TypeBigInteger,
	func(ctx impls.ExecutionContext, args []any) (any, error) {
		name := args[0].(string)
		sequence, ok := ctx.Catalog().Sequences.Get(name)
		if !ok {
			return nil, fmt.Errorf("sequence %s does not exist", name)
		}

		return sequence.Value(), nil
	},
)

var nextval = newFunctionImpl(
	"nextval",
	[]types.Type{types.TypeText},
	types.TypeBigInteger,
	func(ctx impls.ExecutionContext, args []any) (any, error) {
		name := args[0].(string)
		sequence, ok := ctx.Catalog().Sequences.Get(name)
		if !ok {
			return nil, fmt.Errorf("sequence %s does not exist", name)
		}

		return sequence.Next()
	},
)

var setval = newFunctionImpl(
	"setval",
	[]types.Type{types.TypeText, types.TypeBigInteger},
	types.TypeBigInteger,
	func(ctx impls.ExecutionContext, args []any) (any, error) {
		name := args[0].(string)
		sequence, ok := ctx.Catalog().Sequences.Get(name)
		if !ok {
			return nil, fmt.Errorf("sequence %s does not exist", name)
		}

		value := args[1].(int64)
		return nil, sequence.Set(value)
	},
)
