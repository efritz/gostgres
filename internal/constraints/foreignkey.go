package constraints

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/indexes"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/shared"
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

func (c *foreignKeyConstraint) Check(row shared.Row) error {
	var values []any
	for _, expression := range c.expressions {
		// TODO - should to pass functionspace (but not necessary as these should all be named)
		val, err := expression.ValueFrom(expressions.EmptyContext, row)
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
		return err
	}

	return nil
}
