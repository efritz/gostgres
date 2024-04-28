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
	TypeKindBinary
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
	case TypeKindBinary:
		return "bytea"
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
	TypeBool                    = Type{Kind: TypeKindBool, Nullable: false} // TODO - parse true yes on 1 (or unique prefix); false no off 0 (or unique prefix)
	TypeNullableBool            = Type{Kind: TypeKindBool, Nullable: true}
	TypeBinary                  = Type{Kind: TypeKindBinary, Nullable: false}
	TypeNullableBinary          = Type{Kind: TypeKindBinary, Nullable: true}
	TypeTimestampTz             = Type{Kind: TypeKindTimestampTz, Nullable: false} // TODO - add date, time, and interval
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
		if typ.Kind == TypeKindTimestampTz {
			if t, err := time.Parse("2006-01-02 15:04:05-07", v); err == nil {
				return typ, t, true
			}
		} else {
			t, ok := typ.refineToKind(TypeKindText)
			return t, v, ok
		}
	case int16:
		t, ok := typ.refineToKind(TypeKindSmallInteger, TypeKindInteger, TypeKindBigInteger, TypeKindDoublePrecision, TypeKindNumeric)
		return t, v, ok
	case int32:
		t, ok := typ.refineToKind(TypeKindInteger, TypeKindBigInteger, TypeKindDoublePrecision, TypeKindNumeric)
		return t, v, ok
	case int64:
		t, ok := typ.refineToKind(TypeKindBigInteger, TypeKindDoublePrecision, TypeKindNumeric)
		return t, v, ok
	case float32:
		t, ok := typ.refineToKind(TypeKindReal, TypeKindDoublePrecision, TypeKindNumeric)
		return t, v, ok
	case float64:
		t, ok := typ.refineToKind(TypeKindDoublePrecision, TypeKindNumeric)
		return t, v, ok
	case *big.Float:
		t, ok := typ.refineToKind(TypeKindNumeric)
		return t, v, ok
	case bool:
		t, ok := typ.refineToKind(TypeKindBool)
		return t, v, ok
	case time.Time:
		t, ok := typ.refineToKind(TypeKindTimestampTz)
		return t, v, ok
	}

	return Type{}, nil, false
}

func (typ Type) refineToKind(kinds ...TypeKind) (Type, bool) {
	for _, kind := range append(kinds, TypeKindAny) {
		if typ.Kind == kind {
			return Type{Kind: kind, Nullable: typ.Nullable}, true
		}
	}

	return Type{}, false
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
