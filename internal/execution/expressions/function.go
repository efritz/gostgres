package expressions

import (
	"fmt"
	"slices"
	"strings"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
)

type functionExpression struct {
	name string
	args []impls.Expression
}

var _ impls.Expression = functionExpression{}

func NewFunction(name string, args []impls.Expression) impls.Expression {
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

func (e functionExpression) Equal(other impls.Expression) bool {
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

func (e functionExpression) Children() []impls.Expression {
	return slices.Clone(e.args)
}

func (e functionExpression) Fold() impls.Expression {
	args := make([]impls.Expression, 0, len(e.args))
	for _, arg := range e.args {
		args = append(args, arg.Fold())
	}

	return NewFunction(e.name, args)
}

func (e functionExpression) Map(f func(impls.Expression) (impls.Expression, error)) (impls.Expression, error) {
	args := make([]impls.Expression, 0, len(e.args))
	for _, arg := range e.args {
		a, err := arg.Map(f)
		if err != nil {
			return nil, err
		}

		args = append(args, a)
	}

	return f(NewFunction(e.name, args))
}

func (e functionExpression) ValueFrom(ctx impls.ExecutionContext, row rows.Row) (any, error) {
	f, ok := ctx.Catalog.Functions.Get(e.name)
	if !ok {
		return nil, fmt.Errorf("unknown function %q", e.name)
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
