package types

import (
	"fmt"
	"math/big"
	"time"
)

type Type int

const (
	TypeUnknown Type = iota
	TypeText
	TypeSmallInteger
	TypeInteger
	TypeBigInteger
	TypeReal
	TypeDoublePrecision
	TypeNumeric
	TypeBool
	TypeTimestampTz
	TypeAny
)

func (typ Type) String() string {
	switch typ {
	case TypeText:
		return "text"
	case TypeSmallInteger:
		return "smallint"
	case TypeInteger:
		return "integer"
	case TypeBigInteger:
		return "bigint"
	case TypeReal:
		return "real"
	case TypeDoublePrecision:
		return "double precision"
	case TypeNumeric:
		return "numeric"
	case TypeBool:
		return "bool"
	case TypeTimestampTz:
		return "timestamp with time zone"
	case TypeAny:
		return "any"
	}

	return "unknown"
}

func (typ Type) IsNumber() bool {
	switch typ {
	case TypeSmallInteger:
		return true
	case TypeInteger:
		return true
	case TypeBigInteger:
		return true
	case TypeReal:
		return true
	case TypeDoublePrecision:
		return true
	case TypeNumeric:
		return true
	}

	return false
}

func (typ Type) Refine(value any) (Type, any, bool) {
	if value == nil || typ == TypeAny {
		return typ, value, true
	}

	switch v := value.(type) {
	case string:
		return refineStringValue(v, typ)
	case int16, int32, int64, float32, float64, *big.Float:
		return refineNumericValue(v, typ)
	case bool:
		return typ, v, typ == TypeAny || typ == TypeBool
	case time.Time:
		return typ, v, typ == TypeAny || typ == TypeTimestampTz
	}

	return TypeUnknown, nil, false
}

func (typ Type) PromoteToCommonType(other Type) Type {
	if typ < other {
		return other.PromoteToCommonType(typ)
	}

	if typ == other {
		return typ
	}

	switch typ {
	case TypeAny:
		return other

	case TypeSmallInteger, TypeInteger, TypeBigInteger, TypeReal, TypeDoublePrecision, TypeNumeric:
		if common, err := PromoteToCommonNumericTypes(typ, other); err == nil {
			return common
		}

	default:
		// Not equal, not promotable
	}

	return TypeUnknown
}

func TypeKindFromValue(value any) Type {
	switch value.(type) {
	case string:
		return TypeText
	case int16, int32, int64, float32, float64, *big.Float:
		return numericTypeKindFromValue(value)
	case bool:
		return TypeBool
	case time.Time:
		return TypeTimestampTz
	}

	return TypeUnknown
}

func ValueAs[T any](untypedValue any, err error) (*T, error) {
	if err != nil || untypedValue == nil {
		return nil, err
	}

	typedValue, ok := untypedValue.(T)
	if !ok {
		return nil, fmt.Errorf("unexpected type (wanted %T, have %v)", typedValue, untypedValue)
	}

	return &typedValue, nil
}
