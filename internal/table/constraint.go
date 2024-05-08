package table

import (
	"fmt"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type Constraint interface {
	Check(row shared.Row) error
}

type expressionConstraint struct {
	expression expressions.Expression
}

func NewConstraint(expression expressions.Expression) Constraint {
	return &expressionConstraint{
		expression: expression,
	}
}

func (c *expressionConstraint) Check(row shared.Row) error {
	// TODO - need to pass functionspace
	if val, err := c.expression.ValueFrom(expressions.EmptyContext, row); err != nil {
		fmt.Printf("WtF: %#v\n", row)
		return err
	} else if val != true {
		return fmt.Errorf("constraint failed")
	}

	return nil
}
