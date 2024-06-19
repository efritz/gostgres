package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/projector"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/mutation"
)

type InsertBuilder struct {
	Target      TargetTable
	ColumnNames []string
	Source      TableReferenceOrExpression
	Returning   []projector.ProjectionExpression
}

func (b *InsertBuilder) Build(ctx BuildContext) (queries.Node, error) {
	table, ok := ctx.Tables.Get(b.Target.Name)
	if !ok {
		return nil, fmt.Errorf("unknown table %q", b.Target.Name)
	}

	node, err := b.Source.TableExpression(ctx)
	if err != nil {
		return nil, err
	}

	return mutation.NewInsert(node, table, b.Target.Name, b.Target.AliasName, b.ColumnNames, b.Returning)
}
