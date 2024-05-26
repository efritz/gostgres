package constraints

import (
	"fmt"

	"github.com/efritz/gostgres/internal/catalog/table/indexes"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

type foreignKeyConstraint struct {
	name        string
	expressions []types.Expression
	refIndex    indexes.Index[indexes.BtreeIndexScanOptions]
}

var _ types.Constraint = &foreignKeyConstraint{}

func NewForeignKeyConstraint(name string, expressions []types.Expression, refIndex indexes.Index[indexes.BtreeIndexScanOptions]) *foreignKeyConstraint {
	return &foreignKeyConstraint{
		name:        name,
		expressions: expressions,
		refIndex:    refIndex,
	}
}

func (c *foreignKeyConstraint) Name() string {
	return c.name
}

func (c *foreignKeyConstraint) Check(ctx types.Context, row shared.Row) error {
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

	scanner, err := c.refIndex.Scanner(types.EmptyContext, indexes.NewBtreeSearchOptions(values))
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
