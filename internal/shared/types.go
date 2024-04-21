package shared

import "fmt"

type TypeKind int

const (
	TypeKindInvalid TypeKind = iota
	TypeKindText
	TypeKindNumeric
	TypeKindBool
	TypeKindAny
)

func (k TypeKind) String() string {
	switch k {
	case TypeKindInvalid:
		return "invalid"
	case TypeKindText:
		return "text"
	case TypeKindNumeric:
		return "numeric"
	case TypeKindBool:
		return "bool"
	}

	return "any"
}

type Type struct {
	Kind     TypeKind
	Nullable bool
}

var (
	TypeText            = Type{Kind: TypeKindText, Nullable: false}
	TypeNullableText    = Type{Kind: TypeKindText, Nullable: true}
	TypeNumeric         = Type{Kind: TypeKindNumeric, Nullable: false}
	TypeNullableNumeric = Type{Kind: TypeKindNumeric, Nullable: true}
	TypeBool            = Type{Kind: TypeKindBool, Nullable: false}
	TypeNullableBool    = Type{Kind: TypeKindBool, Nullable: true}
	TypeAny             = Type{Kind: TypeKindAny, Nullable: false}
	TypeNullableAny     = Type{Kind: TypeKindAny, Nullable: true}
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

func (typ Type) Refine(value any) (Type, bool) {
	if value == nil {
		return typ, typ.Nullable
	}

	switch value.(type) {
	case string:
		return typ.refineToKind(TypeKindText)
	case int:
		return typ.refineToKind(TypeKindNumeric)
	case bool:
		return typ.refineToKind(TypeKindBool)
	}

	return Type{}, false
}

func (typ Type) refineToKind(kind TypeKind) (Type, bool) {
	if typ.Kind == TypeKindAny || typ.Kind == kind {
		return Type{Kind: kind, Nullable: typ.Nullable}, true
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
