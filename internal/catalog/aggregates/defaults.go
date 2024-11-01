package aggregates

import (
	"fmt"
	"math/big"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/ordering"
	"github.com/efritz/gostgres/internal/shared/types"
	"golang.org/x/exp/constraints"
)

func DefaultAggregates() map[string]impls.Aggregate {
	return map[string]impls.Aggregate{
		"count": aggregatorFuncPair{countStep, countDone},
		"sum":   simpleAggregateFunc(sum),
		"min":   simpleAggregateFunc(min),
		"max":   simpleAggregateFunc(max),
	}
}

func countStep(state any, _ []any) (any, error) {
	switch acc := state.(type) {
	case nil:
		return int64(1), nil
	case int64:
		return acc + 1, nil
	}

	panic("count() requires a numeric state")
}

func countDone(state any) (any, error) {
	switch acc := state.(type) {
	case nil:
		return int64(0), nil
	case int64:
		return acc, nil
	}

	panic("count() requires a numeric state")
}

func sum(state any, args []any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sum() takes one argument")
	}
	if !types.IsNumeric(args[0]) {
		return nil, fmt.Errorf("sum() takes one argument of a numeric type")
	}

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
}

func addNumbers[T constraints.Integer | constraints.Float](a, b T) (T, error) {
	return a + b, nil
}

func min(state any, args []any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("min() takes one argument")
	}

	// TODO - test OrderTypeIncomparable
	if state == nil || ordering.CompareValues(args[0], state) == ordering.OrderTypeBefore {
		state = args[0]
	}

	return state, nil
}

func max(state any, args []any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("max() takes one argument")
	}

	// TODO - test OrderTypeIncomparable
	if state == nil || ordering.CompareValues(args[0], state) == ordering.OrderTypeAfter {
		state = args[0]
	}

	return state, nil
}
