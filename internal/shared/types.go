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
	Kind     TypeKind
	Nullable bool
}

var (
	TypeText                    = Type{Kind: TypeKindText, Nullable: false}
	TypeNullableText            = Type{Kind: TypeKindText, Nullable: true}
	TypeSmallInteger            = Type{Kind: TypeKindSmallInteger, Nullable: false}
	TypeNullableSmallInteger    = Type{Kind: TypeKindSmallInteger, Nullable: true}
	TypeInteger                 = Type{Kind: TypeKindInteger, Nullable: false}
	TypeNullableInteger         = Type{Kind: TypeKindInteger, Nullable: true}
	TypeBigInteger              = Type{Kind: TypeKindBigInteger, Nullable: false}
	TypeNullableBigInteger      = Type{Kind: TypeKindBigInteger, Nullable: true}
	TypeReal                    = Type{Kind: TypeKindReal, Nullable: false}
	TypeNullableReal            = Type{Kind: TypeKindReal, Nullable: true}
	TypeDoublePrecision         = Type{Kind: TypeKindDoublePrecision, Nullable: false}
	TypeNullableDoublePrecision = Type{Kind: TypeKindDoublePrecision, Nullable: true}
	TypeNumeric                 = Type{Kind: TypeKindNumeric, Nullable: false}
	TypeNullableNumeric         = Type{Kind: TypeKindNumeric, Nullable: true}
	TypeBool                    = Type{Kind: TypeKindBool, Nullable: false}
	TypeNullableBool            = Type{Kind: TypeKindBool, Nullable: true}
	TypeTimestampTz             = Type{Kind: TypeKindTimestampTz, Nullable: false}
	TypeNullableTimestampTz     = Type{Kind: TypeKindTimestampTz, Nullable: true}
	TypeAny                     = Type{Kind: TypeKindAny, Nullable: false}
	TypeNullableAny             = Type{Kind: TypeKindAny, Nullable: true}
)

func (typ Type) String() string {
	if typ.Nullable {
		return fmt.Sprintf("%s?", typ.Kind)
	}

	return typ.Kind.String()
}

func (typ Type) Equals(other Type) bool {
	return typ.Kind == other.Kind && typ.Nullable == other.Nullable
}

func (typ Type) NonNullable() Type {
	return Type{Kind: typ.Kind, Nullable: false}
}

func (typ Type) Refine(value any) (Type, any, bool) {
	if value == nil {
		return typ, nil, typ.Nullable
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
