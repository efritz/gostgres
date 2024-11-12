package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/projector"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/mutation"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type InsertBuilder struct {
	Target      TargetTable
	ColumnNames []string
	Source      TableReferenceOrExpression
	Returning   []projector.ProjectionExpression

	table impls.Table
}

func (b *InsertBuilder) Resolve(ctx impls.ResolutionContext) error {
	table, ok := ctx.Catalog.Tables.Get(b.Target.Name)
	if !ok {
		return fmt.Errorf("unknown table %q", b.Target.Name)
	}
	b.table = table

	if err := b.Source.Resolve(ctx); err != nil {
		return err
	}

	return nil
}

func (b *InsertBuilder) Build() (queries.Node, error) {
	node, err := b.Source.TableExpression()
	if err != nil {
		return nil, err
	}

	insert, err := mutation.NewInsert(node, b.table, b.ColumnNames)
	if err != nil {
		return nil, err
	}

	return wrapReturning(insert, b.table, b.Target.AliasName, b.Returning)
}
