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

func EnsureInt(val any, err error) (int, error) {
	if err != nil {
		return 0, err
	}

	typedVal, ok := val.(int)
	if !ok {
		return 0, fmt.Errorf("unexpected type (wanted int, have %v)", val)
	}
	return typedVal, nil
}

func EnsureBool(val any, err error) (bool, error) {
	if err != nil {
		return false, err
	}

	typedVal, ok := val.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected type (wanted bool, have %v)", val)
	}
	return typedVal, nil
}

func EnsureNullableBool(val any, err error) (*bool, error) {
	if err != nil || val == nil {
		return nil, err
	}

	typedVal, ok := val.(bool)
	if !ok {
		return nil, fmt.Errorf("unexpected type (wanted bool, have %v)", val)
	}
	return &typedVal, nil
}

func EnsureString(val any, err error) (string, error) {
	if err != nil {
		return "", err
	}

	typedVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("unexpected type (wanted string, have %v)", val)
	}
	return typedVal, nil
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
