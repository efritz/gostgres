package constraints

import (
	"fmt"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/indexes"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/table"
)

type foreignKeyConstraint struct {
	name        string
	expressions []expressions.Expression
	refIndex    indexes.Index[indexes.BtreeIndexScanOptions]
}

var _ Constraint = &foreignKeyConstraint{}

func NewForeignKeyConstraint(name string, expressions []expressions.Expression, refIndex indexes.Index[indexes.BtreeIndexScanOptions]) *foreignKeyConstraint {
	return &foreignKeyConstraint{
		name:        name,
		expressions: expressions,
		refIndex:    refIndex,
	}
}

func (c *foreignKeyConstraint) Name() string {
	return c.name
}

func (c *foreignKeyConstraint) Check(ctx table.ConstraintContext, row shared.Row) error {
	var values []any
	for _, expression := range c.expressions {
		val, err := expression.ValueFrom(ctx, row)
		if err != nil {
			return err
		}

		values = append(values, val)
	}

	scanner, err := c.refIndex.Scanner(queries.Context{}, indexes.NewBtreeSearchOptions(values))
	if err != nil {
		return err
	}

	if _, err := scanner.Scan(); err != nil {
		if err == scan.ErrNoRows {
			return fmt.Errorf("foreign key constraint %q failed", c.name)
		}

		return err
	}

	return nil
}