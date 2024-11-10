package aggregates

import (
	"math/big"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/ordering"
	"github.com/efritz/gostgres/internal/shared/types"
	"golang.org/x/exp/constraints"
)

func DefaultAggregates() map[string]impls.Aggregate {
	m := map[string]impls.Aggregate{}
	for _, a := range []impls.Aggregate{
		count,
		sum,
		min,
		max,
	} {
		m[a.Name()] = a
	}

	return m
}

var count = newAggregateImpl(
	"count",
	[]types.Type{types.TypeAny},
	types.TypeBigInteger,
	func(ctx impls.ExecutionContext, state any, args []any) (any, error) {
		switch acc := state.(type) {
		case nil:
			return int64(1), nil
		case int64:
			if args[0] == nil {
				return acc, nil
			}

			return acc + 1, nil
		}

		panic("invalid state for count")
	},
	func(ctx impls.ExecutionContext, state any) (any, error) {
		switch acc := state.(type) {
		case nil:
			return int64(0), nil
		case int64:
			return acc, nil
		}

		panic("count() requires a numeric state")
	},
)

var sum = newAggregateImpl(
	"sum",
	[]types.Type{types.TypeBigInteger},
	types.TypeInteger,
	func(ctx impls.ExecutionContext, state any, args []any) (any, error) {
		if state == nil {
			return args[0], nil
		}

		a, b, err := types.PromoteToCommonNumericValues(state, args[0])
		if err != nil {
			return nil, err
		}

		// TODO - roll this into a numeric shared package to use with expressions
		switch v := a.(type) {
		case int16:
			return addNumbers(v, b.(int16))
		case int32:
			return addNumbers(v, b.(int32))
		case int64:
			return addNumbers(v, b.(int64))
		case float32:
			return addNumbers(v, b.(float32))
		case float64:
			return addNumbers(v, b.(float64))
		case *big.Float:
			return new(big.Float).Add(v, b.(*big.Float)), nil
		}

		panic("unreachable after promotion")
	},
	func(ctx impls.ExecutionContext, state any) (any, error) {
		return state, nil
	},
)

var min = newAggregateImpl(
	"min",
	[]types.Type{types.TypeAny},
	types.TypeAny,
	func(ctx impls.ExecutionContext, state any, args []any) (any, error) {
		// TODO - test OrderTypeIncomparable
		if state == nil || ordering.CompareValues(args[0], state) == ordering.OrderTypeBefore {
			state = args[0]
		}

		return state, nil
	},
	func(ctx impls.ExecutionContext, state any) (any, error) {
		return state, nil
	},
)

var max = newAggregateImpl(
	"max",
	[]types.Type{types.TypeAny},
	types.TypeAny,
	func(ctx impls.ExecutionContext, state any, args []any) (any, error) {
		// TODO - test OrderTypeIncomparable
		if state == nil || ordering.CompareValues(args[0], state) == ordering.OrderTypeAfter {
			state = args[0]
		}

		return state, nil
	},
	func(ctx impls.ExecutionContext, state any) (any, error) {
		return state, nil
	},
)

func addNumbers[T constraints.Integer | constraints.Float](a, b T) (T, error) {
	return a + b, nil
}
