package expressions

import (
	"fmt"
	"strings"
	"time"

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

func (e functionExpression) Fields() []shared.Field {
	fields := []shared.Field{}
	for _, arg := range e.args {
		fields = append(fields, arg.Fields()...)
	}

	return fields
}

func (e functionExpression) Named() (shared.Field, bool) {
	return shared.Field{}, false
}

func (e functionExpression) Fold() Expression {
	args := make([]Expression, 0, len(e.args))
	for _, arg := range e.args {
		args = append(args, arg.Fold())
	}

	return NewFunction(e.name, args)
}

func (e functionExpression) Conjunctions() []Expression {
	return []Expression{e}
}

func (e functionExpression) Alias(field shared.Field, expression Expression) Expression {
	args := make([]Expression, 0, len(e.args))
	for _, arg := range e.args {
		args = append(args, arg.Alias(field, expression))
	}

	return NewFunction(e.name, args)
}

// TODO - This is temporary. Longer term, we need to make functions accessible
// in execution contexts so that we can create and replace functions by name at
// runtime vai DDL.
var functions = map[string]func([]any) (any, error){
	"now": func(_ []any) (any, error) { return time.Now(), nil },
}

func (e functionExpression) ValueFrom(row shared.Row) (any, error) {
	args := make([]any, 0, len(e.args))
	for _, arg := range e.args {
		value, err := arg.ValueFrom(row)
		if err != nil {
			return nil, err
		}

		args = append(args, value)
	}

	if f, ok := functions[e.name]; ok {
		return f(args)
	}

	return nil, fmt.Errorf("unknown function %s", e.name)
}
