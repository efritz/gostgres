package constraints

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
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

func (c *checkConstraint) Check(ctx expressions.ExpressionContext, row shared.Row) error {
	if val, err := c.expression.ValueFrom(ctx, row); err != nil {
		return err
	} else if val != true {
		return fmt.Errorf("check constraint failed")
	}

	return nil
}
