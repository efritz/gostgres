package expressions

import (
	"fmt"
	"slices"
	"strings"

	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

type functionExpression struct {
	name string
	args []types.Expression
}

var _ types.Expression = functionExpression{}

func NewFunction(name string, args []types.Expression) types.Expression {
	return functionExpression{
		name: name,
		args: args,
	}
}

func (e functionExpression) String() string {
	var args []string
	for _, arg := range e.args {
		args = append(args, arg.String())
	}

	return fmt.Sprintf("%s(%s)", e.name, strings.Join(args, ", "))
}

func (e functionExpression) Equal(other types.Expression) bool {
	if o, ok := other.(functionExpression); ok {
		if e.name == o.name && len(e.args) == len(o.args) {
			for i, arg := range e.args {
				if !arg.Equal(o.args[i]) {
					return false
				}
			}

			return true
		}
	}

	return false
}

func (e functionExpression) Name() string {
	return e.name
}

func (e functionExpression) Children() []types.Expression {
	return slices.Clone(e.args)
}

func (e functionExpression) Fold() types.Expression {
	args := make([]types.Expression, 0, len(e.args))
	for _, arg := range e.args {
		args = append(args, arg.Fold())
	}

	return NewFunction(e.name, args)
}

func (e functionExpression) Map(f func(types.Expression) types.Expression) types.Expression {
	args := make([]types.Expression, 0, len(e.args))
	for _, arg := range e.args {
		args = append(args, arg.Map(f))
	}

	return f(NewFunction(e.name, args))
}

func (e functionExpression) ValueFrom(ctx types.Context, row shared.Row) (any, error) {
	f, ok := ctx.GetFunction(e.name)
	if !ok {
		return nil, fmt.Errorf("unknown function %s", e.name)
	}

	args := make([]any, 0, len(e.args))
	for _, arg := range e.args {
		value, err := arg.ValueFrom(ctx, row)
		if err != nil {
			return nil, err
		}

		args = append(args, value)
	}

	return f.Invoke(ctx, args)
}
