package constraints

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

type checkConstraint struct {
	name       string
	expression types.Expression
}

var _ types.Constraint = &checkConstraint{}

func NewCheckConstraint(name string, expression types.Expression) types.Constraint {
	return &checkConstraint{
		name:       name,
		expression: expression,
	}
}

func (c *checkConstraint) Name() string {
	return c.name
}

func (c *checkConstraint) Check(ctx types.Context, row shared.Row) error {
	if val, err := c.expression.ValueFrom(ctx, row); err != nil {
		return err
	} else if val != true {
		return fmt.Errorf("check constraint failed")
	}

	return nil
}
