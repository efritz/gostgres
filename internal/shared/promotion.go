package shared

import (
	"fmt"
	"math/big"
)

func PromoteToCommonNumericValues(left, right any) (any, any, error) {
	for {
		lIndex := numericTypeIndex(left)
		if lIndex < 0 {
			return nil, nil, fmt.Errorf("unexpected type (wanted numeric type, have %v)", left)
		}

		rIndex := numericTypeIndex(right)
		if rIndex < 0 {
			return nil, nil, fmt.Errorf("unexpected type (wanted numeric type, have %v)", right)
		}

		if lIndex == rIndex {
			break
		}

		if lIndex < rIndex {
			left = promoters[lIndex][rIndex](left)
		} else if rIndex < lIndex {
			right = promoters[rIndex][lIndex](right)
		}
	}

	return left, right, nil
}

var promoters = [][]func(value any) any{
	{
		// smallint ->
		func(value any) any { panic("impossible promotion") },               //  -> smallint
		func(value any) any { return int32(value.(int16)) },                 //  -> integer
		func(value any) any { return int64(value.(int16)) },                 //  -> bigint
		func(value any) any { return float32(value.(int16)) },               //  -> real
		func(value any) any { return float64(value.(int16)) },               //  -> double precision
		func(value any) any { return big.NewFloat(float64(value.(int16))) }, //  -> numeric
	},
	{
		// integer ->
		func(value any) any { panic("impossible promotion") },               // -> smallint
		func(value any) any { panic("impossible promotion") },               // -> integer
		func(value any) any { return int64(value.(int32)) },                 // -> bigint
		func(value any) any { return float32(value.(int32)) },               // -> real
		func(value any) any { return float64(value.(int32)) },               // -> double precision
		func(value any) any { return big.NewFloat(float64(value.(int32))) }, // -> numeric
	},
	{
		// bigint ->
		func(value any) any { panic("impossible promotion") },               // -> smallint
		func(value any) any { panic("impossible promotion") },               // -> integer
		func(value any) any { panic("impossible promotion") },               // -> bigint
		func(value any) any { return float64(value.(int64)) },               // -> real (NOTE: skips to double precision)
		func(value any) any { return float64(value.(int64)) },               // -> double precision
		func(value any) any { return big.NewFloat(float64(value.(int64))) }, // -> numeric
	},
	{
		// real ->
		func(value any) any { panic("impossible promotion") },                 // -> smallint
		func(value any) any { panic("impossible promotion") },                 // -> integer
		func(value any) any { panic("impossible promotion") },                 // -> bigint
		func(value any) any { panic("impossible promotion") },                 // -> real
		func(value any) any { return float64(value.(float32)) },               // -> double precision
		func(value any) any { return big.NewFloat(float64(value.(float32))) }, // -> numeric
	},
	{
		// double precision ->
		func(value any) any { panic("impossible promotion") },                 // -> smallint
		func(value any) any { panic("impossible promotion") },                 // -> integer
		func(value any) any { panic("impossible promotion") },                 // -> bigint
		func(value any) any { panic("impossible promotion") },                 // -> real
		func(value any) any { panic("impossible promotion") },                 // -> double precision
		func(value any) any { return big.NewFloat(float64(value.(float64))) }, // -> numeric
	},
	{
		// numeric ->
		func(value any) any { panic("impossible promotion") }, // -> smallint
		func(value any) any { panic("impossible promotion") }, // -> integer
		func(value any) any { panic("impossible promotion") }, // -> bigint
		func(value any) any { panic("impossible promotion") }, // -> real
		func(value any) any { panic("impossible promotion") }, // -> double precision
		func(value any) any { panic("impossible promotion") }, // -> numeric
	},
}

func numericTypeIndex(value any) int {
	switch value.(type) {
	case int16:
		return 0
	case int32:
		return 1
	case int64:
		return 2
	case float32:
		return 3
	case float64:
		return 4
	case *big.Float:
		return 5
	}

	return -1
}
