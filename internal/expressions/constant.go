package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type constantExpression struct {
	value interface{}
}

var _ Expression = &constantExpression{}

func NewConstant(value interface{}) Expression {
	return constantExpression{
		value: value,
	}
}

func (e constantExpression) Equal(other Expression) bool {
	if o, ok := other.(constantExpression); ok {
		return e.value == o.value
	}

	return false
}

func (e constantExpression) String() string {
	return fmt.Sprintf("%v", e.value)
}

func (e constantExpression) Fields() []shared.Field {
	return nil
}

func (e constantExpression) Named() (shared.Field, bool) {
	return shared.Field{}, false
}

func (e constantExpression) Fold() Expression {
	return e
}

func (e constantExpression) Conjunctions() []Expression {
	return []Expression{e}
}

func (e constantExpression) Alias(field shared.Field, expression Expression) Expression {
	return e
}

func (e constantExpression) ValueFrom(row shared.Row) (interface{}, error) {
	return e.value, nil
}
