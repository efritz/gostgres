package functions

import (
	"fmt"
	"time"

	"github.com/efritz/gostgres/internal/shared/impls"
)

func DefaultFunctions() map[string]impls.Function {
	return map[string]impls.Function{
		"now":     simpleFunction(now),
		"nextval": simpleFunction(nextval),
		"setval":  simpleFunction(setval),
		"currval": simpleFunction(currval),
	}
}

func now(ctx impls.Context, args []any) (any, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("now() takes no arguments")
	}

	return time.Now(), nil
}

func nextval(ctx impls.Context, args []any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("nextval() takes one argument")
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("nextval() takes one argument of type string")
	}

	sequence, ok := ctx.GetSequence(name)
	if !ok {
		return nil, fmt.Errorf("sequence %s does not exist", name)
	}

	return sequence.Next()
}

func setval(ctx impls.Context, args []any) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("setval() takes two arguments")
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("setval() takes two arguments of type string, biginteger")
	}
	value, ok := args[1].(int64)
	if !ok {
		return nil, fmt.Errorf("setval() takes two arguments of type string, biginteger")
	}

	sequence, ok := ctx.GetSequence(name)
	if !ok {
		return nil, fmt.Errorf("sequence %s does not exist", name)
	}

	return nil, sequence.Set(value)
}

func currval(ctx impls.Context, args []any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("currval() takes one argument")
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("currval() takes one argument of type string")
	}

	sequence, ok := ctx.GetSequence(name)
	if !ok {
		return nil, fmt.Errorf("sequence %s does not exist", name)
	}

	return sequence.Value(), nil
}
