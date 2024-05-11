package expressions

import (
	"fmt"
	"slices"
	"strings"

	"github.com/efritz/gostgres/internal/shared"
)

type functionExpression struct {
	name string
	args []Expression
}

var _ Expression = &functionExpression{}

func NewFunction(name string, args []Expression) Expression {
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

func (e functionExpression) Equal(other Expression) bool {
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

func (e functionExpression) Children() []Expression {
	return slices.Clone(e.args)
}

func (e functionExpression) Fold() Expression {
	args := make([]Expression, 0, len(e.args))
	for _, arg := range e.args {
		args = append(args, arg.Fold())
	}

	return NewFunction(e.name, args)
}

func (e functionExpression) Map(f func(Expression) Expression) Expression {
	args := make([]Expression, 0, len(e.args))
	for _, arg := range e.args {
		args = append(args, arg.Map(f))
	}

	return f(NewFunction(e.name, args))
}

func (e functionExpression) ValueFrom(ctx ExpressionContext, row shared.Row) (any, error) {
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

	return f(ctx, args)
}
