package functions

import (
	"fmt"
	"time"

	"github.com/efritz/gostgres/internal/types"
)

func DefaultFunctions() map[string]types.Function {
	return map[string]types.Function{
		"now":     function(now),
		"nextval": function(nextval),
		"setval":  function(setval),
		"currval": function(currval),
	}
}

func now(ctx types.Context, args []any) (any, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("now() takes no arguments")
	}

	return time.Now(), nil
}

func nextval(ctx types.Context, args []any) (any, error) {
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

func setval(ctx types.Context, args []any) (any, error) {
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

func currval(ctx types.Context, args []any) (any, error) {
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
