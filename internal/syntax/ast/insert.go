package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/projector"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/mutation"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type InsertBuilder struct {
	Target      TargetTable
	ColumnNames []string
	Source      TableReferenceOrExpression
	Returning   []projector.ProjectionExpression
}

func (b *InsertBuilder) Resolve(ctx ResolveContext) ([]fields.Field, error) {
	return nil, fmt.Errorf("insert resolve unimplemented")
}

func (b *InsertBuilder) Build() (queries.Node, error) {
	// table, ok := ctx.Tables.Get(b.Target.Name)
	// if !ok {
	// 	return nil, fmt.Errorf("unknown table %q", b.Target.Name)
	// }
	var table impls.Table // TODO

	node, err := b.Source.Build()
	if err != nil {
		return nil, err
	}

	return mutation.NewInsert(node, table, b.Target.Name, b.Target.AliasName, b.ColumnNames, b.Returning)
}
