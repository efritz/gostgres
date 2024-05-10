package shared

import (
	"fmt"
	"math/big"
	"time"
)

type TypeKind int

const (
	TypeKindInvalid TypeKind = iota
	TypeKindText
	TypeKindSmallInteger
	TypeKindInteger
	TypeKindBigInteger
	TypeKindReal
	TypeKindDoublePrecision
	TypeKindNumeric
	TypeKindBool
	TypeKindTimestampTz
	TypeKindAny
)

func (k TypeKind) String() string {
	switch k {
	case TypeKindInvalid:
		return "invalid"
	case TypeKindText:
		return "text"
	case TypeKindSmallInteger:
		return "smallint"
	case TypeKindInteger:
		return "integer"
	case TypeKindBigInteger:
		return "bigint"
	case TypeKindReal:
		return "real"
	case TypeKindDoublePrecision:
		return "double precision"
	case TypeKindNumeric:
		return "numeric"
	case TypeKindBool:
		return "bool"
	case TypeKindTimestampTz:
		return "timestamp with time zone"
	}

	return "any"
}

type Type struct {
	Kind TypeKind // TODO - unwrap
}

var (
	TypeText            = Type{Kind: TypeKindText}
	TypeSmallInteger    = Type{Kind: TypeKindSmallInteger}
	TypeInteger         = Type{Kind: TypeKindInteger}
	TypeBigInteger      = Type{Kind: TypeKindBigInteger}
	TypeReal            = Type{Kind: TypeKindReal}
	TypeDoublePrecision = Type{Kind: TypeKindDoublePrecision}
	TypeNumeric         = Type{Kind: TypeKindNumeric}
	TypeBool            = Type{Kind: TypeKindBool}
	TypeTimestampTz     = Type{Kind: TypeKindTimestampTz}
	TypeAny             = Type{Kind: TypeKindAny}
)

func (typ Type) String() string {
	return typ.Kind.String()
}

func (typ Type) Equals(other Type) bool {
	return typ.Kind == other.Kind
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
		return typ, v, typ.Kind == TypeKindAny || typ.Kind == TypeKindBool
	case time.Time:
		return typ, v, typ.Kind == TypeKindAny || typ.Kind == TypeKindTimestampTz
	}

	return Type{}, nil, false
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
