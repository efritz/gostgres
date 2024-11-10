package impls

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/types"
)

type Callable interface {
	Name() string
	ParamTypes() []types.Type
	ReturnType() types.Type
	ValidateArgTypes(argTypes []types.Type) error
	RefineArgValues(args []any) ([]any, error)
}

type callable struct {
	name       string
	paramTypes []types.Type
	returnType types.Type
}

var _ Callable = callable{}

func NewCallable(
	name string,
	paramTypes []types.Type,
	returnType types.Type,
) Callable {
	return callable{
		name:       name,
		paramTypes: paramTypes,
		returnType: returnType,
	}
}

func (c callable) Name() string             { return c.name }
func (c callable) ParamTypes() []types.Type { return c.paramTypes }
func (c callable) ReturnType() types.Type   { return c.returnType }

func (c callable) ValidateArgTypes(argTypes []types.Type) error {
	if len(argTypes) != len(c.paramTypes) {
		return fmt.Errorf("%s expects %d arguments, got %d", c.name, len(c.paramTypes), len(argTypes))
	}

	for i, expectedType := range c.paramTypes {
		if expectedType != types.TypeAny && argTypes[i] != expectedType && !(argTypes[i].IsNumber() && expectedType.IsNumber()) {
			return fmt.Errorf("argument %d to %s expects type %s, got %s", i+1, c.name, expectedType, argTypes[i])
		}
	}

	return nil
}

func (c callable) RefineArgValues(args []any) ([]any, error) {
	var ts []types.Type
	for _, arg := range args {
		ts = append(ts, types.TypeKindFromValue(arg))
	}

	if err := c.ValidateArgTypes(ts); err != nil {
		return nil, err
	}

	var refinedArgs []any
	for i, arg := range args {
		_, refinedArg, _ := c.paramTypes[i].Refine(arg)
		refinedArgs = append(refinedArgs, refinedArg)
	}

	return refinedArgs, nil
}
