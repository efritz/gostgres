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

func refineType(expectedType TypeKind, value any) TypeKind {
	if _, ok := value.(string); ok {
		if expectedType == TypeKindAny || expectedType == TypeKindText {
			return TypeKindText
		}
	}
	if _, ok := value.(int); ok {
		if expectedType == TypeKindAny || expectedType == TypeKindNumeric {
			return TypeKindNumeric
		}
	}
	if _, ok := value.(bool); ok {
		if expectedType == TypeKindAny || expectedType == TypeKindBool {
			return TypeKindBool
		}
	}

	return TypeKindInvalid
}
