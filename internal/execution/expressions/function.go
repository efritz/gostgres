package expressions

import (
	"fmt"
	"slices"
	"strings"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

type functionExpression struct {
	name string
	args []impls.Expression
	typ  types.Type
}

var _ impls.Expression = &functionExpression{}

func NewFunction(name string, args []impls.Expression) impls.Expression {
	return &functionExpression{
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

func (e *functionExpression) Resolve(ctx impls.ExpressionResolutionContext) error {
	f, isAggregate, ok := lookupFunction(ctx, e.name)
	if !ok {
		return fmt.Errorf("unknown function %q", e.name)
	}
	if isAggregate && !ctx.AllowAggregateFunctions() {
		return fmt.Errorf("aggregate function %q not allowed in this context", e.name)
	}

	var argTypes []types.Type
	for _, arg := range e.args {
		if err := arg.Resolve(ctx); err != nil {
			return err
		}

		argTypes = append(argTypes, arg.Type())
	}

	if err := f.ValidateArgTypes(argTypes); err != nil {
		return err
	}

	e.typ = f.ReturnType()
	return nil
}

func lookupFunction(ctx impls.Cataloger, name string) (_ impls.Callable, isAggregate bool, ok bool) {
	if f, ok := ctx.Catalog().Functions.Get(name); ok {
		return f, false, true
	}

	if a, ok := ctx.Catalog().Aggregates.Get(name); ok {
		return a, true, true
	}

	return nil, false, false
}

func (e functionExpression) Type() types.Type {
	return e.typ
}

func (e functionExpression) Equal(other impls.Expression) bool {
	if o, ok := other.(*functionExpression); ok {
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
	f, ok := ctx.Catalog().Functions.Get(e.name)
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
