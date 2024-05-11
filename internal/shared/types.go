package shared

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

func (typ Type) Refine(value any) (Type, any, bool) {
	if value == nil {
		return typ, nil, true
	}

	switch v := value.(type) {
	case string:
		return refineString(v, typ)
	case int16, int32, int64, float32, float64, *big.Float:
		return refineNumeric(v, typ)
	case bool:
		return typ, v, typ == TypeAny || typ == TypeBool
	case time.Time:
		return typ, v, typ == TypeAny || typ == TypeTimestampTz
	}

	return TypeUnknown, nil, false
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

func refineTypes(fields []Field, values []any) ([]Field, []any, error) {
	if len(fields) != len(values) {
		return nil, nil, fmt.Errorf("unexpected number of columns")
	}

	refinedFields := make([]Field, 0, len(fields))
	refinedValues := make([]any, 0, len(values))

	for i, field := range fields {
		refinedType, refinedValue, ok := field.typ.Refine(values[i])
		if !ok {
			return nil, nil, fmt.Errorf("type error (%v is not %s)", values[i], field.Type())
		}

		refinedFields = append(refinedFields, field.WithType(refinedType))
		refinedValues = append(refinedValues, refinedValue)
	}

	return refinedFields, refinedValues, nil
}
