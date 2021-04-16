package shared

import "fmt"

func EnsureInt(val interface{}, err error) (int, error) {
	if err != nil {
		return 0, err
	}

	typedVal, ok := val.(int)
	if !ok {
		return 0, fmt.Errorf("unexpected type (wanted int, have %v)", val)
	}
	return typedVal, nil
}

func EnsureBool(val interface{}, err error) (bool, error) {
	if err != nil {
		return false, err
	}

	typedVal, ok := val.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected type (wanted bool, have %v)", val)
	}
	return typedVal, nil
}
