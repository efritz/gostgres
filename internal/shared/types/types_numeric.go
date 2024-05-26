package types

import (
	"fmt"
	"math"
	"math/big"

	"golang.org/x/exp/constraints"
)

func refineNumeric(value any, typ Type) (Type, any, bool) {
	if typ == TypeAny {
		return typ, value, true
	}

	lIndex := numericTypeIndex(numericTypeKindFromValue(value))
	rIndex := numericTypeIndex(typ)

	if lIndex >= 0 && rIndex >= 0 {
		if vx, err := converters[lIndex][rIndex](value); err == nil {
			return typ, vx, true
		}
	}

	return TypeUnknown, nil, false
}

//
//

func PromoteToCommonNumericValues(left, right any) (any, any, error) {
	lIndex := numericTypeIndex(numericTypeKindFromValue(left))
	if lIndex < 0 {
		return nil, nil, fmt.Errorf("unexpected type (wanted numeric type, have %v)", left)
	}

	rIndex := numericTypeIndex(numericTypeKindFromValue(right))
	if rIndex < 0 {
		return nil, nil, fmt.Errorf("unexpected type (wanted numeric type, have %v)", right)
	}

	if lIndex < rIndex {
		temp, err := converters[lIndex][rIndex](left)
		if err != nil {
			return nil, nil, err
		}

		left = temp
	}

	if rIndex < lIndex {
		temp, err := converters[rIndex][lIndex](right)
		if err != nil {
			return nil, nil, err
		}

		right = temp
	}

	return left, right, nil
}

//
//

var converters = [][]func(value any) (any, error){
	convertFrom[int16](),
	convertFrom[int32](),
	convertFrom[int64](),
	convertFrom[float32](),
	convertFrom[float64](),
	convertFromBigFloat(),
}

func convertFrom[T constraints.Integer | constraints.Float]() []func(value any) (any, error) {
	return []func(value any) (any, error){
		func(value any) (any, error) { return toSmallInteger(value.(T)) },
		func(value any) (any, error) { return toInteger(value.(T)) },
		func(value any) (any, error) { return toBigInteger(value.(T)) },
		func(value any) (any, error) { return toReal(value.(T)) },
		func(value any) (any, error) { return toDoublePrecision(value.(T)) },
		func(value any) (any, error) { return big.NewFloat(float64(value.(T))), nil },
	}
}

func convertFromBigFloat() []func(value any) (any, error) {
	fromBigFloat := func(f func(value float64) (any, error)) func(value any) (any, error) {
		return func(value any) (any, error) {
			if v, accuracy := value.(*big.Float).Float64(); accuracy == big.Exact {
				return f(v)
			}

			return nil, fmt.Errorf("value %v is out of range for double precision", value)
		}
	}

	return []func(value any) (any, error){
		fromBigFloat(toSmallInteger),
		fromBigFloat(toInteger),
		fromBigFloat(toBigInteger),
		fromBigFloat(toReal),
		fromBigFloat(toDoublePrecision),
		func(value any) (any, error) { return value, nil },
	}
}

func toSmallInteger[T constraints.Integer | constraints.Float](value T) (any, error) {
	return convertInteger[int16](value, "small integer", math.MaxInt16)
}

func toInteger[T constraints.Integer | constraints.Float](value T) (any, error) {
	return convertInteger[int32](value, "integer", math.MaxInt32)
}

func toBigInteger[T constraints.Integer | constraints.Float](value T) (any, error) {
	return convertInteger[int64](value, "bigint", math.MaxInt64)
}

func toReal[T constraints.Integer | constraints.Float](value T) (any, error) {
	return convertFloat[float32](value, "real", math.MaxFloat32)
}

func toDoublePrecision[T constraints.Integer | constraints.Float](value T) (any, error) {
	return convertFloat[float64](value, "double precision", math.MaxFloat64)
}

func convertInteger[T constraints.Integer, R constraints.Integer | constraints.Float](value R, name string, maxValue T) (any, error) {
	if int64(value) > int64(maxValue) {
		return nil, fmt.Errorf("value %v is out of range for %s", value, name)
	}

	return T(value), nil
}

func convertFloat[T constraints.Float, R constraints.Integer | constraints.Float](value R, name string, maxValue T) (any, error) {
	if float64(value) > float64(maxValue) {
		return nil, fmt.Errorf("value %v is out of range for %s", value, name)
	}

	return T(value), nil
}

//
//

func IsNumeric(value any) bool {
	return numericTypeKindFromValue(value) >= 0
}

func numericTypeIndex(value Type) int {
	return int(value - TypeSmallInteger)
}

func numericTypeKindFromValue(value any) Type {
	switch value.(type) {
	case int16:
		return TypeSmallInteger
	case int32:
		return TypeInteger
	case int64:
		return TypeBigInteger
	case float32:
		return TypeReal
	case float64:
		return TypeDoublePrecision
	case *big.Float:
		return TypeNumeric
	}

	return -1
}
