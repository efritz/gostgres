package constraints

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
)

type checkConstraint struct {
	name       string
	expression impls.Expression
}

var _ impls.Constraint = &checkConstraint{}

func NewCheckConstraint(name string, expression impls.Expression) impls.Constraint {
	return &checkConstraint{
		name:       name,
		expression: expression,
	}
}

func (c *checkConstraint) Name() string {
	return c.name
}

func (c *checkConstraint) Check(ctx impls.Context, row rows.Row) error {
	if val, err := c.expression.ValueFrom(ctx, row); err != nil {
		return err
	} else if val != true {
		return fmt.Errorf("check constraint failed")
	}

	return nil
}
