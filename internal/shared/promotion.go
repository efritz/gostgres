package shared

import (
	"fmt"
	"math/big"
)

func PromoteToCommonNumericValues(left, right any) (any, any, error) {
	lIndex := numericTypeIndex(left)
	if lIndex < 0 {
		return nil, nil, fmt.Errorf("unexpected type (wanted numeric type, have %v)", left)
	}

	rIndex := numericTypeIndex(right)
	if rIndex < 0 {
		return nil, nil, fmt.Errorf("unexpected type (wanted numeric type, have %v)", right)
	}

	if lIndex < rIndex {
		return promoters[lIndex][rIndex](left), right, nil
	}

	if rIndex < lIndex {
		return left, promoters[rIndex][lIndex](right), nil
	}

	return left, right, nil
}

var promoters = [][]func(value any) any{
	{
		func(value any) any { panic("impossible promotion") },               // smallint -> smallint
		func(value any) any { return int32(value.(int16)) },                 // smallint -> integer
		func(value any) any { return int64(value.(int16)) },                 // smallint -> bigint
		func(value any) any { return float32(value.(int16)) },               // smallint -> real
		func(value any) any { return float64(value.(int16)) },               // smallint -> double precision
		func(value any) any { return big.NewFloat(float64(value.(int16))) }, // smallint -> numeric
	},
	{
		func(value any) any { panic("impossible promotion") },               // integer -> smallint
		func(value any) any { panic("impossible promotion") },               // integer -> integer
		func(value any) any { return int64(value.(int32)) },                 // integer -> bigint
		func(value any) any { return float32(value.(int32)) },               // integer -> real
		func(value any) any { return float64(value.(int32)) },               // integer -> double precision
		func(value any) any { return big.NewFloat(float64(value.(int32))) }, // integer -> numeric
	},
	{
		func(value any) any { panic("impossible promotion") },               // bigint -> smallint
		func(value any) any { panic("impossible promotion") },               // bigint -> integer
		func(value any) any { panic("impossible promotion") },               // bigint -> bigint
		func(value any) any { return float32(value.(int64)) },               // bigint -> real (as double precision)
		func(value any) any { return float64(value.(int64)) },               // bigint -> double precision
		func(value any) any { return big.NewFloat(float64(value.(int64))) }, // bigint -> numeric
	},
	{
		func(value any) any { panic("impossible promotion") },                 // real -> smallint
		func(value any) any { panic("impossible promotion") },                 // real -> integer
		func(value any) any { panic("impossible promotion") },                 // real -> bigint
		func(value any) any { panic("impossible promotion") },                 // real -> real
		func(value any) any { return float64(value.(float32)) },               // real -> double precision
		func(value any) any { return big.NewFloat(float64(value.(float32))) }, // real -> numeric
	},
	{
		func(value any) any { panic("impossible promotion") },                 // double precision -> smallint
		func(value any) any { panic("impossible promotion") },                 // double precision -> integer
		func(value any) any { panic("impossible promotion") },                 // double precision -> bigint
		func(value any) any { panic("impossible promotion") },                 // double precision -> real
		func(value any) any { panic("impossible promotion") },                 // double precision -> double precision
		func(value any) any { return big.NewFloat(float64(value.(float64))) }, // double precision -> numeric
	},
	{
		func(value any) any { panic("impossible promotion") }, // numeric -> smallint
		func(value any) any { panic("impossible promotion") }, // smallint -> integer
		func(value any) any { panic("impossible promotion") }, // smallint -> bigint
		func(value any) any { panic("impossible promotion") }, // numeric -> real
		func(value any) any { panic("impossible promotion") }, // numeric -> double precision
		func(value any) any { panic("impossible promotion") }, // numeric -> numeric
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
