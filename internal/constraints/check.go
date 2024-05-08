package constraints

import (
	"fmt"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type checkConstraint struct {
	name       string
	expression expressions.Expression
}

var _ Constraint = &checkConstraint{}

func NewCheckConstraint(name string, expression expressions.Expression) *checkConstraint {
	return &checkConstraint{
		name:       name,
		expression: expression,
	}
}

func (c *checkConstraint) Name() string {
	return c.name
}

func (c *checkConstraint) Check(row shared.Row) error {
	// TODO - need to pass functionspace
	if val, err := c.expression.ValueFrom(expressions.EmptyContext, row); err != nil {
		fmt.Printf("WtF: %#v\n", row)
		return err
	} else if val != true {
		return fmt.Errorf("constraint failed")
	}

	return nil
}
