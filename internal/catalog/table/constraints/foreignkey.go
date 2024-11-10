package constraints

import (
	"fmt"

	"github.com/efritz/gostgres/internal/catalog/table/indexes"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type foreignKeyConstraint struct {
	name        string
	expressions []impls.Expression
	refIndex    impls.Index[indexes.BtreeIndexScanOptions]
}

var _ impls.Constraint = &foreignKeyConstraint{}

func NewForeignKeyConstraint(name string, expressions []impls.Expression, refIndex impls.Index[indexes.BtreeIndexScanOptions]) impls.Constraint {
	return &foreignKeyConstraint{
		name:        name,
		expressions: expressions,
		refIndex:    refIndex,
	}
}

func (c *foreignKeyConstraint) Name() string {
	return c.name
}

func (c *foreignKeyConstraint) Check(ctx impls.ExecutionContext, row rows.Row) error {
	var values []any
	for _, expression := range c.expressions {
		val, err := expression.ValueFrom(ctx, row)
		if err != nil {
			return err
		}

		if val == nil {
			// TODO: What to do with multiple values where one is null?
			return nil
		}

		values = append(values, val)
	}

	scanner, err := c.refIndex.Scanner(impls.EmptyExecutionContext, indexes.NewBtreeSearchOptions(values))
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
